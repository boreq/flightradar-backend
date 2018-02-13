VERSION = `git rev-parse HEAD`
DATE = `date --iso-8601=seconds`
LDFLAGS =  -X github.com/boreq/flightradar-backend/main/commands.buildCommit=$(VERSION)
LDFLAGS += -X github.com/boreq/flightradar-backend/main/commands.buildDate=$(DATE)

all: build

build:
	mkdir -p build
	go build -ldflags "$(LDFLAGS)" -o ./build/flightradar-backend ./main

build-rpi:
	mkdir -p build-rpi
	CGO_ENABLED=1 CC=arm-linux-gnueabi-gcc GOOS=linux GOARCH=arm go build -ldflags "$(LDFLAGS)" -o ./build-rpi/flightradar-backend ./main

run:
	./main/main

proto:
	protoc --proto_path="storage/bolt/messages" --go_out="storage/bolt/messages" storage/bolt/messages/messages.proto

doc:
	@echo "http://localhost:6060/pkg/github.com/boreq/flightradar-backend/"
	godoc -http=:6060

test:
	go test ./...

test-verbose:
	go test -v ./...

test-short:
	go test -short ./...

bench:
	go test -v -run=XXX -bench=. ./...

clean:
	rm -rf ./build
	rm -rf ./build-rpi

.PHONY: all build build-rpi proto run doc test test-verbose test-short bench clean
