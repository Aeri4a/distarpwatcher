# Add Go bin directory to PATH so protoc can find protoc-gen-go and protoc-gen-go-grpc
export PATH := $(shell go env GOPATH)/bin:$(PATH)

.PHONY: all proto agent server clean cert-gen test-setup test-poison test-flood test-alert-server test-poison-container

all: proto agent server

proto:
	@echo "Generating Protobuf for Go (Server)..."
	mkdir -p server/pb
	protoc --proto_path=common --go_out=server/pb --go_opt=paths=source_relative \
	       --go-grpc_out=server/pb --go-grpc_opt=paths=source_relative \
	       common/collector.proto
	@echo "Generating Protobuf for C++ (Agent wrapper)..."
	mkdir -p agent/src/pb
	protoc --proto_path=common --cpp_out=agent/src/pb \
	       --grpc_out=agent/src/pb --plugin=protoc-gen-grpc=/usr/bin/grpc_cpp_plugin \
	       common/collector.proto

cert-gen:
	@echo "Generating mTLS Certificates..."
	chmod +x common/cert_gen.sh
	./common/cert_gen.sh

test-setup:
	@echo "Setting up Python test environment using uv..."
	cd tests && uv venv && uv pip install -e .

test-poison:
	@echo "Running ARP Poisoning Simulation (Requires sudo due to scapy raw sockets)"
	@if [ -z "$(IFACE)" ]; then echo "Error: Please provide IFACE (e.g. make test-poison IFACE=eth0)"; exit 1; fi
	@if [ -z "$(TARGET)" ]; then echo "Error: Please provide TARGET IP (e.g. make test-poison TARGET=192.168.1.100)"; exit 1; fi
	@if [ -z "$(SPOOF)" ]; then echo "Error: Please provide SPOOF IP (e.g. make test-poison SPOOF=192.168.1.1)"; exit 1; fi
	cd tests && sudo .venv/bin/arp-poison \
		-i $(IFACE) \
		-t $(TARGET) \
		-s $(SPOOF) \
		$(if $(MAC),-m $(MAC)) \
		-c 1

test-flood:
	@echo "Running ARP Flood Simulation (Requires sudo due to scapy raw sockets)"
	@if [ -z "$(IFACE)" ]; then echo "Error: Please provide IFACE (e.g. make test-flood IFACE=eth0)"; exit 1; fi
	cd tests && sudo .venv/bin/arp-flood \
		-i $(IFACE) \
		-c 100 \
		-d 0.005

test-alert-server:
	@echo "Starting Mock Webhook Server on port 9000..."
	cd tests && .venv/bin/alert-server -p 9000

test-poison-container:
	@echo "Running ARP Poisoning test with container..."
	chmod +x common/arp_poisoning_container.sh
	./common/arp_poisoning_container.sh

agent: proto
	$(MAKE) -C agent

server: proto
	$(MAKE) -C server

clean:
	rm -rf server/pb
	rm -rf agent/src/pb
	rm -rf certs
	$(MAKE) -C agent clean
	$(MAKE) -C server clean
