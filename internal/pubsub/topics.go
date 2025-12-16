package pubsub

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/api/iterator"
)

// TopicInfo represents information about a Pub/Sub topic
type TopicInfo struct {
	Name     string // Short name (without project prefix)
	FullName string // Full resource name
}

// ListTopics retrieves all topics in the project
func (c *Client) ListTopics(ctx context.Context) ([]TopicInfo, error) {
	var topics []TopicInfo

	it := c.client.Topics(ctx)
	for {
		topic, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		topics = append(topics, TopicInfo{
			Name:     extractName(topic.ID()),
			FullName: topic.String(),
		})
	}

	return topics, nil
}

// CreateTopic creates a new topic with the given ID
func (c *Client) CreateTopic(ctx context.Context, topicID string) error {
	if err := validateResourceID(topicID); err != nil {
		return err
	}

	topic := c.client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check topic existence: %w", err)
	}
	if exists {
		return fmt.Errorf("topic %q already exists", topicID)
	}

	_, err = c.client.CreateTopic(ctx, topicID)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}

// DeleteTopic deletes a topic by ID
func (c *Client) DeleteTopic(ctx context.Context, topicID string) error {
	topic := c.client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check topic existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("topic %q does not exist", topicID)
	}

	if err := topic.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	return nil
}

// extractName extracts the short name from a full resource path
// e.g., "projects/my-project/topics/my-topic" -> "my-topic"
func extractName(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}

// validateResourceID validates a Pub/Sub resource ID
// Must be 3-255 characters, start with a letter, and contain only
// letters, numbers, dashes, periods, underscores, and tildes
func validateResourceID(id string) error {
	if len(id) < 3 || len(id) > 255 {
		return fmt.Errorf("resource ID must be 3-255 characters, got %d", len(id))
	}

	// Must start with a letter
	if !regexp.MustCompile(`^[a-zA-Z]`).MatchString(id) {
		return fmt.Errorf("resource ID must start with a letter")
	}

	// Can only contain letters, numbers, dashes, periods, underscores, tildes
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-._~]*$`).MatchString(id) {
		return fmt.Errorf("resource ID can only contain letters, numbers, dashes, periods, underscores, and tildes")
	}

	return nil
}
