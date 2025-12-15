package pubsub

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
)

// ReceivedMessage represents a message received from a subscription
type ReceivedMessage struct {
	ID          string
	Data        []byte
	Attributes  map[string]string
	PublishTime time.Time
	AckID       string

	// Internal fields for ack/nack
	ackFunc  func()
	nackFunc func()
	acked    bool
	mu       sync.Mutex
}

// Ack acknowledges the message
func (m *ReceivedMessage) Ack() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.acked && m.ackFunc != nil {
		m.ackFunc()
		m.acked = true
	}
}

// Nack negative-acknowledges the message (will be redelivered)
func (m *ReceivedMessage) Nack() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.acked && m.nackFunc != nil {
		m.nackFunc()
	}
}

// IsAcked returns whether the message has been acknowledged
func (m *ReceivedMessage) IsAcked() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.acked
}

// SetAcked marks the message as acknowledged (for display purposes)
func (m *ReceivedMessage) SetAcked(acked bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.acked = acked
}

// Subscription wraps a Pub/Sub subscription for streaming messages
type Subscription struct {
	client       *Client
	subscription *pubsub.Subscription
	cancel       context.CancelFunc
	messages     chan *ReceivedMessage
	errors       chan error
	running      bool
	mu           sync.Mutex
}

// Subscribe creates a new subscription stream
func (c *Client) Subscribe(subscriptionName string) *Subscription {
	sub := c.client.Subscription(subscriptionName)

	// Configure subscription settings
	sub.ReceiveSettings.MaxOutstandingMessages = 100
	sub.ReceiveSettings.MaxOutstandingBytes = 10 * 1024 * 1024 // 10 MB

	return &Subscription{
		client:       c,
		subscription: sub,
		messages:     make(chan *ReceivedMessage, 100),
		errors:       make(chan error, 10),
	}
}

// Start begins receiving messages from the subscription
func (s *Subscription) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	// Create cancellable context
	ctx, s.cancel = context.WithCancel(ctx)

	go func() {
		err := s.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			received := &ReceivedMessage{
				ID:          msg.ID,
				Data:        msg.Data,
				Attributes:  msg.Attributes,
				PublishTime: msg.PublishTime,
				AckID:       msg.ID,
				ackFunc:     msg.Ack,
				nackFunc:    msg.Nack,
			}

			select {
			case s.messages <- received:
			case <-ctx.Done():
				msg.Nack()
				return
			}
		})

		if err != nil && ctx.Err() == nil {
			select {
			case s.errors <- err:
			default:
			}
		}

		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()
}

// Stop stops receiving messages
func (s *Subscription) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.running = false
}

// Messages returns the channel for receiving messages
func (s *Subscription) Messages() <-chan *ReceivedMessage {
	return s.messages
}

// Errors returns the channel for receiving errors
func (s *Subscription) Errors() <-chan error {
	return s.errors
}

// IsRunning returns whether the subscription is actively receiving
func (s *Subscription) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// SubscriptionExists checks if a subscription exists
func (c *Client) SubscriptionExists(ctx context.Context, subscriptionName string) (bool, error) {
	sub := c.client.Subscription(subscriptionName)
	return sub.Exists(ctx)
}

