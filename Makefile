default: build

build:
		go fmt
		./go-bindata sparkline
		go build -ldflags "-X main.build \"SHA: $(shell git rev-parse HEAD) (Built $(shell date) with $(shell go version))\""
