#!/bin/bash

set -e

if [ "$EUID" -ne 0 ]; then
  echo "Run this script as root (sudo ./install_agent.sh)"
  exit 1
fi

echo "[*] Compiling the DistARPWatcher Agent..."
make -C ../agent clean
make -C ../agent

echo "[*] Creating configuration directories..."
CONFIG_DIR="/etc/distarpwatcher"
CERT_DIR="${CONFIG_DIR}/certs"
mkdir -p ${CERT_DIR}

echo "[*] Copying binary to /usr/local/bin..."
cp ../agent/build/arp_watcher /usr/local/bin/distarpwatcher-agent
chmod +x /usr/local/bin/distarpwatcher-agent

echo "[*] Setting up certificates..."
if [ ! -f "../certs/ca.pem" ] || [ ! -f "../certs/client.pem" ]; then
    echo "[!] Certificates not found in ../certs/."
    echo "    Run 'make cert-gen' to generate them."
    exit 1
fi

cp ../certs/ca.pem ${CERT_DIR}/
cp ../certs/client.pem ${CERT_DIR}/
cp ../certs/client.key ${CERT_DIR}/
chmod 600 ${CERT_DIR}/*

echo "[*] Setting up configuration file..."
cat > ${CONFIG_DIR}/agent.conf <<EOF
# Distributed ARP Watcher - Production Agent Configuration

[agent]
agent_id = $(hostname)
interface = any

[server]
# TODO: To change
server_address = localhost:50051

# mTLS Certificate Paths (Absolute paths for systemd)
ca_cert = ${CERT_DIR}/ca.pem
client_cert = ${CERT_DIR}/client.pem
client_key = ${CERT_DIR}/client.key
EOF
chmod 644 ${CONFIG_DIR}/agent.conf

echo "[*] Creating Systemd Service..."
SERVICE_FILE="/etc/systemd/system/distarpwatcher-agent.service"
cat > ${SERVICE_FILE} <<EOF
[Unit]
Description=Distributed ARP Watcher Agent
After=network.target

[Service]
Type=simple
# Must run as root to use libpcap
User=root
ExecStart=/usr/local/bin/distarpwatcher-agent /etc/distarpwatcher/agent.conf
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

echo "[*] Reloading Systemd daemon..."
systemctl daemon-reload

echo "------------------------------------------------------------------------"
echo "Installation Complete!"
echo "Config: ${CONFIG_DIR}/agent.conf"
echo ""
echo "Start service:"
echo "  sudo systemctl start distarpwatcher-agent"
echo ""
echo "Enable & automatically start:"
echo "  sudo systemctl enable distarpwatcher-agent"
echo ""
echo "Logs view (with severity colors):"
echo "  sudo journalctl -u distarpwatcher-agent -f"
echo "------------------------------------------------------------------------"
