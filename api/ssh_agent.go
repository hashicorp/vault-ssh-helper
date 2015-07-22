package api

import "github.com/mitchellh/mapstructure"

type SSHAgent struct {
	c *Client
}

func (c *Client) SSHAgent() *SSHAgent {
	return &SSHAgent{c: c}
}

func (c *SSHAgent) Verify(otp string) (*SSHVerifyResp, error) {
	data := map[string]interface{}{
		"otp": otp,
	}
	r := c.c.NewRequest("PUT", "/v1/ssh/verify")
	if err := r.SetJSONBody(data); err != nil {
		return nil, err
	}

	resp, err := c.c.RawRequest(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	secret, err := ParseSecret(resp.Body)
	if err != nil {
		return nil, err
	}

	var verifyResp SSHVerifyResp
	err = mapstructure.Decode(secret.Data, &verifyResp)
	if err != nil {
		return nil, err
	}
	return &verifyResp, nil
}

type SSHVerifyResp struct {
	Username string `json:"username"`
	IP       string `json:"ip"`
	Valid    string `json:"valid"`
}
