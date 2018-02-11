VERSION = `git rev-parse HEAD`
DATE = `date --iso-8601=seconds`
LDFLAGS =  -X github.com/boreq/flightradar-backend/main/commands.buildCommit=$(VERSION)
LDFLAGS += -X github.com/boreq/flightradar-backend/main/commands.buildDate=$(DATE)

all: build

static:
	./build_static

build:
	mkdir -p build
	go build -ldflags "$(LDFLAGS)" -o ./build/flightradar-backend ./main

build-rpi:
	mkdir -p build-rpi
	CGO_ENABLED=1 CC=arm-linux-gnueabi-gcc GOOS=linux GOARCH=arm go build -ldflags "$(LDFLAGS)" -o ./build-rpi/flightradar-backend ./main

build-xgo:
	mkdir -p build-xgo
	cd build-xgo; xgo ../main
	#CGO_ENABLED=1 CC=arm-linux-gnueabi-gcc GOOS=linux GOARCH=arm go build -ldflags "$(LDFLAGS)" -o ./build-rpi/flightradar-backend ./main

run:
	./main/main

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
	rm -rf ./build-xgo

.PHONY: all build build-rpi build-xgo run doc test test-verbose test-short bench clean
