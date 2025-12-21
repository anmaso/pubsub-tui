package pubsub

import (
	"testing"
)

func TestValidateResourceID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
		errMsg  string
	}{
		// Valid IDs
		{
			name:    "valid simple ID",
			id:      "my-topic",
			wantErr: false,
		},
		{
			name:    "valid ID with numbers",
			id:      "topic123",
			wantErr: false,
		},
		{
			name:    "valid ID with all allowed chars",
			id:      "my-topic_v1.0~test",
			wantErr: false,
		},
		{
			name:    "valid minimum length (3 chars)",
			id:      "abc",
			wantErr: false,
		},
		{
			name:    "valid ID starting with uppercase",
			id:      "MyTopic",
			wantErr: false,
		},

		// Invalid: length constraints
		{
			name:    "too short (2 chars)",
			id:      "ab",
			wantErr: true,
			errMsg:  "resource ID must be 3-255 characters",
		},
		{
			name:    "too short (1 char)",
			id:      "a",
			wantErr: true,
			errMsg:  "resource ID must be 3-255 characters",
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
			errMsg:  "resource ID must be 3-255 characters",
		},

		// Invalid: must start with letter
		{
			name:    "starts with number",
			id:      "1topic",
			wantErr: true,
			errMsg:  "resource ID must start with a letter",
		},
		{
			name:    "starts with dash",
			id:      "-topic",
			wantErr: true,
			errMsg:  "resource ID must start with a letter",
		},
		{
			name:    "starts with underscore",
			id:      "_topic",
			wantErr: true,
			errMsg:  "resource ID must start with a letter",
		},

		// Invalid: illegal characters
		{
			name:    "contains space",
			id:      "my topic",
			wantErr: true,
			errMsg:  "resource ID can only contain",
		},
		{
			name:    "contains at sign",
			id:      "topic@test",
			wantErr: true,
			errMsg:  "resource ID can only contain",
		},
		{
			name:    "contains slash",
			id:      "topic/test",
			wantErr: true,
			errMsg:  "resource ID can only contain",
		},
		{
			name:    "contains hash",
			id:      "topic#1",
			wantErr: true,
			errMsg:  "resource ID can only contain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceID(tt.id)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateResourceID(%q) expected error, got nil", tt.id)
					return
				}
				if tt.errMsg != "" && !containsSubstring(err.Error(), tt.errMsg) {
					t.Errorf("validateResourceID(%q) error = %q, want error containing %q", tt.id, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateResourceID(%q) unexpected error: %v", tt.id, err)
				}
			}
		})
	}
}

func TestValidateResourceID_MaxLength(t *testing.T) {
	// Test exactly 255 characters (valid)
	validLongID := "a" + string(make([]byte, 254))
	for i := 1; i < 255; i++ {
		validLongID = validLongID[:i] + "a" + validLongID[i+1:]
	}
	validLongID = make255CharID()
	if err := validateResourceID(validLongID); err != nil {
		t.Errorf("validateResourceID(255 chars) should be valid, got: %v", err)
	}

	// Test 256 characters (invalid)
	invalidLongID := validLongID + "x"
	if err := validateResourceID(invalidLongID); err == nil {
		t.Error("validateResourceID(256 chars) should be invalid")
	}
}

func make255CharID() string {
	id := make([]byte, 255)
	id[0] = 'a' // Must start with letter
	for i := 1; i < 255; i++ {
		id[i] = 'b'
	}
	return string(id)
}

func TestExtractName(t *testing.T) {
	tests := []struct {
		name     string
		fullPath string
		want     string
	}{
		{
			name:     "standard topic path",
			fullPath: "projects/my-project/topics/my-topic",
			want:     "my-topic",
		},
		{
			name:     "standard subscription path",
			fullPath: "projects/my-project/subscriptions/my-sub",
			want:     "my-sub",
		},
		{
			name:     "path with dashes in project",
			fullPath: "projects/my-gcp-project/topics/test-topic",
			want:     "test-topic",
		},
		{
			name:     "just the name (no path)",
			fullPath: "my-topic",
			want:     "my-topic",
		},
		{
			name:     "empty string",
			fullPath: "",
			want:     "",
		},
		{
			name:     "path with trailing slash",
			fullPath: "projects/my-project/topics/",
			want:     "",
		},
		{
			name:     "single slash",
			fullPath: "/",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractName(tt.fullPath)
			if got != tt.want {
				t.Errorf("extractName(%q) = %q, want %q", tt.fullPath, got, tt.want)
			}
		})
	}
}

// Helper to check if string contains substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}


