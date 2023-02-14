// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"net"
	"testing"
)

func TestValidateIP(t *testing.T) {
	t.Parallel()
	// This allows the testing of the validateIP function
	netInterfaceAddrs = func() ([]net.Addr, error) {
		var ips []net.Addr
		var err error
		//var ip net.IP
		ips = append(ips, &net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)})
		ips = append(ips, &net.IPNet{IP: net.ParseIP("10.50.100.101"), Mask: net.CIDRMask(24, 32)})
		ips = append(ips, &net.IPNet{IP: net.ParseIP("::1"), Mask: net.CIDRMask(128, 128)})

		return ips, err
	}
	var testIP string
	var testCIDR string
	var err error

	testIP = "10.50.100.101"
	testCIDR = ""
	err = validateIP(testIP, testCIDR)
	// Pass if err == nil
	if err != nil {
		t.Fatalf("Actual IP Match: expected nill, actual: %s", err)
	}

	testIP = "10.50.100.102"
	testCIDR = "10.50.100.0/24"
	err = validateIP(testIP, testCIDR)
	// Pass if err == nil
	if err != nil {
		t.Fatalf("IP in CIDR: expected nill, actual: %s", err)
	}

	testIP = "10.50.100.102"
	testCIDR = ""
	err = validateIP(testIP, testCIDR)
	// Fail if err == nil
	if err == nil {
		t.Fatalf("IP Does Not Match: expected error, actual: nil")
	}

}
func TestBelongsToCIDR(t *testing.T) {
	t.Parallel()
	testIP := net.ParseIP("10.50.100.101")
	testCIDR := "0.0.0.0/0"
	belongs, err := belongsToCIDR(testIP, testCIDR)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if !belongs {
		t.Fatalf("bad: expected:true, actual:false")
	}

	testCIDR = "192.168.0.1/16"
	belongs, err = belongsToCIDR(testIP, testCIDR)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if belongs {
		t.Fatalf("bad: expected:false, actual:true")
	}

	testCIDR = "invalid"
	_, err = belongsToCIDR(testIP, testCIDR)
	if err == nil {
		t.Fatalf("expected error")
	}
}
