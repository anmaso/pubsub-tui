package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
)

// SubscriptionInfo represents information about a Pub/Sub subscription
type SubscriptionInfo struct {
	Name      string // Short name (without project prefix)
	FullName  string // Full resource name
	TopicName string // Associated topic short name
	TopicFull string // Associated topic full name
}

// ListSubscriptions retrieves all subscriptions in the project
func (c *Client) ListSubscriptions(ctx context.Context) ([]SubscriptionInfo, error) {
	var subscriptions []SubscriptionInfo

	it := c.client.Subscriptions(ctx)
	for {
		sub, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		// Get subscription config to retrieve the associated topic
		cfg, err := sub.Config(ctx)
		if err != nil {
			// If we can't get config, still include the subscription with empty topic
			subscriptions = append(subscriptions, SubscriptionInfo{
				Name:     extractName(sub.ID()),
				FullName: sub.String(),
			})
			continue
		}

		if cfg.Topic == nil {
			subscriptions = append(subscriptions, SubscriptionInfo{
				Name:     extractName(sub.ID()),
				FullName: sub.String(),
			})
			continue
		}

		// Safely extract topic name - avoid calling ID() if it might panic
		topicFull := cfg.Topic.String()
		topicName := extractName(topicFull)

		subscriptions = append(subscriptions, SubscriptionInfo{
			Name:      extractName(sub.ID()),
			FullName:  sub.String(),
			TopicName: topicName,
			TopicFull: topicFull,
		})
	}

	return subscriptions, nil
}

// GetSubscriptionInfo retrieves info for a single subscription by name
func (c *Client) GetSubscriptionInfo(ctx context.Context, subscriptionName string) (*SubscriptionInfo, error) {
	sub := c.client.Subscription(subscriptionName)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check subscription: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("subscription %q does not exist", subscriptionName)
	}

	cfg, err := sub.Config(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription config: %w", err)
	}

	info := &SubscriptionInfo{
		Name:     extractName(sub.ID()),
		FullName: sub.String(),
	}

	if cfg.Topic != nil {
		info.TopicFull = cfg.Topic.String()
		info.TopicName = extractName(info.TopicFull)
	}

	return info, nil
}

// CreateSubscription creates a new subscription for the given topic
func (c *Client) CreateSubscription(ctx context.Context, subscriptionID, topicID string) error {
	if err := validateResourceID(subscriptionID); err != nil {
		return err
	}

	sub := c.client.Subscription(subscriptionID)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check subscription existence: %w", err)
	}
	if exists {
		return fmt.Errorf("subscription %q already exists", subscriptionID)
	}

	topic := c.client.Topic(topicID)
	topicExists, err := topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check topic existence: %w", err)
	}
	if !topicExists {
		return fmt.Errorf("topic %q does not exist", topicID)
	}

	_, err = c.client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// DeleteSubscription deletes a subscription by ID
func (c *Client) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	sub := c.client.Subscription(subscriptionID)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check subscription existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("subscription %q does not exist", subscriptionID)
	}

	if err := sub.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}
