package agent

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
)

// Echo request and response messages. This has to be in sync with the constants used
// in Vault's source code.
const (
	VerifyEchoRequest  = "verify-echo-request"
	VerifyEchoResponse = "verify-echo-response"
)

// Structure representing the agent's verification request.
type SSHVerifyRequest struct {
	// Http client to communicate with Vault
	Client *api.Client

	// Mount point of SSH backend at Vault
	MountPoint string

	// This can be either an echo message (see #VerifyEchoRequest), which if set
	// Vault will respond with echo response (see #VerifyEchoResponse). OR, it
	// should be the one-time-password entered by the user at the prompt.
	OTP string

	// Structure containing configuration parameters of SSH agent
	Config *api.SSHAgentConfig
}

// Reads the OTP from the prompt and sends the OTP to vault server. Server searches
// for an entry corresponding to the OTP. If there exists one, it responds with the
// IP address and username associated with it. The username returned should match the
// username for which authentication is requested (environment variable PAM_USER holds
// this value).
//
// IP address returned by vault should match the addresses of network interfaces or
// it should belong to the list of allowed CIDR blocks in the config file.
func VerifyOTP(req *SSHVerifyRequest) error {
	// Checking if an entry with supplied OTP exists in vault server.
	resp, err := req.Client.SSHAgentWithMountPoint(req.MountPoint).Verify(req.OTP)
	if err != nil {
		return err
	}

	// If OTP was an echo request, check the response for echo response and return
	if req.OTP == VerifyEchoRequest {
		if resp.Message == VerifyEchoResponse {
			log.Printf("[INFO] Agent verification successful")
			return nil
		} else {
			return fmt.Errorf("Invalid echo response")
		}
	}

	// PAM_USER represents the username for which authentication is being
	// requested. If the response from vault server mentions the username
	// associated with the OTP. It has to be a match.
	if resp.Username != os.Getenv("PAM_USER") {
		return fmt.Errorf("Username name mismatch")
	}

	// The IP address to which the OTP is associated should be one among
	// the network interface addresses of the machine in which agent is
	// running.
	if err := validateIP(resp.IP, req.Config.AllowedCidrList); err != nil {
		return err
	}

	log.Printf("[INFO] %s@%s Authenticated!", resp.Username, resp.IP)
	return nil
}

// Finds out if given IP address belongs to the IP addresses associated with
// the network interfaces of the machine in which agent is running.
//
// If none of the interface addresses match the given IP, then it is search in
// the comma seperated list of CIDR blocks passed in as second parameter. This
// list is supplied as part of agent's configuration.
func validateIP(ipStr string, cidrList string) error {
	ip := net.ParseIP(ipStr)

	// Scanning network interfaces to find an address match
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
			valid, err := validateCIDR(ip, addr.String())
			if err != nil {
				return err
			}
			if valid {
				return nil
			}
		}
	}

	// None of the network interface addresses matched the given IP.
	// Now, try to find a match with the given CIDR blocks.
	cidrs := strings.Split(cidrList, ",")
	for _, cidr := range cidrs {
		valid, err := validateCIDR(ip, cidr)
		if err != nil {
			return err
		}
		if valid {
			return nil
		}
	}

	return fmt.Errorf("Invalid IP")
}

// Checks if the given CIDR block encompasses the given IP address.
func validateCIDR(ip net.IP, cidr string) (bool, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, err
	}
	if ipnet.Contains(ip) {
		return true, nil
	}
	return false, nil
}
