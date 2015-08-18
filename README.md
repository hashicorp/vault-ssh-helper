Vault SSH Agent
===============

Vault SSH Agent is a counterpart to Vault's (https://github.com/hashicorp/vault)
SSH backend. It enables creation of One-Time-Passwords (OTP) by Vault servers.
OTPs will be used as client authentication credentials while establishing SSH
connections with remote hosts.

All the remote hosts that belong to SSH backend's role of type OTP, will need this
agent to be installed, get its SSH configuration changed to enable keyboard-interactive
authentication and redirect its client authentication responsibility to Vault SSH Agent.

Vault authenticated users contact Vault server and get an OTP issued for any specific
username and IP address. While establishing an SSH connection, agent reads the OTP
from the password prompt and sends it to Vault server for verification. Only if Vault
server verifies the OTP, client is authenticated and the SSH connection is allowed.

This agent is not a PAM module, but it does the job of one. Agent's binary is run as
an external command using `pam_exec.so` with access to password. Graceful execution
and exit of this command is a 'requisite' for authentication to be successful. If
the OTP is not validated, the binary exits with a non-zero status and hence the
desired effect is achieved. 

PAM modules are supposed to be shared object files and Go (currently) does not
support creation of `.so` files. It was a choice between writing a PAM module in
C and maintain it for all platforms vs using this workaround to get the job done,
but with the convenience of using Go.

## Usage
-----
`vault-ssh-agent [options]`

### Options
|Option       |Description|
|-------------|-----------|
|`verify`     |To verify that the agent is installed correctly and is able to talk to Vault successfully.
|`config-file`|The path to the configuration file. The properties of config file are mentioned below.

## Installation
-----

Install `Go` in your machine (1.4+) and set `GOPATH` accordingly. Clone this repository
in $GOPATH/src/github.com/hashicorp/vault-ssh-agent. Install all the dependant binaries
like godep, gox, vet etc by bootstrapping the environment.

```shell
$ make bootstrap
```

Build and install Vault SSH Agent.

```shell
$ make
$ make install
```

Follow the instructions below and modify SSH server, PAM configurations and configure
the agent. Check if the agent is installed and configured correctly and is able to
communicate with Vault server properly.

```shell
$ vault-ssh-agent -verify -config-file=<path-to-config-file>
Using SSH Mount point: ssh
Agent verification successful!
```

If you intend to contribute to this project, compile a development version of agent,
using `make dev`. This will put the binary in `bin` and `$GOPATH/bin` folders.

```shell
$ make dev
```

If you're developing a specific package, you can run tests for just that package by
specifying the `TEST` variable. For example below, only `agent` package tests will be run.

```sh
$ make test TEST=./agent
...
```

If you intend to cross compile the binary, run `make bin`.

**[Note]: Below configuration is only applicable for Linux and it differs with each platform.**

Agent Configuration
-------------------
Agent's configuration is written in [HashiCorp Configuration Language (HCL)][HCL].
By proxy, this means that Agent's configuration is JSON-compatible. For more
information, please see the [HCL Specification][HCL].

### Properties 
|Property           |Description|
|-------------------|-----------|
|`vault_addr`       |[Required]Address of the Vault server.
|`ssh_mount_point`  |[Required]Mount point of SSH backend in Vault server.
|`ca_cert`          |Path of PEM encoded CA certificate file used to verify Vault server's SSL certificate.
|`ca_path`          |Path to directory of PEM encoded CA certificate files used to verify Vault server.
|`allowed_cidr_list`|List of comma seperated CIDR blocks. If the IP used by user to connect to host is different than the addresses of host's network interfaces, in other words, if the address is NATed, then agent cannot authenticate the IP. In these cases, the IP returned by Vault will be matched with the CIDR blocks in this list. If it matches, the authentication succeeds. (Use with caution)
|`tls_skip_verify`  |Skip TLS certificate verification. Highly not recommended.

Sample `agent_config.hcl`:
```hcl
vault_addr="http://127.0.0.1:8200"
ssh_mount_point="ssh"
ca_cert="/etc/vault.d/vault.crt"
tls_skip_verify=false
```

PAM Configuration
--------------------------------
Modify `/etc/pam.d/sshd` file. 

```hcl
#@include common-auth
auth requisite pam_exec.so quiet expose_authtok log=/tmp/vaultssh.log /usr/local/bin/vault-ssh-agent -config-file=/etc/vault.d/agent_config.hcl
auth optional pam_unix.so no_set_pass use_first_pass nodelay
```

Firstly, comment out the previous authentication mechanism `common-auth`, standard linux authentication module.

Next, configure the agent.

|Keyword          |Description |
|-----------------|------------|
|`auth`           |PAM type that the configuration applies to.
|`requisite`      |If the external command fails, the authentication should fail.
|`pam_exec.so`    |PAM module that runs an external command. In this case, an SSH agent.
|`quiet`          |Supress the exit status of agent from being displayed.
|`expose_authtok` |Binary can read the password from stdin.
|`vault-ssh-agent`|Absolute path to agent's binary.
|`log`            |Path to agent's log file.
|`config-file`    |Parameter to `vault-ssh-agent`, the path to config file.

Lastly, return if agent authenticates the client successfully. This is a workaround
to gracefully return by closing an open pipe.

|Option          |Description |
|----------------|------------|
|`auth`          |PAM type that the configuration applies to.
|`optional`      |If the module fails, authentication does not fail. This is a hack to properly return from the PAM flow. It closes an open pipe which agent fails to close.
|`pam_unix.so`   |Linux's standard authentication module.
|`no_set_pass`   |Module should not be allowed to set or modify passwords.
|`use_first_pass`|Do not display password prompt again. Use the password from the previous module.
|`nodelay`       |Avoids the induced delay after entering a wrong password.

SSHD Configuration
--------------------------------
Modify `/etc/ssh/sshd_config` file.

```hcl
ChallengeResponseAuthentication yes
UsePAM yes
PasswordAuthentication no
```

|Option          |Description |
|----------------|------------|
|`ChallengeResponseAuthentication yes`|[Required]Enable challenge response (keyboard-interactive) authentication.
|`UsePAM yes`                         |[Required]Enable PAM authentication modules.
|`PasswordAuthentication no`          |Disable password authentication.

-----------------------

Vault SSH Backend's pull request: https://github.com/hashicorp/vault/pull/385


[HCL]: https://github.com/hashicorp/hcl "HashiCorp Configuration Language (HCL)"

