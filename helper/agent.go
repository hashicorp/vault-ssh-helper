package helper

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
)

// Structure representing the ssh-helper's verification request.
type SSHVerifyRequest struct {
	// Http client to communicate with Vault
	Client *api.Client

	// Mount point of SSH backend at Vault
	MountPoint string

	// This can be either an echo request message, which if set Vault will
	// respond with echo response message. OR, it can be the one-time-password
	// entered by the user at the prompt.
	OTP string

	// Structure containing configuration parameters of ssh-helper
	Config *api.SSHHelperConfig
}

// Reads the OTP from the prompt and sends the OTP to vault server. Server searches
// for an entry corresponding to the OTP. If there exists one, it responds with the
// IP address and username associated with it. The username returned should match the
// username for which authentication is requested (environment variable PAM_USER holds
// this value).
//
// IP address returned by vault should match the addresses of network interfaces or
// it should belong to the list of allowed CIDR blocks in the config file.
//
// This method is also used to verify if the communication between ssh-helper and Vault
// server can be established with the given configuration data. If OTP in the request
// matches the echo request message, then the echo response message is expected in
// the response, which indicates successful connection establishment.
func VerifyOTP(req *SSHVerifyRequest) error {
	// Validating the OTP from Vault server. The response from server can have
	// either the response message set OR username and IP set.
	resp, err := req.Client.SSHHelperWithMountPoint(req.MountPoint).Verify(req.OTP)
	if err != nil {
		return err
	}

	// If OTP sent was an echo request, look for echo response message in the
	// response and return
	if req.OTP == api.VerifyEchoRequest {
		if resp.Message == api.VerifyEchoResponse {
			log.Printf("[INFO] vault-ssh-helper verification successful!")
			return nil
		} else {
			return fmt.Errorf("Invalid echo response")
		}
	}

	// PAM_USER represents the username for which authentication is being
	// requested. If the response from vault server mentions the username
	// associated with the OTP. It has to be a match.
	if resp.Username != os.Getenv("PAM_USER") {
		return fmt.Errorf("Username mismatch")
	}

	// The IP address to which the OTP is associated should be one among
	// the network interface addresses of the machine in which helper is
	// running. OR it should be present in allowed_cidr_list.
	if err := validateIP(resp.IP, req.Config.AllowedCidrList); err != nil {
		return err
	}

	// Reaching here means that there were no problems. Returning nil will
	// gracefully terminate the binary and client will be authenticated to
	// establish the session.
	log.Printf("[INFO] %s@%s Authenticated!", resp.Username, resp.IP)
	return nil
}

// Finds out if given IP address belongs to the IP addresses associated with
// the network interfaces of the machine in which helper is running.
//
// If none of the interface addresses match the given IP, then it is search in
// the comma seperated list of CIDR blocks. This list is supplied as part of
// helper's configuration.
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
			belongs, err := belongsToCIDR(ip, addr.String())
			if err != nil {
				return err
			}
			if belongs {
				return nil
			}
		}
	}

	if len(cidrList) == 0 {
		return fmt.Errorf("IP did not match any of the network interface address. configure 'allowed_cidr_list' option")
	}

	// None of the network interface addresses matched the given IP.
	// Now, try to find a match with the given CIDR blocks.
	cidrs := strings.Split(cidrList, ",")
	for _, cidr := range cidrs {
		belongs, err := belongsToCIDR(ip, cidr)
		if err != nil {
			return err
		}
		if belongs {
			return nil
		}
	}

	return fmt.Errorf("Invalid IP")
}

// Checks if the given CIDR block encompasses the given IP address.
func belongsToCIDR(ip net.IP, cidr string) (bool, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, err
	}
	if ipnet.Contains(ip) {
		return true, nil
	}
	return false, nil
}
