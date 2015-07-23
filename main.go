package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/vault-ssh-agent/api"

	"github.com/hashicorp/hcl"
)

func main() {
	var vaultConfig VaultConfig
	contents, err := ioutil.ReadFile("/etc/vault/vault.hcl")
	if !os.IsNotExist(err) {
		obj, err := hcl.Parse(string(contents))
		if err != nil {
			log.Println("Error parsing Vault address")
			os.Exit(1)
		}

		if err := hcl.DecodeObject(&vaultConfig, obj); err != nil {
			log.Println("Error decoding Vault address")
			os.Exit(1)
		}
	}

	clientConfig := api.DefaultConfig()
	clientConfig.Address = vaultConfig.Key
	client, err := api.NewClient(clientConfig)
	if err != nil {
		log.Printf("Error creating api client: %s\n", err)
		os.Exit(1)
	}

	bytes, _ := ioutil.ReadAll(os.Stdin)
	otp := strings.TrimSuffix(string(bytes), string('\x00'))

	log.Printf("OTP: %s\n", otp)

	response, err := client.SSHAgent().Verify(otp)
	if err != nil {
		log.Printf("OTP verification failed")
		os.Exit(1)
	}

	if response.Username != os.Getenv("PAM_USER") {
		log.Println("Username name mismatched")
		os.Exit(1)
	}

	if response.Valid != "yes" {
		log.Println("OTP mismatched")
		os.Exit(1)
	}

	if err := validateIP(response.IP); err != nil {
		log.Printf("IP mismatch: %s\n", err)
		os.Exit(1)
	}

	log.Printf("Authentication successful\n")
}

func validateIP(ipStr string) error {
	ip := net.ParseIP(ipStr)
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, iface := range interfaces {
		addrs, _ := iface.Addrs()
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

type VaultConfig struct {
	Key string `hcl:"VAULT_ADDR"`
}
