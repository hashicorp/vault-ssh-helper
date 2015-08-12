package agent

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/vault/api"
)

// Returns a new client for the given configuration. If the configuration
// supplies Vault SSL certificates, then the client will have tls configured
// in its transport.
func NewClient(config *VaultConfig) (*api.Client, error) {
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
			return nil, err
		}
		clientConfig.HttpClient = config.TLSClient(certPool)
	}

	// Creating the client object for the given configuration
	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Loads the certificate from given path and creates a certificate pool from it.
func loadCACert(path string) (*x509.CertPool, error) {
	certs, err := loadCertFromPEM(path)
	if err != nil {
		return nil, err
	}

	result := x509.NewCertPool()
	for _, cert := range certs {
		result.AddCert(cert)
	}

	return result, nil
}

// Loads the certificates present in the given directory and creates a
// certificate pool from it.
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
			return err
		}

		for _, cert := range certs {
			result.AddCert(cert)
		}
		return nil
	}

	return result, filepath.Walk(path, fn)
}

// Creates a certificate from the given path
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
