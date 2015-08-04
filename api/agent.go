package api

import (
	"fmt"

	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

// SSHAgent is used to perform authentication related operations
type Agent struct {
	c    *api.Client
	Path string
}

// SSHAgent is used to return the client for authentication related API calls.
func SSHAgent(c *api.Client, path string) *Agent {
	return &Agent{
		c:    c,
		Path: path,
	}
}

// SSHVerifyResp is a structure representing the fields in vault server's
// response.
type SSHVerifyResp struct {
	Username string `mapstructure:"username"`
	IP       string `mapstructure:"ip"`
}

// SSHEchoResp is a structure representing the fields in vault server's
// echo response
type SSHEchoResp struct {
	Msg string `mapstructure:"echo"`
}

func (c *Agent) VaultEcho() (*SSHEchoResp, error) {
	echoPath := fmt.Sprintf("/v1/%s/echo", c.Path)
	r := c.c.NewRequest("GET", echoPath)

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
		return nil, nil
	}

	var echoResp SSHEchoResp
	err = mapstructure.Decode(secret.Data, &echoResp)
	if err != nil {
		return nil, err
	}
	return &echoResp, nil
}

// Verifies if the key provided by user is present in vault server. If yes,
// the response will contain the IP address and username associated with the
// key.
func (c *Agent) Verify(otp string) (*SSHVerifyResp, error) {
	data := map[string]interface{}{
		"otp": otp,
	}
	verifyPath := fmt.Sprintf("/v1/%s/verify", c.Path)
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
		return nil, nil
	}

	var verifyResp SSHVerifyResp
	err = mapstructure.Decode(secret.Data, &verifyResp)
	if err != nil {
		return nil, err
	}
	return &verifyResp, nil
}
