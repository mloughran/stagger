default: stagger

SHELL := /bin/bash
PATH := .:$(PATH)
GOPATH := $(PWD)/Godeps/_workspace
export GOPATH

stagger: *.go */*.go
	go fmt ./...
	go build -ldflags "-X main.build \"SHA: $(shell git rev-parse HEAD) (Built $(shell date -u) with $(shell go version))\""

clean:
	rm -f stagger

.PHONY: clean default
