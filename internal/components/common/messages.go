package common

import (
	"time"
)

// TopicSelectedMsg is sent when a topic is selected in the topics panel
type TopicSelectedMsg struct {
	TopicName string
	TopicFull string
}

// SubscriptionSelectedMsg is sent when a subscription is selected
type SubscriptionSelectedMsg struct {
	SubscriptionName string
	SubscriptionFull string
	TopicName        string
}

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogInfo LogLevel = iota
	LogSuccess
	LogWarning
	LogError
	LogNetwork
)

// LogMsg represents a message to be added to the activity log
type LogMsg struct {
	Level   LogLevel
	Message string
	Time    time.Time
}

// NewLogMsg creates a new log message with the current time
func NewLogMsg(level LogLevel, message string) LogMsg {
	return LogMsg{
		Level:   level,
		Message: message,
		Time:    time.Now(),
	}
}

// Info creates an info log message
func Info(message string) LogMsg {
	return NewLogMsg(LogInfo, message)
}

// Success creates a success log message
func Success(message string) LogMsg {
	return NewLogMsg(LogSuccess, message)
}

// Warning creates a warning log message
func Warning(message string) LogMsg {
	return NewLogMsg(LogWarning, message)
}

// Error creates an error log message
func Error(message string) LogMsg {
	return NewLogMsg(LogError, message)
}

// Network creates a network log message
func Network(message string) LogMsg {
	return NewLogMsg(LogNetwork, message)
}

// TopicsLoadedMsg is sent when topics are loaded from GCP
type TopicsLoadedMsg struct {
	Topics []TopicData
	Err    error
}

// SubscriptionsLoadedMsg is sent when subscriptions are loaded from GCP
type SubscriptionsLoadedMsg struct {
	Subscriptions []SubscriptionData
	Err           error
}

// TopicData represents topic data for UI display
type TopicData struct {
	Name     string
	FullName string
}

// SubscriptionData represents subscription data for UI display
type SubscriptionData struct {
	Name      string
	FullName  string
	TopicName string
	TopicFull string
}

// WindowSizeMsg is sent when the window size changes (re-exported for convenience)
type WindowSizeMsg struct {
	Width  int
	Height int
}

// FocusMsg indicates which panel should receive focus
type FocusMsg struct {
	Panel string
}

// TopicCreatedMsg is sent when a topic is created
type TopicCreatedMsg struct {
	TopicName string
	Err       error
}

// TopicDeletedMsg is sent when a topic is deleted
type TopicDeletedMsg struct {
	TopicName string
	Err       error
}

// SubscriptionCreatedMsg is sent when a subscription is created
type SubscriptionCreatedMsg struct {
	SubscriptionName string
	TopicName        string
	Err              error
}

// SubscriptionDeletedMsg is sent when a subscription is deleted
type SubscriptionDeletedMsg struct {
	SubscriptionName string
	Err              error
}

// RefreshTopicsMsg requests a refresh of the topics list
type RefreshTopicsMsg struct{}

// RefreshSubscriptionsMsg requests a refresh of the subscriptions list
type RefreshSubscriptionsMsg struct{}

// ConfirmDisconnectMsg is sent to confirm disconnecting an active subscription
type ConfirmDisconnectMsg struct {
	NewTopicName string
	Confirmed    bool
}

// StopSubscriptionMsg is sent to stop the active subscription
type StopSubscriptionMsg struct{}

// SubscriptionStoppedMsg is sent when the subscription has been stopped
type SubscriptionStoppedMsg struct{}
