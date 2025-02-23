# Variables
APP_NAME = faker
GOOS = linux
GOARCH = amd64
CGO_ENABLED = 0
BUILD_DIR = bin
SRC = main.go

# Build target
all: build

# Create the build directory if it doesn't exist
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build the application for the specified OS and architecture
build: $(BUILD_DIR)
	env GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -o $(BUILD_DIR)/$(APP_NAME) $(SRC)

# Clean up build artifacts
clean:
	rm -rf $(BUILD_DIR)

.PHONY: all build clean help
