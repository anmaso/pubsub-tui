package pubsub

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"cloud.google.com/go/pubsub"
)

// Client wraps the GCP Pub/Sub client with additional functionality
type Client struct {
	client    *pubsub.Client
	projectID string
}

// NewClient creates a new Pub/Sub client for the given project
func NewClient(projectID string) (*Client, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:    client,
		projectID: projectID,
	}, nil
}

// Close closes the underlying Pub/Sub client
func (c *Client) Close() error {
	return c.client.Close()
}

// ProjectID returns the project ID
func (c *Client) ProjectID() string {
	return c.projectID
}

// GetProjectID retrieves the GCP project ID from environment or gcloud config
func GetProjectID() (string, error) {
	// First check environment variable
	if projectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); projectID != "" {
		return projectID, nil
	}

	// Also check GCLOUD_PROJECT (alternative env var)
	if projectID := os.Getenv("GCLOUD_PROJECT"); projectID != "" {
		return projectID, nil
	}

	// Fallback to gcloud config
	cmd := exec.Command("gcloud", "config", "get-value", "project")
	output, err := cmd.Output()
	if err != nil {
		return "", &ProjectNotFoundError{}
	}

	projectID := strings.TrimSpace(string(output))
	if projectID == "" || projectID == "(unset)" {
		return "", &ProjectNotFoundError{}
	}

	return projectID, nil
}

// ProjectNotFoundError indicates that no GCP project could be determined
type ProjectNotFoundError struct{}

func (e *ProjectNotFoundError) Error() string {
	return "GCP project ID not found. Set GOOGLE_CLOUD_PROJECT or configure gcloud."
}

