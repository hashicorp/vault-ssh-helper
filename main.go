package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/vault-ssh-agent/api"
	vaultapi "github.com/hashicorp/vault/api"

	"github.com/hashicorp/hcl"
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

	// Reading the location of vault server from config file.
	var vaultConfig VaultConfig
	contents, err := ioutil.ReadFile(configFilePath)
	if !os.IsNotExist(err) {
		obj, err := hcl.Parse(string(contents))
		if err != nil {
			log.Println("Error parsing Vault address")
			return 1
		}

		if err := hcl.DecodeObject(&vaultConfig, obj); err != nil {
			log.Println("Error decoding Vault address")
			return 1
		}
	} else {
		log.Println("Error finding vault agent config file")
		return 1
	}

	client, err := client(&vaultConfig)
	if err != nil {
		log.Printf("Error creating api client: %s\n", err)
		return 1
	}

	// Reading the one-time-password from the prompt. This is enabled
	// by supplying 'expose_authtok' option to pam module config.
	bytes, _ := ioutil.ReadAll(os.Stdin)
	otp := strings.TrimSuffix(string(bytes), string('\x00'))

	// Checking if an entry with supplied OTP exists in vault server.
	response, err := api.Agent(client, vaultConfig.SSHMountPoint).Verify(otp)
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

func client(config *VaultConfig) (*vaultapi.Client, error) {
	// Creating a default client configuration for communicating with vault server.
	clientConfig := vaultapi.DefaultConfig()

	// Pointing the client to the actual address of vault server.
	clientConfig.Address = config.VaultAddr

	if config.CACert != "" || config.CAPath != "" || config.TLSSkipVerify {
		var certPool *x509.CertPool
		var err error
		if config.CACert != "" {
			certPool, err = loadCACert(config.CACert)
		} else if config.CAPath != "" {
			certPool, err = loadCAPath(config.CAPath)
		}
		if err != nil {
			return nil, fmt.Errorf("Error setting up CA path: %s", err)
		}

		tlsConfig := &tls.Config{
			InsecureSkipVerify: config.TLSSkipVerify,
			MinVersion:         tls.VersionTLS12,
			RootCAs:            certPool,
		}

		client := *http.DefaultClient
		client.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSClientConfig:     tlsConfig,
			TLSHandshakeTimeout: 10 * time.Second,
		}

		clientConfig.HttpClient = &client
	}

	// Creating the client object
	client, err := vaultapi.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func loadCACert(path string) (*x509.CertPool, error) {
	certs, err := loadCertFromPEM(path)
	if err != nil {
		return nil, fmt.Errorf("Error loading %s: %s", path, err)
	}

	result := x509.NewCertPool()
	for _, cert := range certs {
		result.AddCert(cert)
	}

	return result, nil
}

func loadCAPath(path string) (*x509.CertPool, error) {
	result := x509.NewCertPool()
	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		certs, err := loadCertFromPEM(path)
		if err != nil {
			return fmt.Errorf("Error loading %s: %s", path, err)
		}

		for _, cert := range certs {
			result.AddCert(cert)
		}
		return nil
	}

	return result, filepath.Walk(path, fn)
}

func loadCertFromPEM(path string) ([]*x509.Certificate, error) {
	pemCerts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	certs := make([]*x509.Certificate, 0, 5)
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		certs = append(certs, cert)
	}

	return certs, nil
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

type VaultConfig struct {
	VaultAddr     string `hcl:"VAULT_ADDR"`
	SSHMountPoint string `hcl:"SSH_MOUNT_POINT"`
	CACert        string `hcl:"CA_CERT"`
	CAPath        string `hcl:"CA_PATH"`
	TLSSkipVerify bool   `hcl:"TLS_SKIP_VERIFY"`
}
