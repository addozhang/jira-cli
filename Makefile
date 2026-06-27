.PHONY: build test lint fmt release-snapshot help

build:
	go build -o ./bin/jr ./cmd/jr

test:
	go test -race ./...

lint:
	golangci-lint run

fmt:
	gofmt -w ./cmd ./internal

release-snapshot:
	goreleaser release --snapshot --clean

help:
	@awk -F: '/^[a-zA-Z0-9_-]+:/ {print $$1}' Makefile
