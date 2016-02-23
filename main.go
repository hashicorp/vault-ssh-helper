package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

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
	err := Run(os.Args[1:])
	if err != nil {
		// All the errors are logged using this one statement. All the methods
		// simply return appropriate error message.
		log.Printf("[ERROR]: %s", err)

		// Since this is not a PAM module, exiting with appropriate error
		// code does not make sense. Any non-zero exit value is considered
		// authentication failure.
		os.Exit(1)
	}
	os.Exit(0)
}

// Retrieves OTP from user and validates it with Vault server. Also, if -verify
// option is chosen, a echo request message is sent to Vault instead of OTP. If
// a proper echo message is responded, the verification will be successful.
func Run(args []string) error {
	for _, arg := range args {
		if arg == "-v" || arg == "-version" || arg == "--version" {
			fmt.Println(formattedVersion())
			return nil
		}
	}

	var config string
	var dev, verifyOnly bool
	flags := flag.NewFlagSet("ssh-helper", flag.ContinueOnError)
	flags.StringVar(&config, "config", "", "")
	flags.BoolVar(&verifyOnly, "verify-only", false, "")
	flags.BoolVar(&dev, "dev", false, "")

	flags.Usage = func() {
		fmt.Printf("%s\n", Help())
		os.Exit(0)
	}

	if err := flags.Parse(args); err != nil {
		return err
	}

	args = flags.Args()

	if len(config) == 0 {
		return fmt.Errorf("at least one config path must be specified with -config")
	}

	// Load the configuration for this helper
	clientConfig, err := api.LoadSSHAgentConfig(config)
	if err != nil {
		return err
	}
	log.Printf("vishal: clientConfig: %#v\n", clientConfig)

	if dev {
		log.Printf("==> WARNING: Dev mode is enabled!")
		if strings.HasPrefix(strings.ToLower(clientConfig.VaultAddr), "https://") {
			return fmt.Errorf("unsupported scheme in 'dev' mode")
		}
		clientConfig.CACert = ""
		clientConfig.CAPath = ""
	} else if clientConfig.CACert == "" && clientConfig.CAPath == "" {
		return fmt.Errorf("certification information needs to be provided using ca_cert or ca_path option")
	}

	// Get an http client to interact with Vault server based on the configuration
	client, err := clientConfig.NewClient()
	if err != nil {
		return err
	}

	// Logging SSH mount point since SSH backend mount point at Vault server
	// can vary and helper has no way of knowing it automatically. ssh-helper reads
	// the mount point from the configuration file and uses the same to talk
	// to Vault. In case of errors, this can be used for debugging.
	//
	// If mount point is not mentioned in the config file, default mount point
	// of the SSH backend will be used.
	log.Printf("[INFO] Using SSH Mount point: %s", clientConfig.SSHMountPoint)
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
	}

	// If OTP is echo request, this will be a verify request. Otherwise, this
	// will be a OTP validation request.
	return helper.VerifyOTP(&helper.SSHVerifyRequest{
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
  -verify-only                Verify the installation and communication with Vault server
  -version                    Display version.
`
	return strings.TrimSpace(helpText)
}
