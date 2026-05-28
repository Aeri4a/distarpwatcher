#!/bin/bash

# Configuration Defaults
DAYS_VALID=3650
KEY_SIZE=2048
OUT_DIR="./certs"

# Default Variables
GEN_SERVER=false
GEN_CLIENT=false
SERVER_IP=""
CLIENT_NAME="client"

# Print Help
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Generates mTLS certificates for DistARPWatcher."
    echo ""
    echo "Options:"
    echo "  -s              Generate Server certificate"
    echo "  -c              Generate Client certificate"
    echo "  -i <IP>         Optional: LAN IP address for the Server SAN (e.g., 192.168.1.50)"
    echo "  -n <NAME>       Optional: Name for the Client certificate (defaults to 'client')"
    echo "  -h              Show this help message"
    echo ""
    echo "If no flags are provided, it generates a CA, 1 Server cert, and 1 Client cert by default."
    exit 1
}

# Parse Flags
while getopts "sci:n:h" opt; do
    case ${opt} in
        s ) GEN_SERVER=true ;;
        c ) GEN_CLIENT=true ;;
        i ) SERVER_IP=$OPTARG ;;
        n ) CLIENT_NAME=$OPTARG ;;
        h ) usage ;;
        * ) usage ;;
    esac
done

# If no specific mode is selected, default to generating everything
if [ "$GEN_SERVER" = false ] && [ "$GEN_CLIENT" = false ]; then
    GEN_SERVER=true
    GEN_CLIENT=true
fi

mkdir -p ${OUT_DIR}

# 1. Generate Root CA (if it doesn't exist)
generate_ca() {
    if [ ! -f "${OUT_DIR}/ca.key" ] || [ ! -f "${OUT_DIR}/ca.pem" ]; then
        echo "[*] Generating new Root CA..."
        openssl genrsa -out ${OUT_DIR}/ca.key 4096
        openssl req -x509 -new -nodes -key ${OUT_DIR}/ca.key -sha256 -days ${DAYS_VALID} \
            -out ${OUT_DIR}/ca.pem \
            -subj "/C=US/ST=State/L=City/O=DistARPWatcher/CN=RootCA"
    else
        echo "[*] Found existing Root CA. Reusing..."
    fi
}

# 2. Generate Server Certificate
generate_server() {
    echo "[*] Generating Server Certificate..."
    openssl genrsa -out ${OUT_DIR}/server.key ${KEY_SIZE} 2>/dev/null

    # Create SAN Config
    cat > ${OUT_DIR}/server_ext.cnf <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOF

    # Add custom IP if provided
    if [ -n "$SERVER_IP" ]; then
        echo "IP.2 = ${SERVER_IP}" >> ${OUT_DIR}/server_ext.cnf
        echo "    Added SAN: ${SERVER_IP}"
    fi

    openssl req -new -key ${OUT_DIR}/server.key -out ${OUT_DIR}/server.csr \
        -subj "/C=US/ST=State/L=City/O=DistARPWatcher/CN=localhost" 2>/dev/null

    openssl x509 -req -in ${OUT_DIR}/server.csr -CA ${OUT_DIR}/ca.pem -CAkey ${OUT_DIR}/ca.key \
        -CAcreateserial -out ${OUT_DIR}/server.pem -days ${DAYS_VALID} -sha256 \
        -extfile ${OUT_DIR}/server_ext.cnf 2>/dev/null
    
    echo "    Created: server.pem, server.key"
}

# 3. Generate Client Certificate
generate_client() {
    echo "[*] Generating Client Certificate for: ${CLIENT_NAME}"
    
    local CLIENT_KEY="${OUT_DIR}/${CLIENT_NAME}.key"
    local CLIENT_CSR="${OUT_DIR}/${CLIENT_NAME}.csr"
    local CLIENT_PEM="${OUT_DIR}/${CLIENT_NAME}.pem"
    local CLIENT_EXT="${OUT_DIR}/${CLIENT_NAME}_ext.cnf"

    openssl genrsa -out ${CLIENT_KEY} ${KEY_SIZE} 2>/dev/null

    cat > ${CLIENT_EXT} <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = clientAuth
EOF

    openssl req -new -key ${CLIENT_KEY} -out ${CLIENT_CSR} \
        -subj "/C=US/ST=State/L=City/O=DistARPWatcher/CN=${CLIENT_NAME}" 2>/dev/null

    openssl x509 -req -in ${CLIENT_CSR} -CA ${OUT_DIR}/ca.pem -CAkey ${OUT_DIR}/ca.key \
        -CAcreateserial -out ${CLIENT_PEM} -days ${DAYS_VALID} -sha256 \
        -extfile ${CLIENT_EXT} 2>/dev/null

    echo "    Created: ${CLIENT_NAME}.pem, ${CLIENT_NAME}.key"
}

# Execute sequence
echo "----------------------------------------"
generate_ca

if [ "$GEN_SERVER" = true ]; then
    generate_server
fi

if [ "$GEN_CLIENT" = true ]; then
    generate_client
fi

# Cleanup temp files
rm -f ${OUT_DIR}/*.csr ${OUT_DIR}/*.cnf ${OUT_DIR}/*.srl

echo "----------------------------------------"
echo "Done! Certificates located in ${OUT_DIR}"
