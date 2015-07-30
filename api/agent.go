package api

import (
	"fmt"
	"log"

	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

// SSHAgent is used to perform authentication related operations
type SSHAgent struct {
	c    *api.Client
	Path string
}

// SSHAgent is used to return the client for authentication related API calls.
func Agent(c *api.Client, path string) *SSHAgent {
	return &SSHAgent{
		c:    c,
		Path: path,
	}
}

// Verifies if the key provided by user is present in vault server. If yes,
// the response will contain the IP address and username associated with the
// key.
func (c *SSHAgent) Verify(otp string) (*SSHVerifyResp, error) {
	data := map[string]interface{}{
		"otp": otp,
	}
	verifyPath := fmt.Sprintf("/v1/%s/verify", c.Path)
	log.Printf("URL Verify: %s", verifyPath)
	r := c.c.NewRequest("PUT", verifyPath)
	if err := r.SetJSONBody(data); err != nil {
		return nil, err
	}

	resp, err := c.c.RawRequest(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	secret, err := api.ParseSecret(resp.Body)
	if err != nil {
		return nil, err
	}

	if secret.Data == nil {
		return nil, err
	}

	var verifyResp SSHVerifyResp
	err = mapstructure.Decode(secret.Data, &verifyResp)
	if err != nil {
		return nil, err
	}
	return &verifyResp, nil
}

// SSHVerifyResp is a structure representing the fields in vault server's
// response.
type SSHVerifyResp struct {
	Username string `json:"username"`
	IP       string `json:"ip"`
}
