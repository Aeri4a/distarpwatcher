#!/bin/bash

set -e

echo "[*] Starting E2E ARP Poisoning Test"

echo "[*] Spinning up 'victim' container..."
docker run -d --rm --privileged --name arp-victim alpine ping 8.8.8.8 > /dev/null

trap "echo '[*] Cleaning up victim container...'; docker stop arp-victim > /dev/null" EXIT

sleep 2

VICTIM_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' arp-victim)
echo "[+] Victim container is running at IP: ${VICTIM_IP}"

echo ""
echo "[*] Victim's ARP Table (BEFORE ATTACK):"
docker exec arp-victim ip neigh show
echo ""

echo "[*] Launching ARP Poison Attack from Host -> Victim..."
make test-poison \
    IFACE=docker0 \
    TARGET=${VICTIM_IP} \
    SPOOF=172.17.0.1 \
    MAC=de:ad:be:ef:ca:fe

sleep 1

echo ""
echo "[*] Victim's ARP Table (AFTER ATTACK):"
docker exec arp-victim ip neigh show
echo ""

echo "[*] ARP Poison attack completed"
