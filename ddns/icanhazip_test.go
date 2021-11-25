package ddns

import (
	"context"
	"testing"
	"time"
)

func TestGetSelfIpv4Addr(t *testing.T) {
	ctx := context.Background()
	timed, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ip, err := getSelfIpv4Addr(timed)
	if err != nil {
		t.Fatal(err)
	}
	if len(ip) == 0 {
		t.Fatalf("response is empty")
	}
	println(ip)
}

func TestGetSelfIpv6Addr(t *testing.T) {
	ctx := context.Background()
	timed, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ip, err := getSelfIpv6Addr(timed)
	if err != nil {
		t.Fatal(err)
	}
	if len(ip) == 0 {
		t.Fatalf("response is empty")
	}
	println(ip)
}
