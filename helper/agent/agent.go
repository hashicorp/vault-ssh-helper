package agent

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	agentapi "github.com/hashicorp/vault-ssh-agent/api"
	"github.com/hashicorp/vault/api"
)

func VerifyOTP(client *api.Client, mountPoint string) error {
	// Reading the one-time-password from the prompt. This is enabled
	// by supplying 'expose_authtok' option to pam module config.
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("Error reading OTP from prompt: %s\n", err)
	}

	otp := strings.TrimSuffix(string(bytes), string('\x00'))

	// Checking if an entry with supplied OTP exists in vault server.
	response, err := agentapi.SSHAgent(client, mountPoint).Verify(otp)
	if err != nil {
		return fmt.Errorf("OTP verification failed: %s\n", err)
	}

	// PAM_USER represents the username for which authentication is being
	// requested. If the response from vault server mentions the username
	// associated with the OTP. It has to be a match.
	if response.Username != os.Getenv("PAM_USER") {
		return fmt.Errorf("Username name mismatched. VaultEntry:%s, AgentUsername:%s\n", response.Username, os.Getenv("PAM_USER"))
	}

	// The IP address to which the OTP is associated should be one among
	// the network interface addresses of the machine in which agent is
	// running.
	if err := validateIP(response.IP); err != nil {
		return fmt.Errorf("Error validating IP: %s\n", err)
	}
	log.Printf("%s@%s Authenticated!\n", response.Username, response.IP)
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
