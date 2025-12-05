// Copyright IBM Corp. 2015, 2025
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/api"
)

func TestVSH_EchoRequestAsOTP(t *testing.T) {
	// Check that verify echo request being used as OTP always fails
	testRun(t, api.VerifyEchoRequest, false, "uuid string is wrong length")

	// Check that a random non-UUID string is caught
	testRun(t, "non-uuid-input", false, "uuid string is wrong length")

	// Passing in a valid UUID should not result in this very same error
	uuid, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	testRun(t, uuid, false, "")
}

func testRun(t *testing.T, otp string, expectSuccess bool, errStr string) {
	args := []string{"-config=test-fixtures/config.hcl", "-dev"}

	tempFile, err := ioutil.TempFile("", "test")
	if err != nil || tempFile == nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	n, err := tempFile.WriteString(otp)
	if err != nil {
		t.Fatal(err)
	}

	if n != len(otp) {
		t.Fatalf("failed to write otp to temp file")
	}

	// Replace stdin for the Run method
	os.Stdin = tempFile

	// Reset the offset to the beginning
	tempFile.Seek(0, 0)

	err = Run(hclog.Default(), args)
	switch {
	case expectSuccess:
		if err != nil {
			t.Fatal(err)
		}
	default:
		if err == nil {
			t.Fatalf("expected an error")
		}
		if errStr != "" && err.Error() != errStr {
			t.Fatalf("expected a different error: got %v expected %v", err, errStr)
		}
	}
}
