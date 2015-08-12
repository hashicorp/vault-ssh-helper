Vault SSH Agent
===============

Vault SSH Agent is a counterpart to Vault's (https://github.com/hashicorp/vault)
SSH backend.

Vault authenticated users create SSH OTPs to connect to remote hosts. During SSH
connection establishment, the keyboard-interactive password prompt receives the
OTP entered by the user and provides it to agent. Agent authenticates clients by
verifying the OTP with Vault server.

For enabling Vault OTP authentication, agent needs to be installed on all the hosts.
SSHD configuration should be modified to enable keyboard-interactive authentication.
SSHD PAM configuration should be modified to redirect client authentication to agent.

Usage
-----
### Options
|Option       |Description|
|-------------|-----------|
|`verify`     |To verify that the agent is installed correctly and is able to talk to Vault successfully.
|`config-file`|The path to the configuration file. The properties of config file are mentioned below.

**[Note]: Refer the below configuration for Linux. It will differ for each platform.**

Agent Configuration
-------------------
Agent's configuration is written in [HashiCorp Configuration Language (HCL)][HCL]. By proxy, this means that Agent's configuration is JSON-compatible. For more information, please see the [HCL Specification][HCL].

### Properties 
|Property           |Description|
|-------------------|-----------|
|`vault_addr`       |Address of the Vault server.
|`ssh_mount_point`  |Mount point of SSH backend in Vault server.
|`ca_cert`          |Path of PEM encoded CA certificate file used to verify Vault server's SSL certificate.
|`ca_path`          |Path to directory of PEM encoded CA certificate files used to verify Vault serer.
|`tls_skip_verify`  |Skip TLS certificate verification. Highly not recommended.
|`allowed_cidr_list`|List of comma seperated CIDR blocks. If the IP used by user to connect to this machine is different than the address of network interface, in other words, if the address is NATed, then agent will not authenticate the IP returned by Vault server. In those cases, the IP returned by Vault will be matched with the CIDR blocks mentioned here. If it matches, the authentication succeeds. Use with caution.

**NOTE: `vault_addr` and `ssh_mount_point` are required properties.**

Sample `agent_config.hcl`:
```shell
vault_addr="http://127.0.0.1:8200"
ssh_mount_point="ssh"
ca_cert="/etc/vault.d/vault.crt"
tls_skip_verify=false
```

PAM Configuration
--------------------------------

```
/etc/pam.d/sshd
```

1) Comment out the previous authentication mechanism. "common-auth" represents
the standard linux authentication module and is used by many applications.
Do not modify "common-auth" file.

```
#@include common-auth
```

2) pam_exec.so runs external commands. The external command in this case is
vault-ssh-agent. If the agent binary terminates with exit code 0 if authentication
is successful. If not it fails with exit code 1.

```
auth requisite pam_exec.so quiet expose_authtok log=/tmp/vaultssh.log /usr/local/bin/vault-ssh-agent -config-file=/etc/vault.d/agent_config.hcl
```

|Keywor           |Description |
|-----------------|------------|
|`auth`           |PAM type that the configuration applies to.
|`requisite`      |If the external command fails, the authentication should fail.
|`pam_exec.so`    |PAM module that runs an external command. In this case, an SSH agent.
|`quiet`          |Supress the messages (error) from being displayed at the prompt.
|`expose_authtok` |Binary can access the password entered at the prompt.
|`vault-ssh-agent`|Absolute path to agent's binary.
|`log`            |Agent's log file.
|`config-file`    |Parameter to `vault-ssh-agent`, the path to config file.

3) Proper return from authentication flow. There will be a pipe open which listens
to the response of the authentication. Unfortunately, pam_exec.so is not closing
this pipe and pam_unix is.

```
auth optional pam_unix.so no_set_pass use_first_pass nodelay
```
 
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

```
ChallengeResponseAuthentication yes
UsePAM yes
PasswordAuthentication no
```

|Option          |Description |
|----------------|------------|
|`ChallengeResponseAuthentication yes`|[Required]Enable challenge response (keyboard-interactive) authentication.
|`UsePAM yes`                         |[Required]Enable PAM authentication modules.
|`PasswordAuthentication no`          |[Optional]Disable password authentication.

Developing Vault-ssh-agent
---------------------------

If you wish to work on agent itself or any of its built-in systems, you'll
first need [Go](https://www.golang.org) installed on your machine
(version 1.4+ is required).

For local dev first make sure Go is properly installed, including setting up a
[GOPATH](https://golang.org/doc/code.html#GOPATH). After setting up Go, you can
download the required build tools such as vet, gox, godep etc by bootstrapping
your environment.

```sh
$ make bootstrap
...
```

Next, clone this repository into `$GOPATH/src/github.com/hashicorp/vault-ssh-agent`.
Then type `make`. This will run the tests. If this exits with exit status 0,
then everything is working 

```sh
$ make
...
```

To compile a development version of Vault-ssh-agent, run `make dev`. This will put
the vault-ssh-agent binary in `bin` and `$GOPATH/bin` folders:

```sh
$ make dev
...
$ bin/vault-ssh-agent
...
```

If you're developing a specific package, you can run tests for just that
package by specifying the `TEST` variable. For example below, only
`agent` package tests will be run.

```sh
$ make test TEST=./agent
...
```


[HCL]: https://github.com/hashicorp/hcl "HashiCorp Configuration Language (HCL)"
