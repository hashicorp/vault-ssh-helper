vault-ssh-helper[![Build Status](https://travis-ci.org/hashicorp/vault-ssh-helper.svg)](https://travis-ci.org/hashicorp/vault-ssh-helper)
===============

ssh-helper is a counterpart to Vault's (https://github.com/hashicorp/vault)
SSH backend. It enables creation of One-Time-Passwords (OTP) by Vault servers.
OTPs will be used as client authentication credentials while establishing SSH
connections with remote hosts.

All the remote hosts that belong to SSH backend's role of type OTP, will need this
ssh-helper to be installed, get its SSH configuration changed to enable keyboard-interactive
authentication and redirect its client authentication responsibility to ssh-helper.

Vault authenticated users contact Vault server and get an OTP issued for any specific
username and IP address. While establishing an SSH connection, ssh-helper reads the OTP
from the password prompt and sends it to Vault server for verification. Only if Vault
server verifies the OTP, client is authenticated and the SSH connection is allowed.

ssh-helper is not a PAM module, but it does the job of one. ssh-helper's binary is run as
an external command using `pam_exec.so` with access to password. Graceful execution
and exit of this command is a 'requisite' for authentication to be successful. If
the OTP is not validated, the binary exits with a non-zero status and hence the
authentication fails.

PAM modules are supposed to be shared object files. A decisionto write an ssh-helper
in Go was a choice between writing a PAM module in C and maintaining it for all platforms
vs using this workaround to get the job done, but with the convenience of using Go.

## Usage
-----
`vault-ssh-helper [options]`

### Options
|Option       |Description|
|-------------|-----------|
|`verify-only`|Verifies that ssh-helper is installed correctly and is able to communicate with Vault.
|`config`     |The path to the configuration file. Configuration options are mentioned below.
|`dev`        |ssh-helper communicates with Vault with TLS disabled. This is NOT recommended for production use. Use with caution.

## Download vault-ssh-helper

Download the latest version of vault-ssh-helper <a href="https://releases.hashicorp.com/vault-ssh-helper/0.1.0/">here</a>.

## Installation
-----
Install `Go` in your machine and set `GOPATH` accordingly. Clone this repository
in $GOPATH/src/github.com/hashicorp/vault-ssh-helper. Install all the dependant binaries
like godep, gox, vet etc by bootstrapping the environment.

```shell
$ make updatedeps
```

Build and install vault-ssh-helper.

```shell
$ make
$ make install
```

Follow the instructions below and modify SSH server configuration, PAM configuration
and ssh-helper configuration. Check if ssh-helper is installed and configured correctly
and also is able to communicate with Vault server properly. Before verifying the ssh-helper,
make sure that Vault server is up and running and it has mounted the SSH backend.
Also make sure that the mount path of the SSH backend is properly updated in the ssh-helper's
config file.

```shell
$ vault-ssh-helper -verify-only -config=<path-to-config-file>
Using SSH Mount point: ssh
ssh-helper verification successful!
```

If you intend to contribute to this project, compile a development version of ssh-helper,
using `make dev`. This will put the binary in `bin` and `$GOPATH/bin` folders.

```shell
$ make dev
```

If you're developing a specific package, you can run tests for just that package by
specifying the `TEST` variable. For example below, only `helper` package tests will be run.

```sh
$ make test TEST=./helper
...
```

If you intend to cross compile the binary, run `make bin`.

**[Note]: Below configuration is only applicable for Ubuntu 14.04 and the configurations differ
with with each platform and distribution.**

ssh-helper Configuration
-------------------
ssh-helper's configuration is written in [HashiCorp Configuration Language (HCL)][HCL].
By proxy, this means that ssh-helper's configuration is JSON-compatible. For more
information, please see the [HCL Specification][HCL].

### Properties
|Property           |Description|
|-------------------|-----------|
|`vault_addr`       |[Required]Address of the Vault server.
|`ssh_mount_point`  |[Required]Mount point of SSH backend in Vault server.
|`ca_cert`          |Path of PEM encoded CA certificate file used to verify Vault server's SSL certificate. `-dev` mode ignores this value.
|`ca_path`          |Path to directory of PEM encoded CA certificate files used to verify Vault server. `-dev` mode ignores this value.
|`allowed_cidr_list`|List of comma seperated CIDR blocks. If the IP used by user to connect to host is different than the addresses of host's network interfaces, in other words, if the address is NATed, then ssh-helper cannot authenticate the IP. In these cases, the IP returned by Vault will be matched with the CIDR blocks in this list. If it matches, the authentication succeeds. (Use with caution)

Sample `config.hcl`:

```hcl
vault_addr = "https://vault.example.com:8200"
ssh_mount_point = "ssh"
ca_cert = "/etc/vault-ssh-helper.d/vault.crt"
```

PAM Configuration
--------------------------------
Modify `/etc/pam.d/sshd` file.

```hcl
#@include common-auth
auth requisite pam_exec.so quiet expose_authtok log=/tmp/vaultssh.log /usr/local/bin/vault-ssh-helper -config=/etc/vault-ssh-helper.d/config.hcl
auth optional pam_unix.so not_set_pass use_first_pass nodelay
```

First, comment out the previous authentication mechanism `common-auth`, standard linux authentication module.

Next, configure the ssh-helper.

|Keyword          |Description |
|-----------------|------------|
|`auth`           |PAM type that the configuration applies to.
|`requisite`      |If the external command fails, the authentication should fail.
|`pam_exec.so`    |PAM module that runs an external command (ssh-helper).
|`quiet`          |Supress the exit status of ssh-helper from being displayed.
|`expose_authtok` |Binary can read the password from stdin.
|`vault-ssh-helper`|Absolute path to ssh-helper's binary.
|`log`            |Path to ssh-helper's log file.
|`config-file`    |Parameter to ssh-helper, the path to config file.

Lastly, return if ssh-helper authenticates the client successfully. This is a workaround
to gracefully return, by closing an open pipe.

|Option          |Description |
|----------------|------------|
|`auth`          |PAM type that the configuration applies to.
|`optional`      |If the module fails, authentication does not fail. This is a hack to properly return from the PAM flow. It closes an open pipe which ssh-helper fails to close.
|`pam_unix.so`   |Linux's standard authentication module.
|`not_set_pass`  |Module should not be allowed to set or modify passwords.
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

