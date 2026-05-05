# Binary name
BINARY_NAME=strmhelper
# Main package path
MAIN_PATH=./cmd/main.go
# Build output directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

.PHONY: all build run test clean fmt vet help dev

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
		air; \
	fi

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

vet:
	@echo "Vetting code..."
	$(GOVET) ./...

help:
	@echo "Available targets:"
	@echo "  build   - Build the binary"
	@echo "  run     - Build and run the binary"
	@echo "  dev     - Run with hot reload (requires air)"
	@echo "  test    - Run tests"
	@echo "  clean   - Remove build artifacts"
	@echo "  fmt     - Format Go code"
	@echo "  vet     - Run go vet"
	@echo "  help    - Show this help message"
