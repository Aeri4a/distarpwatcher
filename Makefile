# Add Go bin directory to PATH so protoc can find protoc-gen-go and protoc-gen-go-grpc
export PATH := $(shell go env GOPATH)/bin:$(PATH)

.PHONY: all proto agent server clean

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

agent: proto
	$(MAKE) -C agent

server: proto
	$(MAKE) -C server

clean:
	rm -rf server/pb
	rm -rf agent/src/pb
	$(MAKE) -C agent clean
	$(MAKE) -C server clean
