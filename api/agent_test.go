package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/hashicorp/vault-ssh-agent/client"
	"github.com/hashicorp/vault-ssh-agent/config"
	"github.com/hashicorp/vault/api"
)

func TestSSHAgent_Verify(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			cookie, err := req.Cookie("token")
			if err != nil {
				t.Fatalf("err: %s", err)
			}
			if cookie.Value != "ssh-agent" {
				t.Fatalf("bad cookie")
			}
			var httpResp interface{}
			secret := api.Secret{
				Data: map[string]interface{}{
					"username": "testuser",
					"ip":       "127.0.0.1",
				},
			}
			httpResp = secret
			enc := json.NewEncoder(w)
			enc.Encode(httpResp)
		}),
	}

	go server.Serve(ln)
	defer ln.Close()

	config, err := config.LoadConfig("../config/test-fixtures/vault.hcl")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	config.VaultAddr = fmt.Sprintf("http://%s", ln.Addr())

	client, err := client.NewClient(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	client.SetToken("ssh-agent")

	resp, err := Agent(client, config.SSHMountPoint).Verify("test-otp")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if resp.Username != "testuser" && resp.IP != "127.0.0.1" {
		t.Fatal("bad: response: %#v", resp)
	}
}
