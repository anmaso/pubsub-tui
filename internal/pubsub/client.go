package pubsub

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the GCP Pub/Sub client with additional functionality
type Client struct {
	client    *pubsub.Client
	projectID string
}

// NewClient creates a new Pub/Sub client for the given project.
// When PUBSUB_EMULATOR_HOST is set, it connects to the emulator with
// insecure transport and no authentication.
func NewClient(projectID string) (*Client, error) {
	ctx := context.Background()

	var client *pubsub.Client
	var err error

	if IsEmulatorEnabled() {
		// Connect to emulator with insecure transport and no auth
		emulatorHost := GetEmulatorHost()
		conn, dialErr := grpc.Dial(
			emulatorHost,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if dialErr != nil {
			return nil, dialErr
		}
		client, err = pubsub.NewClient(ctx, projectID,
			option.WithGRPCConn(conn),
			option.WithoutAuthentication(),
		)
	} else {
		// Connect to real GCP with default credentials
		client, err = pubsub.NewClient(ctx, projectID)
	}

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

// GetProjectID retrieves the GCP project ID from environment or gcloud config.
// When emulator mode is enabled (PUBSUB_EMULATOR_HOST is set), only environment
// variables are checked and gcloud fallback is skipped.
func GetProjectID() (string, error) {
	// First check environment variable
	if projectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); projectID != "" {
		return projectID, nil
	}

	// Also check GCLOUD_PROJECT (alternative env var)
	if projectID := os.Getenv("GCLOUD_PROJECT"); projectID != "" {
		return projectID, nil
	}

	// Also check PUBSUB_PROJECT_ID (useful for emulator setups)
	if projectID := os.Getenv("PUBSUB_PROJECT_ID"); projectID != "" {
		return projectID, nil
	}

	// In emulator mode, do not fall back to gcloud - require explicit env var
	if IsEmulatorEnabled() {
		return "", &ProjectNotFoundError{emulatorMode: true}
	}

	// Fallback to gcloud config (only for real GCP)
	cmd := exec.Command("gcloud", "config", "get-value", "project")
	output, err := cmd.Output()
	if err != nil {
		return "", &ProjectNotFoundError{emulatorMode: false}
	}

	projectID := strings.TrimSpace(string(output))
	if projectID == "" || projectID == "(unset)" {
		return "", &ProjectNotFoundError{emulatorMode: false}
	}

	return projectID, nil
}

// ProjectNotFoundError indicates that no GCP project could be determined
type ProjectNotFoundError struct {
	emulatorMode bool
}

func (e *ProjectNotFoundError) Error() string {
	if e.emulatorMode {
		return "GCP project ID not found. Set GOOGLE_CLOUD_PROJECT or PUBSUB_PROJECT_ID for emulator mode."
	}
	return "GCP project ID not found. Set GOOGLE_CLOUD_PROJECT or configure gcloud."
}

// IsEmulatorMode returns whether the error occurred in emulator mode
func (e *ProjectNotFoundError) IsEmulatorMode() bool {
	return e.emulatorMode
}



