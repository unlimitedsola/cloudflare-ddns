package main

import (
	"cloudflare-ddns/internal/ddns"
	"cloudflare-ddns/internal/service"
	"fmt"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
)

func usage(errmsg string) {
	log.Printf(
		"%s\n\n"+
			"usage: %s <command>\n"+
			"       where <command> is one of\n"+
			"       install, remove, debug, start or stop.\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

const svcName = "CloudflareDDNSClient"
const svcDesc = "Cloudflare DDNS Client Service"

func main() {

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		runService(svcName, false)
		return
	}

	if len(os.Args) < 2 {
		usage("no command specified")
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "debug":
		runService(svcName, true)
		return
	case "install":
		err = service.InstallService(svcName, svcDesc)
	case "remove":
		err = service.RemoveService(svcName)
	case "start":
		err = service.StartService(svcName)
	case "stop":
		err = service.ControlService(svcName, svc.Stop, svc.Stopped)
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, svcName, err)
	}
	return
}

var elog debug.Log

type ddnsService struct {
	client *ddns.Client
}

func (m *ddnsService) work() {
	hasChanged, oldIP, newIP, err := m.client.Update()
	if err != nil {
		elog.Error(1, fmt.Sprintf("failed to update record: %s", err))
		return
	}
	if hasChanged {
		elog.Info(1, fmt.Sprintf("updated existing record %s with %s", oldIP, newIP))
	}
}

func (m *ddnsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	tick := time.Tick(5 * time.Minute)
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	m.work()
loop:
	for {
		select {
		case <-tick:
			m.work()
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	client, err := ddns.New()
	if err != nil {
		elog.Error(1, fmt.Sprintf("failed to start service %s: %v", name, err))
		return
	}
	err = run(name, &ddnsService{client: client})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}
