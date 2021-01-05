VERSION := "1.0.0"
BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := $(shell basename "$(PWD)")
BUILD_DIR := 'pdl'

# Go related variables.
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard *.go)

LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

pdl:
	@echo "  >  Building pdl binary..."
	go build $(LDFLAGS) -o $(GOBIN)/pdl cmd/pdl/main.go

compile:
	@echo "  >  Building binary..."
	go build $(LDFLAGS) -o $(GOBIN)/$(PROJECTNAME) cmd/$(BUILD_DIR)/main.go

clean:
	@echo "  >  Cleaning build cache"
	go clean

generate:
	@echo "  >  Generating dependency files..."
	go generate $(generate)

run:
	go run cmd/debugger/main.go

build:
	@echo "  >  Building binary..."
	# MacOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(GOBIN)/$(PROJECTNAME)-darwin-amd64 cmd/$(BUILD_DIR)/main.go
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(GOBIN)/$(PROJECTNAME)-linux-amd64 cmd/$(BUILD_DIR)/main.go
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(GOBIN)/$(PROJECTNAME)-windows-amd64.exe cmd/$(BUILD_DIR)/main.go
