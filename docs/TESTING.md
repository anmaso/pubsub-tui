# Testing Guide

This document describes how to test the pubsub-tui application.

## Test Architecture

The test suite is organized into three layers:

1. **Unit Tests** - Fast tests that run without external dependencies
2. **Integration Tests** - Tests that run against the Pub/Sub emulator
3. **Manual Smoke Tests** - End-to-end verification through the TUI

```
┌─────────────────────────────────────────┐
│          Manual Smoke Tests             │
│    (Full TUI interaction with emulator) │
├─────────────────────────────────────────┤
│         Integration Tests               │
│   (Pub/Sub operations against emulator) │
├─────────────────────────────────────────┤
│            Unit Tests                   │
│   (Pure logic, no external dependencies)│
└─────────────────────────────────────────┘
```

## Running Tests

### Quick Start

```bash
# Run unit tests only (fast, no dependencies)
make test

# Run all tests including integration (requires Docker)
make test-integration-full
```

### Unit Tests

Unit tests cover:
- Resource ID validation (`validateResourceID`)
- Path extraction (`extractName`)
- Message ack/nack behavior (`ReceivedMessage`)
- Subscriber model state transitions
- Utility functions (JSON formatting, regex filtering)

```bash
# Run unit tests
make test-unit
# or
go test ./... -v
```

### Integration Tests

Integration tests require the Pub/Sub emulator running in Docker.

#### Starting the Emulator

```bash
# Start the emulator
make emulator-up

# This runs:
# docker compose -f docker-compose.test.yml up -d
```

The emulator will be available at `localhost:8085`.

#### Running Integration Tests

```bash
# With emulator already running:
make test-integration

# Or run everything (starts emulator, runs tests, stops emulator):
make test-integration-full
```

#### Stopping the Emulator

```bash
make emulator-down
```

### Environment Variables

For integration tests and running against the emulator:

| Variable | Description | Example |
|----------|-------------|---------|
| `PUBSUB_EMULATOR_HOST` | Emulator host:port | `localhost:8085` |
| `GOOGLE_CLOUD_PROJECT` | Project ID for emulator | `test-project` |

## Manual Smoke Test

This verifies the complete TUI workflow.

### Prerequisites

1. Docker installed and running
2. Terminal with at least 80x24 dimensions

### Steps

#### 1. Start the Emulator

```bash
make emulator-up
```

#### 2. Run the Application

```bash
make run-emulator
# or manually:
export PUBSUB_EMULATOR_HOST=localhost:8085
export GOOGLE_CLOUD_PROJECT=test-project
go run .
```

#### 3. Create a Topic

1. Focus is on **Topics** panel (or press `1` to switch)
2. Press `n` to create a new topic
3. Enter topic name: `smoke-test-topic`
4. Press `Enter` to confirm
5. Verify topic appears in the list
6. Press `Enter` to select the topic

#### 4. Create a Subscription

1. Press `2` to focus **Subscriptions** panel
2. Press `n` to create a new subscription
3. Enter subscription name: `smoke-test-sub`
4. Confirm the topic (should be pre-filled with `smoke-test-topic`)
5. Press `Enter` to confirm
6. Verify subscription appears in the list

#### 5. Start Subscription

1. With subscription selected, press `Enter` to start listening
2. Verify the status bar shows the active subscription
3. The subscriber panel should show "Listening..." indicator

#### 6. Publish a Message

1. Press `3` to focus **Publisher** panel
2. Select a JSON file from the list (navigate with `j`/`k`)
3. Verify the preview shows the file content
4. Press `Enter` to publish

#### 7. Receive and Acknowledge

1. Press `4` to focus **Subscriber** panel
2. Verify the message appears in the list
3. Navigate to the message and verify content in detail view
4. Press `a` to acknowledge the message
5. Verify the ack indicator changes (○ → ✓)

#### 8. Test Auto-Ack

1. Press `A` to toggle auto-ack mode
2. Publish another message from Publisher panel
3. Verify the message is automatically acknowledged

#### 9. Test Filtering

1. In Subscriber panel, press `/` to enter filter mode
2. Type a regex pattern (e.g., `test`)
3. Verify messages are filtered
4. Press `Esc` to clear filter

#### 10. Clean Up

1. Stop the subscription: Press `Esc` in Subscriber panel or Subscriptions panel
2. Delete subscription: Select it, press `d`, confirm
3. Delete topic: Focus Topics panel, select topic, press `d`, confirm
4. Exit: Press `q`

#### 11. Stop Emulator

```bash
make emulator-down
```

### Expected Behavior

| Action | Expected Result |
|--------|-----------------|
| Create topic | Topic appears in list, success message in activity log |
| Create subscription | Subscription appears, linked to topic |
| Start subscription | Status bar shows subscription, spinner in subscriber panel |
| Publish message | Success message, message ID in activity log |
| Receive message | Message appears in subscriber list with details |
| Ack message | Indicator changes, no redelivery |
| Filter messages | Only matching messages shown |
| Stop subscription | Status bar clears, subscriber shows disconnected |

### Troubleshooting

#### "Connection refused" error
- Ensure emulator is running: `docker compose -f docker-compose.test.yml ps`
- Check emulator logs: `docker compose -f docker-compose.test.yml logs`

#### No messages received
- Verify subscription was created after the topic
- Check that the subscription is actually started (spinner visible)
- Try creating a new subscription

#### Emulator not starting
- Check Docker is running
- Ensure port 8085 is not in use: `lsof -i :8085`
- Pull the image manually: `docker pull gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators`

## Writing New Tests

### Unit Tests

Place in the same package with `_test.go` suffix:

```go
func TestMyFunction(t *testing.T) {
    // Test pure logic without external dependencies
}
```

### Integration Tests

Use the `integration` build tag:

```go
//go:build integration

package pubsub

func TestIntegration_MyFeature(t *testing.T) {
    client := getTestClient(t) // Skips if emulator not available
    defer client.Close()
    
    // Test against real Pub/Sub operations
}
```

### Naming Conventions

- Unit tests: `TestXxx` or `TestXxx_Scenario`
- Integration tests: `TestIntegration_Xxx`
- Table-driven tests: Use `t.Run(name, func(t *testing.T) {...})`


