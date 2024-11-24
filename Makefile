.PHONY: build test lint

build:
	go build -o acme-auto ./...

test:
	go test ./... -v

lint:
	golangci-lint run
