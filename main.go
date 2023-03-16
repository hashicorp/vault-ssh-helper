// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault-ssh-helper/helper"
	"github.com/hashicorp/vault/api"
)

// This binary will be run as a command with the goal of client authentication.
// This is not a PAM module per se, but binary fails if verification of OTP
// fails. The PAM configuration runs this binary as an external command via
// the pam_exec.so module as a 'requisite'. Essentially, if this binary fails,
// then the authentication fails.
//
// After the installation and configuration of this helper, verify the installation
// with -verify-only option.
func main() {
	log := hclog.Default()
	err := Run(log, os.Args[1:])
	if err != nil {
		// All the errors are logged using this one statement. All the methods
		// simply return appropriate error message.
		log.Error(err.Error())

		// Since this is not a PAM module, exiting with appropriate error
		// code does not make sense. Any non-zero exit value is considered
		// authentication failure.
		os.Exit(1)
	}
	os.Exit(0)
}

// Run retrieves OTP from user and validates it with Vault server. Also, if -verify
// option is chosen, an echo request message is sent to Vault instead of OTP. If
// a proper echo message is responded, the verification will be successful.
func Run(log hclog.Logger, args []string) error {
	for _, arg := range args {
		if arg == "version" || arg == "-v" || arg == "-version" || arg == "--version" {
			fmt.Println(formattedVersion())
			return nil
		}
	}

	var config string
	var dev, verifyOnly bool
	var logLevel string
	flags := flag.NewFlagSet("ssh-helper", flag.ContinueOnError)
	flags.StringVar(&config, "config", "", "")
	flags.BoolVar(&verifyOnly, "verify-only", false, "")
	flags.BoolVar(&dev, "dev", false, "")
	flags.StringVar(&logLevel, "log-level", "info", "")

	flags.Usage = func() {
		fmt.Printf("%s\n", Help())
		os.Exit(1)
	}

	if err := flags.Parse(args); err != nil {
		return err
	}

	args = flags.Args()

	log.SetLevel(hclog.LevelFromString(logLevel))

	if len(config) == 0 {
		return fmt.Errorf("at least one config path must be specified with -config")
	}

	// Load the configuration for this helper
	clientConfig, err := api.LoadSSHHelperConfig(config)
	if err != nil {
		return err
	}

	if dev {
		log.Warn("Dev mode is enabled!")
		if strings.HasPrefix(strings.ToLower(clientConfig.VaultAddr), "https://") {
			return fmt.Errorf("unsupported scheme in 'dev' mode")
		}
		clientConfig.CACert = ""
		clientConfig.CAPath = ""
	} else if strings.HasPrefix(strings.ToLower(clientConfig.VaultAddr), "http://") {
		return fmt.Errorf("unsupported scheme. use 'dev' mode")
	}

	// Get an http client to interact with Vault server based on the configuration
	client, err := clientConfig.NewClient()
	if err != nil {
		return err
	}

	// Logging namespace and SSH mount point since SSH backend mount point at Vault server
	// can vary and helper has no way of knowing these automatically. ssh-helper reads
	// the namespace and mount point from the configuration file and uses the same to talk
	// to Vault. In case of errors, this can be used for debugging.
	//
	// If mount point is not mentioned in the config file, default mount point
	// of the SSH backend will be used.
	log.Info(fmt.Sprintf("using SSH mount point: %s", clientConfig.SSHMountPoint))
	log.Info(fmt.Sprintf("using namespace: %s", clientConfig.Namespace))
	var otp string
	if verifyOnly {
		otp = api.VerifyEchoRequest
	} else {
		// Reading the one-time-password from the prompt. This is enabled
		// by supplying 'expose_authtok' option to pam module config.
		otpBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		// Removing the terminator
		otp = strings.TrimSuffix(string(otpBytes), string('\x00'))
		_, err = uuid.ParseUUID(otp)
		if err != nil {
			return err
		}
	}

	// If OTP is echo request, this will be a verify request. Otherwise, this
	// will be a OTP validation request.
	return helper.VerifyOTP(log, &helper.SSHVerifyRequest{
		Client:     client,
		MountPoint: clientConfig.SSHMountPoint,
		OTP:        otp,
		Config:     clientConfig,
	})
}

func Help() string {
	helpText := `
Usage: vault-ssh-helper [options]

  vault-ssh-helper takes the One-Time-Password (OTP) from the client and
  validates it with Vault server. This binary should be used as an external
  command for authenticating clients during for keyboard-interactive auth
  of SSH server.

Options:

  -config=<path>              The path on disk to a configuration file.
  -dev                        Run the helper in "dev" mode, (such as testing or http)
  -log-level                  Level of logs to output. Defaults to "info". Supported values are:
                                "off", "trace", "debug", "info", "warn", and "error".
  -verify-only                Verify the installation and communication with Vault server
  -version                    Display version.
`
	return strings.TrimSpace(helpText)
}
