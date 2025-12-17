package main

import (
	"fmt"
	"os"

	"github.com/anmaso/pubsub-tui/internal/app"
	"github.com/anmaso/pubsub-tui/internal/pubsub"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	emulatorMode := pubsub.IsEmulatorEnabled()

	// Verify GCP credentials and project before starting TUI
	projectID, err := pubsub.GetProjectID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if emulatorMode {
			fmt.Fprintf(os.Stderr, "\nEmulator mode detected (PUBSUB_EMULATOR_HOST is set).\n")
			fmt.Fprintf(os.Stderr, "To fix this, set a project ID:\n")
			fmt.Fprintf(os.Stderr, "  export GOOGLE_CLOUD_PROJECT=your-project-id\n")
		} else {
			fmt.Fprintf(os.Stderr, "\nTo fix this, either:\n")
			fmt.Fprintf(os.Stderr, "  1. Set GOOGLE_CLOUD_PROJECT environment variable\n")
			fmt.Fprintf(os.Stderr, "  2. Run: gcloud config set project YOUR_PROJECT_ID\n")
		}
		os.Exit(1)
	}

	// Verify credentials (skipped in emulator mode)
	if err := pubsub.VerifyCredentials(); err != nil {
		fmt.Fprintf(os.Stderr, "Authentication error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nTo authenticate, run:\n")
		fmt.Fprintf(os.Stderr, "  gcloud auth application-default login\n")
		os.Exit(1)
	}

	// Create Pub/Sub client
	client, err := pubsub.NewClient(projectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Pub/Sub client: %v\n", err)
		if emulatorMode {
			fmt.Fprintf(os.Stderr, "\nEmulator mode: ensure the emulator is running at %s\n", pubsub.GetEmulatorHost())
		}
		os.Exit(1)
	}
	defer client.Close()

	// Print startup info
	if emulatorMode {
		fmt.Fprintf(os.Stderr, "Connecting to Pub/Sub emulator at %s...\n", pubsub.GetEmulatorHost())
	}

	// Initialize and run the TUI application
	p := tea.NewProgram(
		app.New(client, projectID),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}



