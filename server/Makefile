.PHONY: all clean

all: build/server

build/server:
	mkdir -p build
	go build -o build/server cmd/main.go

clean:
	rm -rf build
