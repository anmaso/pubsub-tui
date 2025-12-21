package pubsub

import (
	"sync"
	"testing"
)

func TestReceivedMessage_Ack(t *testing.T) {
	ackCalled := false
	msg := &ReceivedMessage{
		ID:   "test-msg-1",
		Data: []byte("test data"),
		ackFunc: func() {
			ackCalled = true
		},
	}

	// Initially not acked
	if msg.IsAcked() {
		t.Error("message should not be acked initially")
	}

	// Ack the message
	msg.Ack()

	if !ackCalled {
		t.Error("ackFunc should have been called")
	}
	if !msg.IsAcked() {
		t.Error("message should be acked after Ack()")
	}
}

func TestReceivedMessage_Ack_Idempotent(t *testing.T) {
	ackCount := 0
	msg := &ReceivedMessage{
		ID:   "test-msg-2",
		Data: []byte("test data"),
		ackFunc: func() {
			ackCount++
		},
	}

	// Ack multiple times
	msg.Ack()
	msg.Ack()
	msg.Ack()

	// Should only call ackFunc once
	if ackCount != 1 {
		t.Errorf("ackFunc should be called exactly once, was called %d times", ackCount)
	}
}

func TestReceivedMessage_Nack(t *testing.T) {
	nackCalled := false
	msg := &ReceivedMessage{
		ID:   "test-msg-3",
		Data: []byte("test data"),
		nackFunc: func() {
			nackCalled = true
		},
	}

	// Nack the message
	msg.Nack()

	if !nackCalled {
		t.Error("nackFunc should have been called")
	}

	// Nack should NOT mark as acked
	if msg.IsAcked() {
		t.Error("Nack should not mark message as acked")
	}
}

func TestReceivedMessage_Nack_AfterAck_NoOp(t *testing.T) {
	ackCalled := false
	nackCalled := false
	msg := &ReceivedMessage{
		ID:   "test-msg-4",
		Data: []byte("test data"),
		ackFunc: func() {
			ackCalled = true
		},
		nackFunc: func() {
			nackCalled = true
		},
	}

	// Ack first
	msg.Ack()

	// Then try to nack
	msg.Nack()

	if !ackCalled {
		t.Error("ackFunc should have been called")
	}
	if nackCalled {
		t.Error("nackFunc should NOT be called after message is acked")
	}
}

func TestReceivedMessage_SetAcked(t *testing.T) {
	msg := &ReceivedMessage{
		ID:   "test-msg-5",
		Data: []byte("test data"),
	}

	if msg.IsAcked() {
		t.Error("message should not be acked initially")
	}

	msg.SetAcked(true)
	if !msg.IsAcked() {
		t.Error("message should be acked after SetAcked(true)")
	}

	msg.SetAcked(false)
	if msg.IsAcked() {
		t.Error("message should not be acked after SetAcked(false)")
	}
}

func TestReceivedMessage_NilFuncs(t *testing.T) {
	// Test that Ack/Nack don't panic with nil funcs
	msg := &ReceivedMessage{
		ID:      "test-msg-6",
		Data:    []byte("test data"),
		ackFunc: nil,
	}

	// Should not panic
	msg.Ack()
	msg.Nack()

	// Message should not be marked as acked when ackFunc is nil
	if msg.IsAcked() {
		t.Error("message should not be acked when ackFunc is nil")
	}
}

func TestReceivedMessage_Concurrency(t *testing.T) {
	ackCount := 0
	var mu sync.Mutex
	msg := &ReceivedMessage{
		ID:   "test-msg-7",
		Data: []byte("test data"),
		ackFunc: func() {
			mu.Lock()
			ackCount++
			mu.Unlock()
		},
	}

	// Attempt concurrent acks
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg.Ack()
		}()
	}
	wg.Wait()

	// Should only have called ackFunc once
	if ackCount != 1 {
		t.Errorf("ackFunc should be called exactly once under concurrent access, was called %d times", ackCount)
	}
}

func TestSubscription_IsRunning(t *testing.T) {
	sub := &Subscription{
		running:  false,
		messages: make(chan *ReceivedMessage, 10),
		errors:   make(chan error, 10),
	}

	if sub.IsRunning() {
		t.Error("subscription should not be running initially")
	}

	sub.mu.Lock()
	sub.running = true
	sub.mu.Unlock()

	if !sub.IsRunning() {
		t.Error("subscription should be running after setting flag")
	}
}

func TestSubscription_Channels(t *testing.T) {
	messages := make(chan *ReceivedMessage, 10)
	errors := make(chan error, 10)

	sub := &Subscription{
		messages: messages,
		errors:   errors,
	}

	// Verify channels are returned correctly
	if sub.Messages() != messages {
		t.Error("Messages() should return the messages channel")
	}
	if sub.Errors() != errors {
		t.Error("Errors() should return the errors channel")
	}
}


