# Address of Vault server
VAULT_ADDR="http://127.0.0.1:8200"

# Name of the mount point where SSH backend is mounted in Vault server
SSH_MOUNT_POINT="ssh"

# Path to PEM encoded CA Cert file used to verify Vault server SSL certificate
CA_CERT=""

# Path to directory of PEM encoded CA Cert files used to verify Vault server
# SSL certificate.
CA_PATH=""

# Do not verify TLS certificate. 
# Highly not recommended.
#
# (Boolean) 
TLS_SKIP_VERIFY=false
