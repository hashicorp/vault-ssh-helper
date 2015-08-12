package agent

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/hcl"
)

// VaultConfig is a structure which represents the entries from the agent's
// configuration file.
type VaultConfig struct {
	VaultAddr       string `hcl:"vault_addr"`
	SSHMountPoint   string `hcl:"ssh_mount_point"`
	CACert          string `hcl:"ca_cert"`
	CAPath          string `hcl:"ca_path"`
	TLSSkipVerify   bool   `hcl:"tls_skip_verify"`
	AllowedCidrList string `hcl:"allowed_cidr_list"`
}

// Returns a HTTP client that uses TLS verification (TLS 1.2) with the given
// certificate pool.
func (c *VaultConfig) TLSClient(certPool *x509.CertPool) *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.TLSSkipVerify,
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
	return &client
}

// Loads agent's configuration from the file and populates the corresponding
// in memory structure.
func LoadConfig(path string) (*VaultConfig, error) {
	var config VaultConfig
	contents, err := ioutil.ReadFile(path)
	if !os.IsNotExist(err) {
		obj, err := hcl.Parse(string(contents))
		if err != nil {
			return nil, err
		}

		if err := hcl.DecodeObject(&config, obj); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	return &config, nil
}
