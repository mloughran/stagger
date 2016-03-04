default: stagger

SHELL := /bin/bash
PATH := .:$(PATH)
GOPATH := $(PWD)/Godeps/_workspace
export GOPATH

stagger: *.go */*.go
	go fmt ./...
	go build -ldflags '-X main.buildSha=$(shell git rev-parse HEAD) -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'

clean:
	rm -f stagger

.PHONY: clean default
