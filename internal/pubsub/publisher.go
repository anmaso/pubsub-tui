package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
)

// PublishResult contains the result of a publish operation
type PublishResult struct {
	MessageID string
	Error     error
}

// Publish publishes a message to the specified topic
func (c *Client) Publish(ctx context.Context, topicName string, data []byte, attributes map[string]string) PublishResult {
	topic := c.client.Topic(topicName)

	msg := &pubsub.Message{
		Data:       data,
		Attributes: attributes,
	}

	result := topic.Publish(ctx, msg)

	// Block until the result is returned
	id, err := result.Get(ctx)
	if err != nil {
		return PublishResult{Error: err}
	}

	return PublishResult{MessageID: id}
}

// TopicExists checks if a topic exists
func (c *Client) TopicExists(ctx context.Context, topicName string) (bool, error) {
	topic := c.client.Topic(topicName)
	return topic.Exists(ctx)
}


