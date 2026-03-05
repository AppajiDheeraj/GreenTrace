GO ?= go
BINARY ?= carbonqt
BUILD_DIR ?= bin
EXE :=

ifeq ($(OS),Windows_NT)
EXE := .exe
endif

.PHONY: build run test vet fmt clean

build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY)$(EXE)

run:
	$(GO) run .

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .

clean:
	rm -rf $(BUILD_DIR)
