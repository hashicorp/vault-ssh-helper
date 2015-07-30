Vault-ssh-agent
===============

vault-ssh-agent is a counterpart to Vault's (https://github.com/hashicorp/vault) SSH backend.

Vault's SSH backend needs vault-ssh-agent to be installed in remote targets to enable one-time-passwords (OTP).

Agent authenticates the client by verifying the data with Vault server.

Configuring Vault-ssh-agent
---------------------------

```sh
**[Note]: Below configuration is only for Linux. It may differ based on the target platform.**
```

* **Configuring PAM**:
`/etc/pam.d/sshd`

* **Configuring sshd**:

Developing Vault-ssh-agent
---------------------------

If you wish to work on Vault itself or any of its built-in systems, you'll
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
`vault` package tests will be run.

```sh
$ make test TEST=./vault
...
```

