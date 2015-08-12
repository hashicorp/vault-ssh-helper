package agent

import "testing"

func TestSSHAgent_ConfigLoad(t *testing.T) {
	config, err := LoadConfig("./test-fixtures/vault.hcl")
	if err != nil {
		t.Fatalf("error loading config file: %s", err)
	}
	if config.VaultAddr != "http://127.0.0.1:0" {
		t.Fatalf("bad: VaultAddr: %s", config.VaultAddr)
	}
	if config.SSHMountPoint != "ssh" {
		t.Fatalf("bad: SSHMountPoint: %s", config.SSHMountPoint)
	}
}
