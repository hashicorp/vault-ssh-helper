package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/vault-ssh-agent/agent"
)

// This binary will be run as a command as part of pam authentication flow.
// This is not a pam module per se, but binary fails if verification of OTP
// is fails. The pam configuration runs this binary as an external command via
// the pam_exec.so module as a 'requisite'. Essentially, if this binary fails,
// then the authentication fails.
//
// After the installation of this binary, verify the installation with -verify
// -config-file options.
func main() {
	err := Run(os.Args[1:])
	if err != nil {
		// All the errors are logged using this one statement. All the methods
		// simply return appropriate error messages.
		log.Printf("[ERROR]: %s", err)

		// Since this is not a pam module, exiting with appropriate error
		// code does not make sense. Any non-zero exit value is considered
		// authentication failure.
		os.Exit(1)
	}
	os.Exit(0)
}

// Retrieves OTP from user and communicates with Vault server to see if its valid.
// Also, if -verify option is chosen, a echo request message is sent to Vault instead
// of OTP. If a proper echo message is responded, the verification is successful.
func Run(args []string) error {
	var configFilePath string
	var verify bool
	flags := flag.NewFlagSet("ssh-agent", flag.ContinueOnError)
	flags.StringVar(&configFilePath, "config-file", "", "")
	flags.BoolVar(&verify, "verify", false, "")

	flags.Usage = func() {
		log.Println("Usage: vault-ssh-agent -config-file=<config-file>")
	}

	if err := flags.Parse(args); err != nil {
		return err
	}

	args = flags.Args()

	if configFilePath == "" {
		return fmt.Errorf("missing config-file param")
	}

	config, err := agent.LoadConfig(configFilePath)
	if err != nil {
		return err
	}

	client, err := agent.NewClient(config)
	if err != nil {
		return err
	}

	// Logging SSH mount point since SSH backend mount point at Vault server
	// can vary and agent has no way of knowing it automatically. Agent reads
	// the mount point from the configuration file and uses the same to talk
	// to Vault. In case of errors, this can be used for debugging.
	log.Printf("[INFO] Using SSH Mount point: %s", config.SSHMountPoint)
	var otp string
	if verify {
		otp = agent.VerifyEchoRequest
	} else {
		// Reading the one-time-password from the prompt. This is enabled
		// by supplying 'expose_authtok' option to pam module config.
		otpBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		otp = strings.TrimSuffix(string(otpBytes), string('\x00'))
	}

	return agent.VerifyOTP(&agent.SSHVerifyRequest{
		Client:     client,
		MountPoint: config.SSHMountPoint,
		OTP:        otp,
		Config:     config,
	})
}
