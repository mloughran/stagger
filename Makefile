default: build

SHELL := /bin/bash
PATH := .:$(PATH)

build: bindata.go
		go fmt ./...
		go build -ldflags "-X main.build \"SHA: $(shell git rev-parse HEAD) (Built $(shell date) with $(shell go version))\""

bindata.go: sparkline/*
	go get github.com/jteeuwen/go-bindata/...
	go-bindata sparkline
