//go:build integration

package pubsub

import (
	"context"
	"os"
	"testing"
	"time"
)

// skipIfNoEmulator skips the test if the emulator is not configured
func skipIfNoEmulator(t *testing.T) {
	t.Helper()
	if !IsEmulatorEnabled() {
		t.Skip("Skipping integration test: PUBSUB_EMULATOR_HOST not set")
	}
	if os.Getenv("GOOGLE_CLOUD_PROJECT") == "" {
		t.Skip("Skipping integration test: GOOGLE_CLOUD_PROJECT not set")
	}
}

func getTestClient(t *testing.T) *Client {
	t.Helper()
	skipIfNoEmulator(t)

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	client, err := NewClient(projectID)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

func TestIntegration_TopicCRUD(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	topicName := "test-topic-" + time.Now().Format("20060102150405")

	// 1. Create topic
	err := client.CreateTopic(ctx, topicName)
	if err != nil {
		t.Fatalf("CreateTopic failed: %v", err)
	}

	// 2. Verify topic exists in list
	topics, err := client.ListTopics(ctx)
	if err != nil {
		t.Fatalf("ListTopics failed: %v", err)
	}

	found := false
	for _, topic := range topics {
		if topic.Name == topicName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Created topic %q not found in ListTopics", topicName)
	}

	// 3. Verify TopicExists
	exists, err := client.TopicExists(ctx, topicName)
	if err != nil {
		t.Fatalf("TopicExists failed: %v", err)
	}
	if !exists {
		t.Errorf("TopicExists returned false for existing topic")
	}

	// 4. Delete topic
	err = client.DeleteTopic(ctx, topicName)
	if err != nil {
		t.Fatalf("DeleteTopic failed: %v", err)
	}

	// 5. Verify topic no longer exists
	exists, err = client.TopicExists(ctx, topicName)
	if err != nil {
		t.Fatalf("TopicExists after delete failed: %v", err)
	}
	if exists {
		t.Errorf("TopicExists returned true after deletion")
	}
}

func TestIntegration_TopicCreate_AlreadyExists(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	topicName := "test-topic-dup-" + time.Now().Format("20060102150405")

	// Create topic
	err := client.CreateTopic(ctx, topicName)
	if err != nil {
		t.Fatalf("CreateTopic failed: %v", err)
	}
	defer client.DeleteTopic(ctx, topicName) // Cleanup

	// Try to create same topic again
	err = client.CreateTopic(ctx, topicName)
	if err == nil {
		t.Error("CreateTopic should fail for existing topic")
	}
}

func TestIntegration_SubscriptionCRUD(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	topicName := "test-topic-sub-" + time.Now().Format("20060102150405")
	subName := "test-sub-" + time.Now().Format("20060102150405")

	// 1. Create topic first
	err := client.CreateTopic(ctx, topicName)
	if err != nil {
		t.Fatalf("CreateTopic failed: %v", err)
	}
	defer client.DeleteTopic(ctx, topicName)

	// 2. Create subscription
	err = client.CreateSubscription(ctx, subName, topicName)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	// 3. Verify subscription exists in list
	subs, err := client.ListSubscriptions(ctx)
	if err != nil {
		t.Fatalf("ListSubscriptions failed: %v", err)
	}

	found := false
	var foundSub SubscriptionInfo
	for _, sub := range subs {
		if sub.Name == subName {
			found = true
			foundSub = sub
			break
		}
	}
	if !found {
		t.Errorf("Created subscription %q not found in ListSubscriptions", subName)
	}

	// 4. Verify topic association
	if foundSub.TopicName != topicName {
		t.Errorf("Subscription TopicName = %q, want %q", foundSub.TopicName, topicName)
	}

	// 5. Verify SubscriptionExists
	exists, err := client.SubscriptionExists(ctx, subName)
	if err != nil {
		t.Fatalf("SubscriptionExists failed: %v", err)
	}
	if !exists {
		t.Errorf("SubscriptionExists returned false for existing subscription")
	}

	// 6. Delete subscription
	err = client.DeleteSubscription(ctx, subName)
	if err != nil {
		t.Fatalf("DeleteSubscription failed: %v", err)
	}

	// 7. Verify subscription no longer exists
	exists, err = client.SubscriptionExists(ctx, subName)
	if err != nil {
		t.Fatalf("SubscriptionExists after delete failed: %v", err)
	}
	if exists {
		t.Errorf("SubscriptionExists returned true after deletion")
	}
}

func TestIntegration_SubscriptionCreate_TopicNotExists(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	subName := "test-sub-no-topic-" + time.Now().Format("20060102150405")

	// Try to create subscription for non-existent topic
	err := client.CreateSubscription(ctx, subName, "non-existent-topic")
	if err == nil {
		t.Error("CreateSubscription should fail for non-existent topic")
		// Cleanup if it somehow succeeded
		client.DeleteSubscription(ctx, subName)
	}
}

func TestIntegration_PublishReceive(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	topicName := "test-topic-pubsub-" + time.Now().Format("20060102150405")
	subName := "test-sub-pubsub-" + time.Now().Format("20060102150405")

	// Setup: Create topic and subscription
	if err := client.CreateTopic(ctx, topicName); err != nil {
		t.Fatalf("CreateTopic failed: %v", err)
	}
	defer client.DeleteTopic(ctx, topicName)

	if err := client.CreateSubscription(ctx, subName, topicName); err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}
	defer client.DeleteSubscription(ctx, subName)

	// Start subscription
	sub := client.Subscribe(subName)
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	sub.Start(subCtx)

	// Give the subscription time to start
	time.Sleep(100 * time.Millisecond)

	// Publish a message
	testData := []byte(`{"test": "message", "timestamp": "now"}`)
	testAttrs := map[string]string{"key": "value"}

	result := client.Publish(ctx, topicName, testData, testAttrs)
	if result.Error != nil {
		t.Fatalf("Publish failed: %v", result.Error)
	}
	if result.MessageID == "" {
		t.Error("Publish returned empty MessageID")
	}

	// Wait for message with timeout
	select {
	case msg := <-sub.Messages():
		// Verify message content
		if string(msg.Data) != string(testData) {
			t.Errorf("Message Data = %q, want %q", string(msg.Data), string(testData))
		}
		if msg.Attributes["key"] != "value" {
			t.Errorf("Message Attributes[key] = %q, want %q", msg.Attributes["key"], "value")
		}

		// Ack the message
		msg.Ack()
		if !msg.IsAcked() {
			t.Error("Message should be acked after Ack()")
		}

	case err := <-sub.Errors():
		t.Fatalf("Subscription error: %v", err)

	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	// Stop subscription
	sub.Stop()
	if sub.IsRunning() {
		t.Error("Subscription should not be running after Stop()")
	}
}

func TestIntegration_PublishMultipleMessages(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	topicName := "test-topic-multi-" + time.Now().Format("20060102150405")
	subName := "test-sub-multi-" + time.Now().Format("20060102150405")

	// Setup
	if err := client.CreateTopic(ctx, topicName); err != nil {
		t.Fatalf("CreateTopic failed: %v", err)
	}
	defer client.DeleteTopic(ctx, topicName)

	if err := client.CreateSubscription(ctx, subName, topicName); err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}
	defer client.DeleteSubscription(ctx, subName)

	// Start subscription
	sub := client.Subscribe(subName)
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	sub.Start(subCtx)

	time.Sleep(100 * time.Millisecond)

	// Publish 3 messages
	messageCount := 3
	for i := 0; i < messageCount; i++ {
		data := []byte(`{"index": ` + string(rune('0'+i)) + `}`)
		result := client.Publish(ctx, topicName, data, nil)
		if result.Error != nil {
			t.Fatalf("Publish %d failed: %v", i, result.Error)
		}
	}

	// Receive all messages
	received := 0
	timeout := time.After(15 * time.Second)

	for received < messageCount {
		select {
		case msg := <-sub.Messages():
			received++
			msg.Ack()

		case err := <-sub.Errors():
			t.Fatalf("Subscription error: %v", err)

		case <-timeout:
			t.Fatalf("Timeout: received %d/%d messages", received, messageCount)
		}
	}

	if received != messageCount {
		t.Errorf("Received %d messages, want %d", received, messageCount)
	}

	sub.Stop()
}

