.PHONY: build test install fmt

build:
	go build ./...

test:
	go test ./...

install:
	go install .

fmt:
	gofmt -w .
