REPO=github.com/pusher/stagger
SRCDIR=src/$(REPO)

export GOPATH=$(PWD)

default: stagger

test:
	go test

# Builds the stagger binary
stagger: $(SRCDIR) **/*.go
	go fmt
	go build -o "$@" -ldflags "-X main.build \"SHA: $(shell git rev-parse HEAD) (Built $(shell date) with $(shell go version))\""

# Adds the current repo in the GOPATH structure and gets the dependencies
$(SRCDIR):
	mkdir -p `dirname "$@"`
	ln -Ffns ../../.. "$@"
	go get $(REPO)
	ln -fs ../stagger bin/stagger

.PHONY: default test
