PACKAGES ?= $(shell go list ./... | grep -v /vendor/)
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
GOFMT ?= gofmt "-s"
BUILD ?= go build -o ./consul-register cmd/consul-register/main.go

fmt:
	$(GOFMT) -w $(GOFILES)

vet:
	go vet $(PACKAGES)

.PHONY: build
build:
	$(BUILD)
