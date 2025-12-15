package app

import (
	"context"

	"pubsub-tui/internal/components/activity"
	"pubsub-tui/internal/components/common"
	"pubsub-tui/internal/components/publisher"
	"pubsub-tui/internal/components/subscriber"
	"pubsub-tui/internal/components/subscriptions"
	"pubsub-tui/internal/components/topics"
	"pubsub-tui/internal/pubsub"

	tea "github.com/charmbracelet/bubbletea"
)

// FocusPanel represents which panel is currently focused
type FocusPanel string

const (
	FocusTopics        FocusPanel = "topics"
	FocusSubscriptions FocusPanel = "subscriptions"
	FocusPublisher     FocusPanel = "publisher"
	FocusSubscriber    FocusPanel = "subscriber"
)

// Model is the main application model
type Model struct {
	// Pub/Sub client
	client    *pubsub.Client
	projectID string

	// Child components
	topics        topics.Model
	subscriptions subscriptions.Model
	publisher     publisher.Model
	subscriber    subscriber.Model
	activity      activity.Model

	// Subscription management
	activeSubscription *pubsub.Subscription
	subscriptionCtx    context.Context
	subscriptionCancel context.CancelFunc

	// UI state
	focus    FocusPanel
	width    int
	height   int
	ready    bool
	showHelp bool

	// Selected state
	selectedTopic        string
	selectedSubscription string
}

// New creates a new application model
func New(client *pubsub.Client, projectID string) Model {
	return Model{
		client:        client,
		projectID:     projectID,
		topics:        topics.New(),
		subscriptions: subscriptions.New(),
		publisher:     publisher.New(),
		subscriber:    subscriber.New(),
		activity:      activity.New(),
		focus:         FocusTopics,
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadTopics(),
		m.loadSubscriptions(),
		publisher.LoadFiles(),
		func() tea.Msg {
			return common.Info("Application started")
		},
		func() tea.Msg {
			return common.Network("Connected to project: " + m.projectID)
		},
	)
}

// loadTopics loads topics from GCP
func (m Model) loadTopics() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		topicsList, err := m.client.ListTopics(ctx)
		if err != nil {
			return common.TopicsLoadedMsg{Err: err}
		}

		var topics []common.TopicData
		for _, t := range topicsList {
			topics = append(topics, common.TopicData{
				Name:     t.Name,
				FullName: t.FullName,
			})
		}

		return common.TopicsLoadedMsg{Topics: topics}
	}
}

// loadSubscriptions loads subscriptions from GCP
func (m Model) loadSubscriptions() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		subsList, err := m.client.ListSubscriptions(ctx)
		if err != nil {
			return common.SubscriptionsLoadedMsg{Err: err}
		}

		var subs []common.SubscriptionData
		for _, s := range subsList {
			subs = append(subs, common.SubscriptionData{
				Name:      s.Name,
				FullName:  s.FullName,
				TopicName: s.TopicName,
				TopicFull: s.TopicFull,
			})
		}

		return common.SubscriptionsLoadedMsg{Subscriptions: subs}
	}
}

// startSubscription starts receiving messages from a subscription
func (m *Model) startSubscription(subName, topicName string) tea.Cmd {
	// Stop existing subscription first
	m.stopSubscription()

	// Create new subscription
	m.activeSubscription = m.client.Subscribe(subName)
	m.subscriptionCtx, m.subscriptionCancel = context.WithCancel(context.Background())

	// Start receiving
	m.activeSubscription.Start(m.subscriptionCtx)

	// Return command that polls for messages
	return m.pollMessages()
}

// stopSubscription stops the active subscription
func (m *Model) stopSubscription() {
	if m.activeSubscription != nil {
		m.activeSubscription.Stop()
		m.activeSubscription = nil
	}
	if m.subscriptionCancel != nil {
		m.subscriptionCancel()
		m.subscriptionCancel = nil
	}
}

// pollMessages returns a command that polls for new messages
func (m *Model) pollMessages() tea.Cmd {
	if m.activeSubscription == nil {
		return nil
	}

	sub := m.activeSubscription
	return func() tea.Msg {
		select {
		case msg, ok := <-sub.Messages():
			if !ok {
				return nil
			}
			return subscriber.MessageReceivedMsg{Message: msg}
		case err, ok := <-sub.Errors():
			if !ok {
				return nil
			}
			return subscriber.SubscriptionErrorMsg{Error: err}
		}
	}
}

// publishMessage publishes a message to the topic
func (m *Model) publishMessage(topic string, content []byte) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		result := m.client.Publish(ctx, topic, content, nil)
		return publisher.PublishResultMsg{
			MessageID: result.MessageID,
			Err:       result.Error,
		}
	}
}
