# Address of Vault server
# (String)
vault_addr="http://127.0.0.1:8200"

# Name of the mount point where SSH backend is mounted in Vault server
# (String)
ssh_mount_point="ssh"

# Path to PEM encoded CA Cert file used to verify Vault server SSL certificate
# (String)
ca_cert=""

# Path to directory of PEM encoded CA Cert files used to verify Vault server
# SSL certificate.
# (String)
ca_path=""

# Skip TLS certificate verification. Highly not recommended.
# (Boolean) 
tls_skip_verify=false


# List of comma seperated CIDR blocks. If the IP used by user to connect to this
# machine is different than the address of network interface, in other words, if
# the address is NATed, then agent will not authenticate the IP returned by Vault
# server. In those cases, the IP returned by Vault will be matched with the CIDR
# blocks mentioned here. If it matches, the authentication succeeds. Use with caution.
# (String)
allowed_cidr_list=""
