package agent

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
)

// Reads the OTP from the prompt and sends the OTP to vault server. Server searches
// for an entry corresponding to the OTP. If there exists one, it responds with the
// IP address and username associated with it. The username returned should match the
// username for which authentication is requested (environment variable PAM_USER holds
// this value).
//
// IP address returned by vault should match the addresses of network interfaces or
// it should belong to the list of allowed CIDR blocks in the config file.
func VerifyOTP(client *api.Client, mountPoint string) error {
	// Reading the one-time-password from the prompt. This is enabled
	// by supplying 'expose_authtok' option to pam module config.
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	otp := strings.TrimSuffix(string(bytes), string('\x00'))

	// Checking if an entry with supplied OTP exists in vault server.
	response, err := SSHAgent(client, mountPoint).Verify(otp)
	if err != nil {
		return err
	}

	// PAM_USER represents the username for which authentication is being
	// requested. If the response from vault server mentions the username
	// associated with the OTP. It has to be a match.
	if response.Username != os.Getenv("PAM_USER") {
		return fmt.Errorf("[ERROR] Username name mismatch")
	}

	// The IP address to which the OTP is associated should be one among
	// the network interface addresses of the machine in which agent is
	// running.
	if err := validateIP(response.IP); err != nil {
		return err
	}
	log.Printf("[INFO] %s@%s Authenticated!", response.Username, response.IP)
	return nil
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
			return err
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
	return fmt.Errorf("[ERROR] Invalid IP")
}
