package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/vault-ssh-agent/client"
	"github.com/hashicorp/vault-ssh-agent/config"
	"github.com/hashicorp/vault-ssh-agent/helper/agent"
)

func main() {
	err := Run(os.Args[1:])
	if err != nil {
		log.Printf("err: %s", err)
		os.Exit(1)
	}
}

// Retrieves the key from user and talks to vault server to see if it is valid.
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
		return fmt.Errorf("error parsing flags: '%s'", err)
	}

	args = flags.Args()

	if configFilePath == "" {
		return fmt.Errorf("missing config-file param value")
	}

	config, err := config.LoadConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("error loading config file: %s\n", err)
	}

	log.Printf("SSH Mount point: %s\n", config.SSHMountPoint)

	client, err := client.NewClient(config)
	if err != nil {
		return fmt.Errorf("error creating api client: %s\n", err)
	}

	err = agent.VerifyOTP(client, config.SSHMountPoint)
	if err != nil {
		return fmt.Errorf("error verifying OTP: %s", err)
	}
	return nil
}
