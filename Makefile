default: build

SHELL := /bin/bash
PATH := .:$(PATH)

build:
		go fmt
		go get github.com/jteeuwen/go-bindata/...
		go-bindata sparkline
		go build -ldflags "-X main.build \"SHA: $(shell git rev-parse HEAD) (Built $(shell date) with $(shell go version))\""
