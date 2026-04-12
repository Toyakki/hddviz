.PHONY: all build run test clean

BINARY_NAME ?= hddviz

all: build run clean

build:
	go build -v -ldflags=-w .

run:
	go run .

test:
	go test -v ./...

clean:
	go clean
	rm -f ${BINARY_NAME}