CC = gcc
CXX = g++
CFLAGS = -Wall -Wextra -O2 -Isrc -Isrc/pb
CXXFLAGS = -Wall -Wextra -O2 -Isrc -Isrc/pb -std=c++17
LDFLAGS = -lpcap `pkg-config --libs grpc++ protobuf` -lpthread

SRC_C = src/main.c src/signals.c src/capture.c src/config.c
SRC_CXX = src/grpc_client.cpp $(wildcard src/pb/*.cc)

OBJ_C = $(patsubst src/%.c, build/%.o, $(SRC_C))
OBJ_CXX = $(patsubst src/%.cpp, build/%.o, $(filter-out src/pb/%.cc, $(SRC_CXX))) $(patsubst src/pb/%.cc, build/pb/%.o, $(filter src/pb/%.cc, $(SRC_CXX)))

TARGET = build/arp_watcher

all: $(TARGET)

$(TARGET): $(OBJ_C) $(OBJ_CXX) | build
	$(CXX) $(CXXFLAGS) -o $@ $^ $(LDFLAGS)

build/%.o: src/%.c | build
	@mkdir -p $(dir $@)
	$(CC) $(CFLAGS) -c -o $@ $<

build/%.o: src/%.cpp | build
	@mkdir -p $(dir $@)
	$(CXX) $(CXXFLAGS) -c -o $@ $<

build/pb/%.o: src/pb/%.cc | build
	@mkdir -p $(dir $@)
	$(CXX) $(CXXFLAGS) -c -o $@ $<

build:
	mkdir -p build

clean:
	rm -rf build

run: all
	sudo ./$(TARGET)

.PHONY: all clean run
