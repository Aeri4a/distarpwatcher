CC = gcc
CFLAGS = -Wall -Wextra -O2 -Isrc
LDFLAGS = -lpcap

SRC = src/main.c src/signals.c src/capture.c
OBJ = $(patsubst src/%.c, build/%.o, $(SRC))
TARGET = build/arp_watcher

all: $(TARGET)

$(TARGET): $(OBJ) | build
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

build/%.o: src/%.c | build
	$(CC) $(CFLAGS) -c -o $@ $<

build:
	mkdir -p build

clean:
	rm -rf build

run: all
	sudo ./$(TARGET) any

.PHONY: all clean run
