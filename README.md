# Distributed ARP Watcher

Distributed Intrusion Detection System (IDS) specifically designed to monitor, aggregate, and analyze ARP traffic across multiple network segments in real-time.

It utilizes a C Agent with gRPC C++ wrapper to sniff raw packets via `libpcap`, streaming events over an mTLS-encrypted gRPC tunnel to a centralized Go Server. The server leverages PostgreSQL for state tracking and features an Analyzer engine to instantly detect and mitigate network attacks and notifies configured notification channels.

## Core Features
*   **Encrypted Transport:** Mutual TLS (mTLS) ensures all agents are cryptographically verified and all captured network data is encrypted in transit.
*   **Real-Time Analytics Pipeline:**
    *   **ARP Poisoning / Hijacking Detection:** Detects if an attacker attempts to steal a trusted IP by advertising a malicious MAC address.
    *   **MAC Spoofing / Flapping Detection:** Identifies physical anomalies, such as a single MAC address appearing on multiple distinct network segments simultaneously.
    *   **ARP Flood Detection:** Identifies DoS/Reconnaissance attempts by rate-limiting and tracking high-volume burst traffic from specific endpoints.
*   **Dynamic Notification Routing:** Configure Webhook or Email alert destinations dynamically via a REST API.
*   **Systemd Ready:** The Agent is built to run as a robust Linux daemon with structured native `journald` logging and auto-reconnect capabilities.

---

## 1. Prerequisites & Installation

### A. System Dependencies (Fedora / RHEL)
You will need the C++ gRPC libraries, protobuf tools, and `libpcap` to compile the agent.

```bash
sudo dnf install protobuf-devel grpc-devel grpc-plugins libpcap-devel gcc-c++ make pkgconf-pkg-config
```

### B. Go Environment
Ensure Go is installed, and fetch the necessary Protobuf/gRPC plugins:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### C. Database Setup
*(The server handles its own schema migrations automatically on boot using `golang-migrate`).*

The Go server requires a PostgreSQL instance.

You can use Docker with `docker-compose.yml` and a default configuration.
```bash
cd server && docker compose up -d
```

---

## 2. Configuration & Security Setup

### Generate mTLS Certificates
Before building, you must generate the cryptographic trust chain. There is available OpenSSL script for this.

```bash
# Generate the Root CA, Server Cert, and a default Client Cert
make cert-gen
```
*Note: To generate a specific certificate for a remote LAN device, use: `./common/cert_gen.sh -c -n my-remote-agent`*

### Server Configuration
Edit `server/config.yaml` to set your PostgreSQL connection string and desired ports.
```yaml
server:
  port: ":50051" # gRPC Listening Port
api:
  port: ":8080"  # REST API Port
database:
  dsn: "postgres://postgres:postgres@localhost:5432/distarpwatcher?sslmode=disable"
```

### Agent Configuration
Edit `agent/agent.conf` for setting agent_id, interface, server and certificates.

---

## 3. Building and Running

Compile the entire stack (Protobufs, C++ Agent, and Go Server) from the root directory:

```bash
make all
```

### Start the Centralized Server
```bash
make -C server run
```

### Start the Agent
The agent requires `sudo` privileges to sniff the network. 
*Ensure your `agent/agent.conf` points to the correct Server IP and Certificate paths before running.*

```bash
make -C agent run
```
*(If no config path is provided, it defaults to `/etc/distarpwatcher/agent.conf`, falling back to a local `./agent.conf`)*

---

## 4. API & Database Overview

### REST API Endpoints
The Go server exposes a REST API (default `http://localhost:8080/api/v1`) to manage system configurations dynamically without restarting.

**Notification Channels:**
*   `GET /notification_channels` - List all active and inactive notification destinations.
*   `POST /notification_channels` - Create a new destination (Requires JSON: `Name`, `Type` (WEBHOOK/MAIL), `Target`, and optional `MinSeverity`).
*   `PUT /notification_channels/{id}` - Update an existing channel.
*   `DELETE /notification_channels/{id}` - Remove a channel.

**ARP Events:**
*   `GET /arp_events` - (Placeholder) Retrieve historical ARP events.

### Database Schema
The system uses PostgreSQL for both historical auditing and real-time state tracking.
*   **`arp_events`**: The immutable audit log. Stores every single ARP packet received from the gRPC stream.
*   **`ip_mac_bindings`**: The "Current State" table. Managed exclusively by the Analyzer to track the authoritative MAC address for every known IP, including its `last_seen` timestamp and `status` (TRUSTED/CONFLICT).
*   **`notification_channels`**: Stores dynamic routing configurations for the Notifier service.

---

## 5. Testing & Simulations

The project includes a comprehensive, isolated Python testing suite using `uv` to safely simulate network attacks.

### Setup Test Environment
Initialize the `uv` virtual environment and install dependencies (`scapy`):
```bash
make test-setup
```

### A. Mock Webhook Receiver
Start a local HTTP server to catch and print alerts generated by the Go Notifier:
```bash
make test-alert-server
```
*(To test: remember to use the REST API `POST /api/v1/notification_channels` to register `http://localhost:9000` as a webhook destination)*

### B. Simulate an ARP Flood
Blast 100 packets in a fraction of a second to test the `FrequencyDetectorStep` rate limiter.
```bash
make test-flood IFACE=<your_interface>
```

### C. Simulate ARP Poisoning
Send a targeted forged ARP packet to spoof an IP address, testing the `MACChangeDetectorStep`.
```bash
make test-poison IFACE=<your_interface> TARGET=<victim_ip> SPOOF=<gateway_ip>
```
*(Note: Modern Linux kernels aggressively drop unsolicited ARP packets. To ensure a successful cache poisoning test, there is possibility to attact light prepared docker container - use `make test-poison-container`).*