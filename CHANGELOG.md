## Unreleased

Features:

  * Add support for namespaces [GH-44](https://github.com/hashicorp/vault-ssh-helper/pull/44)

Improvements:

Bugs:


## 0.1.4 (November 8 2017)

SECURITY:

  * Make a safe exit when displaying usage text [GH-32]

## 0.1.3 (February 8 2017)

SECURITY:

  * Verify that OTPs conform to UUID format [7a831a5]

## 0.1.2 (August 24 2016)

IMPROVEMENTS:

  * Added `allowed_roles` option to configuration, which enforces specified
    role names to be present in the verification response received by the agent.

UPGRADE NOTES:

  * The option `allowed_roles` is a breaking change. When vault-ssh-helper
    is upgraded, it is required that the existing configuration files have
    an entry for `allowed_roles="*"` to be backwards compatible.

## 0.1.1 (February 25 2016)

SECURITY:

  * Introduced `dev` mode. If `dev` mode is not activated, `vault-ssh-helper`
    can only communicate with Vault that has TLS enabled [f7a8707]

IMPROVEMENTS:

  * Updated the documentation [GH-12]

BUG FIXES:

  * Empty check for `allowed_cidr_list` [9acaa58]

## 0.1.0 (December 2 2015)

  * Initial release
