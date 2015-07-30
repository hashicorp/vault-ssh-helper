package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/vault-ssh-agent/api"
	"github.com/hashicorp/vault-ssh-agent/client"
	"github.com/hashicorp/vault-ssh-agent/config"
)

func main() {
	os.Exit(Run(os.Args[1:]))
}

// Retrieves the key from user and talks to vault server to see if it is valid.
func Run(args []string) int {
	log.Printf("Testing args: %#v\n", args)

	var configFilePath string
	flags := flag.NewFlagSet("ssh-agent", flag.ContinueOnError)
	flags.StringVar(&configFilePath, "config-file", "", "")

	flags.Usage = func() {
		log.Println("Usage: vault-ssh-agent -config-file=<config-file> [-ssh-mount-point=<mount-name>]")
	}

	if err := flags.Parse(args); err != nil {
		log.Println(fmt.Sprintf("Error parsing flags: '%s'", err))
		return 1
	}

	args = flags.Args()

	if configFilePath == "" {
		log.Println("Missing config-file param value")
		return 1
	}

	config, err := config.LoadConfig(configFilePath)
	if err != nil {
		log.Printf("Error loading config file: %s\n", err)
		return 1
	}

	client, err := client.NewClient(config)
	if err != nil {
		log.Printf("Error creating api client: %s\n", err)
		return 1
	}

	// Reading the one-time-password from the prompt. This is enabled
	// by supplying 'expose_authtok' option to pam module config.
	bytes, _ := ioutil.ReadAll(os.Stdin)
	otp := strings.TrimSuffix(string(bytes), string('\x00'))

	// Checking if an entry with supplied OTP exists in vault server.
	response, err := api.Agent(client, config.SSHMountPoint).Verify(otp)
	if err != nil {
		log.Printf("OTP verification failed")
		return 1
	}

	// PAM_USER represents the username for which authentication is being
	// requested. If the response from vault server mentions the username
	// associated with the OTP. It has to be a match.
	if response.Username != os.Getenv("PAM_USER") {
		log.Println("Username name mismatched")
		return 1
	}

	// The IP address to which the OTP is associated should be one among
	// the network interface addresses of the machine in which agent is
	// running.
	if err := validateIP(response.IP); err != nil {
		log.Printf("IP mismatch: %s\n", err)
		return 1
	}

	log.Printf("Authentication successful\n")
	return 0
}

// Finds out if given IP address belongs to the IP addresses associated with
// the network interfaces of the machine in which agent is running.
func validateIP(ipStr string) error {
	ip := net.ParseIP(ipStr)
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return fmt.Errorf("Error finding interface addresses")
		}
		for _, addr := range addrs {
			_, ipnet, err := net.ParseCIDR(addr.String())
			if err != nil {
				return err
			}
			if ipnet.Contains(ip) {
				return nil
			}
		}
	}
	return fmt.Errorf("OTP does not belong to this IP")
}
