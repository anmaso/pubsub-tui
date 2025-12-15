package pubsub

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
)

// VerifyCredentials checks if valid GCP credentials are available
func VerifyCredentials() error {
	ctx := context.Background()

	// Find default credentials
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/pubsub")
	if err != nil {
		return fmt.Errorf("failed to find credentials: %w", err)
	}

	// Try to get a token to verify credentials are valid
	_, err = creds.TokenSource.Token()
	if err != nil {
		return fmt.Errorf("credentials invalid or expired: %w", err)
	}

	return nil
}

