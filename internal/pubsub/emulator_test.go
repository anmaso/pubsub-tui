package pubsub

import (
	"os"
	"testing"
)

func TestIsEmulatorEnabled(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{
			name:     "emulator enabled with host:port",
			envValue: "localhost:8085",
			want:     true,
		},
		{
			name:     "emulator enabled with IP",
			envValue: "127.0.0.1:8085",
			want:     true,
		},
		{
			name:     "emulator disabled (empty)",
			envValue: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := os.Getenv(EmulatorHostEnvVar)
			defer os.Setenv(EmulatorHostEnvVar, original)

			// Set test value
			if tt.envValue == "" {
				os.Unsetenv(EmulatorHostEnvVar)
			} else {
				os.Setenv(EmulatorHostEnvVar, tt.envValue)
			}

			got := IsEmulatorEnabled()
			if got != tt.want {
				t.Errorf("IsEmulatorEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEmulatorHost(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{
			name:     "returns host:port",
			envValue: "localhost:8085",
			want:     "localhost:8085",
		},
		{
			name:     "returns empty when not set",
			envValue: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			original := os.Getenv(EmulatorHostEnvVar)
			defer os.Setenv(EmulatorHostEnvVar, original)

			// Set test value
			if tt.envValue == "" {
				os.Unsetenv(EmulatorHostEnvVar)
			} else {
				os.Setenv(EmulatorHostEnvVar, tt.envValue)
			}

			got := GetEmulatorHost()
			if got != tt.want {
				t.Errorf("GetEmulatorHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProjectID_EmulatorMode(t *testing.T) {
	// Save original environment
	originalEmulator := os.Getenv(EmulatorHostEnvVar)
	originalProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	originalGcloud := os.Getenv("GCLOUD_PROJECT")
	originalPubsub := os.Getenv("PUBSUB_PROJECT_ID")
	defer func() {
		os.Setenv(EmulatorHostEnvVar, originalEmulator)
		os.Setenv("GOOGLE_CLOUD_PROJECT", originalProject)
		os.Setenv("GCLOUD_PROJECT", originalGcloud)
		os.Setenv("PUBSUB_PROJECT_ID", originalPubsub)
	}()

	tests := []struct {
		name           string
		emulatorHost   string
		gcpProject     string
		gcloudProject  string
		pubsubProject  string
		wantProject    string
		wantErr        bool
		wantEmulatorErr bool
	}{
		{
			name:          "emulator mode with GOOGLE_CLOUD_PROJECT",
			emulatorHost:  "localhost:8085",
			gcpProject:    "my-test-project",
			wantProject:   "my-test-project",
			wantErr:       false,
		},
		{
			name:          "emulator mode with GCLOUD_PROJECT",
			emulatorHost:  "localhost:8085",
			gcloudProject: "gcloud-project",
			wantProject:   "gcloud-project",
			wantErr:       false,
		},
		{
			name:          "emulator mode with PUBSUB_PROJECT_ID",
			emulatorHost:  "localhost:8085",
			pubsubProject: "pubsub-project",
			wantProject:   "pubsub-project",
			wantErr:       false,
		},
		{
			name:           "emulator mode without project - should error",
			emulatorHost:   "localhost:8085",
			wantErr:        true,
			wantEmulatorErr: true,
		},
		{
			name:         "non-emulator mode with GOOGLE_CLOUD_PROJECT",
			emulatorHost: "",
			gcpProject:   "real-project",
			wantProject:  "real-project",
			wantErr:      false,
		},
		{
			name:          "GOOGLE_CLOUD_PROJECT takes precedence",
			emulatorHost:  "localhost:8085",
			gcpProject:    "gcp-project",
			gcloudProject: "gcloud-project",
			pubsubProject: "pubsub-project",
			wantProject:   "gcp-project",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			os.Unsetenv(EmulatorHostEnvVar)
			os.Unsetenv("GOOGLE_CLOUD_PROJECT")
			os.Unsetenv("GCLOUD_PROJECT")
			os.Unsetenv("PUBSUB_PROJECT_ID")

			// Set test values
			if tt.emulatorHost != "" {
				os.Setenv(EmulatorHostEnvVar, tt.emulatorHost)
			}
			if tt.gcpProject != "" {
				os.Setenv("GOOGLE_CLOUD_PROJECT", tt.gcpProject)
			}
			if tt.gcloudProject != "" {
				os.Setenv("GCLOUD_PROJECT", tt.gcloudProject)
			}
			if tt.pubsubProject != "" {
				os.Setenv("PUBSUB_PROJECT_ID", tt.pubsubProject)
			}

			got, err := GetProjectID()

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetProjectID() expected error, got nil")
					return
				}
				if tt.wantEmulatorErr {
					if pnfe, ok := err.(*ProjectNotFoundError); ok {
						if !pnfe.IsEmulatorMode() {
							t.Errorf("GetProjectID() error should indicate emulator mode")
						}
					} else {
						t.Errorf("GetProjectID() error should be ProjectNotFoundError")
					}
				}
				return
			}

			if err != nil {
				t.Errorf("GetProjectID() unexpected error: %v", err)
				return
			}

			if got != tt.wantProject {
				t.Errorf("GetProjectID() = %v, want %v", got, tt.wantProject)
			}
		})
	}
}

func TestProjectNotFoundError(t *testing.T) {
	t.Run("emulator mode error message", func(t *testing.T) {
		err := &ProjectNotFoundError{emulatorMode: true}
		msg := err.Error()
		if msg != "GCP project ID not found. Set GOOGLE_CLOUD_PROJECT or PUBSUB_PROJECT_ID for emulator mode." {
			t.Errorf("unexpected error message: %s", msg)
		}
		if !err.IsEmulatorMode() {
			t.Errorf("IsEmulatorMode() should return true")
		}
	})

	t.Run("non-emulator mode error message", func(t *testing.T) {
		err := &ProjectNotFoundError{emulatorMode: false}
		msg := err.Error()
		if msg != "GCP project ID not found. Set GOOGLE_CLOUD_PROJECT or configure gcloud." {
			t.Errorf("unexpected error message: %s", msg)
		}
		if err.IsEmulatorMode() {
			t.Errorf("IsEmulatorMode() should return false")
		}
	})
}

