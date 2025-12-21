.PHONY: build test test-unit test-integration emulator-up emulator-down clean

# Build the binary
build:
	go build -o pubsub-tui .

# Run all tests (unit tests only by default)
test: test-unit

# Run unit tests (fast, no external dependencies)
test-unit:
	go test ./... -v -count=1

# Run integration tests (requires emulator)
test-integration:
	PUBSUB_EMULATOR_HOST=localhost:8085 GOOGLE_CLOUD_PROJECT=test-project \
		go test ./... -v -count=1 -tags=integration -timeout=60s

# Start the Pub/Sub emulator
emulator-up:
	podman compose -f docker-compose.test.yml up -d
	@echo "Waiting for emulator to be ready..."
	@sleep 5
	@echo "Pub/Sub emulator is running at localhost:8085"
	@echo "Export these environment variables to use it:"
	@echo "  export PUBSUB_EMULATOR_HOST=localhost:8085"
	@echo "  export GOOGLE_CLOUD_PROJECT=test-project"

# Stop the Pub/Sub emulator
emulator-down:
	podman compose -f docker-compose.test.yml down

# Run integration tests with emulator lifecycle
test-integration-full: emulator-up
	@echo "Running integration tests..."
	$(MAKE) test-integration || ($(MAKE) emulator-down && exit 1)
	$(MAKE) emulator-down

# Clean build artifacts
clean:
	rm -f pubsub-tui
	go clean

# Run the application (requires GCP credentials or emulator)
run:
	go run .

# Run against emulator (starts emulator if not running)
run-emulator:
	@if ! podman compose -f docker-compose.test.yml ps | grep -q "running"; then \
		$(MAKE) emulator-up; \
	fi
	PUBSUB_EMULATOR_HOST=localhost:8085 GOOGLE_CLOUD_PROJECT=test-project go run .

# Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  test               - Run unit tests (default)"
	@echo "  test-unit          - Run unit tests"
	@echo "  test-integration   - Run integration tests (requires emulator)"
	@echo "  test-integration-full - Run integration tests with emulator lifecycle"
	@echo "  emulator-up        - Start the Pub/Sub emulator"
	@echo "  emulator-down      - Stop the Pub/Sub emulator"
	@echo "  run                - Run the application"
	@echo "  run-emulator       - Run the application against emulator"
	@echo "  clean              - Clean build artifacts"


