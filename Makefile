default: stagger

SHELL := /bin/bash
PATH := .:$(PATH)

stagger: *.go */*.go
	go fmt ./...
	go build -ldflags "-X main.build \"SHA: $(shell git rev-parse HEAD) (Built $(shell date -u) with $(shell go version))\""

get-deps:
	go get -v ./...

clean:
	rm -f stagger

.PHONY: clean default