func TestIntegration_SubscriptionStartStop(t *testing.T) {
	client := getTestClient(t)
	defer client.Close()

	ctx := context.Background()
	topicName := "test-topic-startstop-" + time.Now().Format("20060102150405")
	subName := "test-sub-startstop-" + time.Now().Format("20060102150405")

	// Setup
	if err := client.CreateTopic(ctx, topicName); err != nil {
		t.Fatalf("CreateTopic failed: %v", err)
	}
	defer client.DeleteTopic(ctx, topicName)

	if err := client.CreateSubscription(ctx, subName, topicName); err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}
	defer client.DeleteSubscription(ctx, subName)

	sub := client.Subscribe(subName)

	// Initially not running
	if sub.IsRunning() {
		t.Error("Subscription should not be running before Start()")
	}

	// Start
	subCtx, cancel := context.WithCancel(ctx)
	sub.Start(subCtx)

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	if !sub.IsRunning() {
		t.Error("Subscription should be running after Start()")
	}

	// Start again (should be no-op)
	sub.Start(subCtx)
	if !sub.IsRunning() {
		t.Error("Subscription should still be running after second Start()")
	}

	// Stop
	sub.Stop()
	cancel()

	// Give it time to stop
	time.Sleep(100 * time.Millisecond)

	if sub.IsRunning() {
		t.Error("Subscription should not be running after Stop()")
	}
}


