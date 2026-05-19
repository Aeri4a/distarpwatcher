.PHONY: all clean run

all: build/server

build/server:
	mkdir -p build
	go build -o build/server cmd/main.go

clean:
	rm -rf build

run: all
	./build/server
