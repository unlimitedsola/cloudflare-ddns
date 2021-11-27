package ddns

import (
	"context"
	"testing"
	"time"
)

func TestGetSelfIPv4Addr(t *testing.T) {
	ctx := context.Background()
	timed, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ip, err := getSelfIPv4Addr(timed)
	if err != nil {
		t.Fatal(err)
	}
	if len(ip) == 0 {
		t.Fatalf("response is empty")
	}
	t.Logf("Public IPV4: %s", ip)
}

func TestGetSelfIPv6Addr(t *testing.T) {
	ctx := context.Background()
	timed, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ip, err := getSelfIPv6Addr(timed)
	if err != nil {
		t.Fatal(err)
	}
	if len(ip) == 0 {
		t.Fatalf("response is empty")
	}
	t.Logf("Public IPV6: %s", ip)
}
