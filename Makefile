.PHONY: build build-dev clean package-windows test generate-firebase-config

# Binary name
BINARY_NAME=racemate
BUILD_DIR=build

# Set OS and architecture
GOOS=windows
GOARCH=amd64

# Include .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Generate Firebase config file from environment variables
generate-firebase-config:
	@echo "Generating Firebase configuration..."
	@mkdir -p pkg/webserver/embedded
	@cat pkg/webserver/firebase_config.json.tmpl | \
		sed "s/\$${FIREBASE_API_KEY}/$(FIREBASE_API_KEY)/g" | \
		sed "s/\$${FIREBASE_AUTH_DOMAIN}/$(FIREBASE_AUTH_DOMAIN)/g" | \
		sed "s/\$${FIREBASE_PROJECT_ID}/$(FIREBASE_PROJECT_ID)/g" | \
		sed "s/\$${FIREBASE_STORAGE_BUCKET}/$(FIREBASE_STORAGE_BUCKET)/g" | \
		sed "s/\$${FIREBASE_MESSAGING_SENDER_ID}/$(FIREBASE_MESSAGING_SENDER_ID)/g" | \
		sed "s/\$${FIREBASE_APP_ID}/$(FIREBASE_APP_ID)/g" | \
		sed "s/\$${FIREBASE_MEASUREMENT_ID}/$(FIREBASE_MEASUREMENT_ID)/g" \
		> pkg/webserver/embedded/firebase_config.json

# Default build (production)
build: generate-firebase-config
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -o $(BUILD_DIR)/$(BINARY_NAME).exe

# Development build using values from .env file
build-dev: generate-firebase-config
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -o $(BUILD_DIR)/$(BINARY_NAME).exe

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Package for Windows using Fyne
package-windows: generate-firebase-config
	@echo "Packaging $(BINARY_NAME) for Windows..."
	fyne package -os windows -executable $(BUILD_DIR)/$(BINARY_NAME).exe -icon Icon.ico
	@echo "Windows package created"

# Run the application
run: build
	@echo "Starting $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME).exe

# Run the development version
run-dev: build-dev
	@echo "Starting $(BINARY_NAME) in development mode..."
	@./$(BUILD_DIR)/$(BINARY_NAME).exe

# Run all tests
test: generate-firebase-config
	@echo "Running all tests..."
	go test -v ./...
	