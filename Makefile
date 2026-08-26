.PHONY: build install test lint clean

build:
	go build -trimpath -ldflags "-s -w" -o bin/go-cli-template .

install:
	go install .

test:
	go test -race ./...

lint:
	golangci-lint run

clean:
	rm -rf bin
