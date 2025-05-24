.PHONY: build build-dev clean package-windows test

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

# Default build (production)
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -ldflags="\
		-X main.firebaseAPIKey=$(FIREBASE_API_KEY) \
		-X main.firebaseAuthDomain=$(FIREBASE_AUTH_DOMAIN) \
		-X main.firebaseProjectID=$(FIREBASE_PROJECT_ID) \
		-X main.firebaseStorageBucket=$(FIREBASE_STORAGE_BUCKET) \
		-X main.firebaseMessagingSenderID=$(FIREBASE_MESSAGING_SENDER_ID) \
		-X main.firebaseAppID=$(FIREBASE_APP_ID) \
		-X main.firebaseMeasurementID=$(FIREBASE_MEASUREMENT_ID)" \
	-o $(BUILD_DIR)/$(BINARY_NAME).exe

# Development build using values from .env file
build-dev:
	@echo "Building $(BINARY_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -ldflags="\
		-X main.firebaseAPIKey=$(FIREBASE_API_KEY) \
		-X main.firebaseAuthDomain=$(FIREBASE_AUTH_DOMAIN) \
		-X main.firebaseProjectID=$(FIREBASE_PROJECT_ID) \
		-X main.firebaseStorageBucket=$(FIREBASE_STORAGE_BUCKET) \
		-X main.firebaseMessagingSenderID=$(FIREBASE_MESSAGING_SENDER_ID) \
		-X main.firebaseAppID=$(FIREBASE_APP_ID) \
		-X main.firebaseMeasurementID=$(FIREBASE_MEASUREMENT_ID)" \
	-o $(BUILD_DIR)/$(BINARY_NAME).exe

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Package for Windows using Fyne
package-windows: build
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
test:
	@echo "Running all tests..."
	go test -v ./...
	