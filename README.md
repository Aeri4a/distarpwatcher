# Distributed ARP Watcher

A distributed system for capturing and aggregating ARP events across multiple nodes. It consists of a high-performance C/C++ agent that sniffs network traffic using `libpcap` and streams ARP packets over gRPC to a Go-based centralized collector server.

## Prerequisites

### 1. System Dependencies (Fedora / RHEL)
You will need the C++ gRPC libraries, protobuf tools, and `libpcap` to compile the agent.

```bash
sudo dnf install protobuf-devel grpc-devel grpc-plugins libpcap-devel gcc-c++ make pkgconf-pkg-config
```

### 2. Go Dependencies
You will need Go installed, along with the Protobuf and gRPC plugins for Go.

```bash
# Install the Go protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
*(The root Makefile automatically includes your `$(go env GOPATH)/bin` in its PATH, so you do not need to manually configure your environment variables for the build).*

## Building

The project uses a root Makefile to orchestrate the generation of Protobuf files and the compilation of both the agent and the server.

To build the entire project from the root directory, simply run:

```bash
make all
```

This will:
1. Generate the Go gRPC code in `server/pb/`.
2. Generate the C++ gRPC code in `agent/src/pb/`.
3. Compile the C/C++ agent into `agent/build/arp_watcher`.
4. Compile the Go server into `server/build/server`.

## Running

### 1. Start the Server
Start the Go server to listen for incoming gRPC connections from agents on port `50051`.

```bash
make -C server run
```
*(Or manually: `./server/build/server`)*

### 2. Start the Agent
The agent requires `sudo` privileges to capture packets using `libpcap`. It will automatically connect to `localhost:50051` and start streaming captured ARP packets.

```bash
cd agent
sudo ./build/arp_watcher [interface]
```
- *`[interface]` is optional. If omitted, it defaults to capturing on `any` interface.*
