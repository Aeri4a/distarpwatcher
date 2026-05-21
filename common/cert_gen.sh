#!/bin/bash

# Configuration
DAYS_VALID=3650
KEY_SIZE=2048
OUT_DIR="./certs"

mkdir -p ${OUT_DIR}
echo "Generating certificates in ${OUT_DIR}..."

# 1. Generate Root CA
echo "Generating Root CA..."
openssl genrsa -out ${OUT_DIR}/ca.key 4096
openssl req -x509 -new -nodes -key ${OUT_DIR}/ca.key -sha256 -days ${DAYS_VALID} \
    -out ${OUT_DIR}/ca.pem \
    -subj "/C=US/ST=State/L=City/O=DistARPWatcher/CN=RootCA"

# 2. Generate Server Certificate
echo "Generating Server Certificate..."
openssl genrsa -out ${OUT_DIR}/server.key ${KEY_SIZE}

# Create a config for Server SAN (Subject Alternative Names)
cat > ${OUT_DIR}/server_ext.cnf <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOF

openssl req -new -key ${OUT_DIR}/server.key -out ${OUT_DIR}/server.csr \
    -subj "/C=US/ST=State/L=City/O=DistARPWatcher/CN=localhost"

openssl x509 -req -in ${OUT_DIR}/server.csr -CA ${OUT_DIR}/ca.pem -CAkey ${OUT_DIR}/ca.key \
    -CAcreateserial -out ${OUT_DIR}/server.pem -days ${DAYS_VALID} -sha256 \
    -extfile ${OUT_DIR}/server_ext.cnf

# 3. Generate Client Certificate (Agent)
echo "Generating Client Certificate..."
openssl genrsa -out ${OUT_DIR}/client.key ${KEY_SIZE}

# Create a config for Client SAN
cat > ${OUT_DIR}/client_ext.cnf <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = agent.distarp.local
EOF

openssl req -new -key ${OUT_DIR}/client.key -out ${OUT_DIR}/client.csr \
    -subj "/C=US/ST=State/L=City/O=DistARPWatcher/CN=agent-node"

openssl x509 -req -in ${OUT_DIR}/client.csr -CA ${OUT_DIR}/ca.pem -CAkey ${OUT_DIR}/ca.key \
    -CAcreateserial -out ${OUT_DIR}/client.pem -days ${DAYS_VALID} -sha256 \
    -extfile ${OUT_DIR}/client_ext.cnf

# Cleanup CSRs and temp configs
rm ${OUT_DIR}/*.csr ${OUT_DIR}/*.cnf ${OUT_DIR}/*.srl

echo "----------------------------------------------------------------"
echo "Done! Certificates generated in ${OUT_DIR}"
echo "Files created:"
echo "  - ca.pem          (The Root CA certificate - distribute to everyone)"
echo "  - server.pem/key  (Server certificate and private key)"
echo "  - client.pem/key  (Client/Agent certificate and private key)"
echo "----------------------------------------------------------------"
