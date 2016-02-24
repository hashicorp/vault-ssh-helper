package helper

import (
	"net"
	"testing"
)

func TestBelongsToCIDR(t *testing.T) {
	t.Parallel()
	testIP := net.ParseIP("10.50.100.101")
	boo, err := belongsToCIDR(testIP, "0.0.0.0/0")
	if err != nil || !boo {
		t.Error("belongsToCIDR test 1 failed")
	}
	boo, err = belongsToCIDR(testIP, "192.168.0.1/16")
	if err != nil || boo {
		t.Error("belongsToCIDR test 2 failed")
	}
	boo, err = belongsToCIDR(testIP, "failure")
	if err == nil {
		t.Error("belongsToCIDR test 3 failed")
	}
}
