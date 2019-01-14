NAME ?= watcher
PACKAGES ?= $(shell go list ./... | grep -v /vendor/)
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
GOFMT ?= gofmt "-s"
BUILD ?= go build -ldflags "-s -w" -o ./$(NAME) ./cmd/$(NAME)/main.go
PACK ?= gzip ./$(NAME)

fmt:
	$(GOFMT) -w $(GOFILES)

vet:
	go vet $(PACKAGES)

.PHONY: build
build:
	$(BUILD)

.PHONY: pack
pack:
	$(PACK)
