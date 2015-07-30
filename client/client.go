package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/vault-ssh-agent/config"
	"github.com/hashicorp/vault/api"
)

func NewClient(config *config.VaultConfig) (*api.Client, error) {
	// Creating a default client configuration for communicating with vault server.
	clientConfig := api.DefaultConfig()

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
	client, err := api.NewClient(clientConfig)
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
