# Address of Vault server
vault_addr="http://127.0.0.1:8200"

# Name of the mount point where SSH backend is mounted in Vault server
ssh_mount_point="ssh"

# Path to PEM encoded CA Cert file used to verify Vault server SSL certificate
ca_cert=""

# Path to directory of PEM encoded CA Cert files used to verify Vault server
# SSL certificate.
ca_path=""

# Skip TLS certificate verification. Highly not recommended.
#
# (Boolean) 
tls_skip_verify=false
