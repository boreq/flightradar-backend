VERSION = `git rev-parse HEAD`
DATE = `date --iso-8601=seconds`
LDFLAGS =  -X github.com/boreq/flightradar/main/commands.buildCommit=$(VERSION)
LDFLAGS += -X github.com/boreq/flightradar/main/commands.buildDate=$(DATE)

all: build

static:
	./build_static

build:
	mkdir -p build
	go build -ldflags "$(LDFLAGS)" -o ./build/flightradar ./main

run:
	./main/main

doc:
	@echo "http://localhost:6060/pkg/github.com/boreq/flightradar/"
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

.PHONY: all build run doc test test-verbose test-short bench clean
