package agent

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
)

func TestSSHAgent_NewClient(t *testing.T) {
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
			w.Write([]byte("vault-response"))
		}),
	}

	go server.Serve(ln)
	defer ln.Close()

	config, err := config.LoadConfig("../config/test-fixtures/vault.hcl")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	config.VaultAddr = fmt.Sprintf("http://%s", ln.Addr())

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	client.SetToken("ssh-agent")

	resp, err := client.RawRequest(client.NewRequest("PUT", "/"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)

	if buf.String() != "vault-response" {
		t.Fatalf("bad response")
	}
}
