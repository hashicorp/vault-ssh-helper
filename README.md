vault-ssh-agent
=========
vault-ssh-agent is a counterpart to Vault's (https://github.com/hashicorp/vault) SSH backend.

Vault's SSH backend needs vault-ssh-agent to be installed in remote targets to enable one-time-passwords (OTP).

Agent receives the OTP from the client machine and sends it to vault server. If the response from vault server matches the vault
