package main

import (
	"cloudflare-ddns/ddns"
	"context"
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
	isIntSess, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if isIntSess {
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
		err = installService(svcName, svcDesc)
	case "remove":
		err = removeService(svcName)
	case "start":
		err = startService(svcName)
	case "stop":
		err = controlService(svcName, svc.Stop, svc.Stopped)
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
	client *ddns.DDNS
}

var handler = &ddnsHandler{}

type ddnsHandler struct{}

func (h *ddnsHandler) OnZoneError(zone string, err error) {
	elog.Error(1, fmt.Sprintf("Failed to gather zone info for %s: %s", zone, err))
}

func (h *ddnsHandler) OnError(name string, err error) {
	elog.Error(1, fmt.Sprintf("Failed to update record for %s: %s", name, err))
}

func (h *ddnsHandler) OnUpdate(name string, recordType string, previous string, current string) {
	elog.Info(1, fmt.Sprintf("Updated %s record %s from %s to %s", recordType, name, previous, current))
}

func (h *ddnsHandler) OnCreate(name string, recordType string, current string) {
	elog.Info(1, fmt.Sprintf("Created %s record %s pointed to %s", recordType, name, current))
}

func (m *ddnsService) work(ctx context.Context) {
	err := m.client.Run(ctx, handler)
	if err != nil {
		elog.Error(1, fmt.Sprint(err))
	}
}

func (m *ddnsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	tick := time.Tick(5 * time.Minute)
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	ctx := context.Background()
	ctxWithCancel, cancelFunction := context.WithCancel(ctx)
	m.work(ctxWithCancel)
loop:
	for {
		select {
		case <-tick:
			m.work(ctxWithCancel)
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
	cancelFunction()
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
