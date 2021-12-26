VERSION = $(shell git describe --always --tags --dirty)

all: build

build:
	CGO_ENABLED=0 go build -ldflags "-X main.VERSION=$(VERSION)" -o bin/ .

.PHONY: build all
