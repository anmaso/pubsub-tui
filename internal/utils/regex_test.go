package utils

import (
	"testing"
)

func TestMatchesFilter(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		pattern     string
		wantMatches bool
		wantErr     bool
	}{
		// Empty pattern matches everything
		{
			name:        "empty pattern matches any text",
			text:        "hello world",
			pattern:     "",
			wantMatches: true,
			wantErr:     false,
		},
		{
			name:        "empty pattern matches empty text",
			text:        "",
			pattern:     "",
			wantMatches: true,
			wantErr:     false,
		},

		// Basic matching
		{
			name:        "simple substring match",
			text:        "hello world",
			pattern:     "world",
			wantMatches: true,
			wantErr:     false,
		},
		{
			name:        "no match",
			text:        "hello world",
			pattern:     "foo",
			wantMatches: false,
			wantErr:     false,
		},
		{
			name:        "case sensitive - no match",
			text:        "Hello World",
			pattern:     "hello",
			wantMatches: false,
			wantErr:     false,
		},

		// Regex patterns
		{
			name:        "regex with dot wildcard",
			text:        "hello world",
			pattern:     "hel.o",
			wantMatches: true,
			wantErr:     false,
		},
		{
			name:        "regex with star quantifier",
			text:        "hellooooo",
			pattern:     "hel+o",
			wantMatches: true,
			wantErr:     false,
		},
		{
			name:        "regex with anchors - match",
			text:        "hello",
			pattern:     "^hello$",
			wantMatches: true,
			wantErr:     false,
		},
		{
			name:        "regex with anchors - no match",
			text:        "hello world",
			pattern:     "^hello$",
			wantMatches: false,
			wantErr:     false,
		},
		{
			name:        "regex with character class",
			text:        "test123",
			pattern:     "[0-9]+",
			wantMatches: true,
			wantErr:     false,
		},

		// Invalid regex
		{
			name:        "invalid regex - unclosed bracket",
			text:        "hello",
			pattern:     "[hello",
			wantMatches: false,
			wantErr:     true,
		},
		{
			name:        "invalid regex - bad quantifier",
			text:        "hello",
			pattern:     "*hello",
			wantMatches: false,
			wantErr:     true,
		},
		{
			name:        "invalid regex - unclosed paren",
			text:        "hello",
			pattern:     "(hello",
			wantMatches: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesFilter(tt.text, tt.pattern)

			if tt.wantErr {
				if result.Error == nil {
					t.Errorf("MatchesFilter(%q, %q) expected error, got nil", tt.text, tt.pattern)
				}
			} else {
				if result.Error != nil {
					t.Errorf("MatchesFilter(%q, %q) unexpected error: %v", tt.text, tt.pattern, result.Error)
				}
				if result.Matches != tt.wantMatches {
					t.Errorf("MatchesFilter(%q, %q).Matches = %v, want %v", tt.text, tt.pattern, result.Matches, tt.wantMatches)
				}
			}
		})
	}
}

func TestValidateRegex(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{
			name:    "empty pattern is valid",
			pattern: "",
			wantErr: false,
		},
		{
			name:    "simple pattern",
			pattern: "hello",
			wantErr: false,
		},
		{
			name:    "complex valid pattern",
			pattern: `^[a-zA-Z0-9_.-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			wantErr: false,
		},
		{
			name:    "invalid - unclosed bracket",
			pattern: "[abc",
			wantErr: true,
		},
		{
			name:    "invalid - bad escape",
			pattern: `\`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegex(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateRegex(%q) expected error, got nil", tt.pattern)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateRegex(%q) unexpected error: %v", tt.pattern, err)
				}
			}
		})
	}
}

func TestFilterStrings(t *testing.T) {
	tests := []struct {
		name    string
		items   []string
		pattern string
		want    []string
		wantErr bool
	}{
		{
			name:    "empty pattern returns all",
			items:   []string{"apple", "banana", "cherry"},
			pattern: "",
			want:    []string{"apple", "banana", "cherry"},
			wantErr: false,
		},
		{
			name:    "filter by prefix",
			items:   []string{"apple", "apricot", "banana", "avocado"},
			pattern: "^a",
			want:    []string{"apple", "apricot", "avocado"},
			wantErr: false,
		},
		{
			name:    "filter by suffix",
			items:   []string{"test.json", "data.json", "config.yaml"},
			pattern: `\.json$`,
			want:    []string{"test.json", "data.json"},
			wantErr: false,
		},
		{
			name:    "no matches",
			items:   []string{"apple", "banana", "cherry"},
			pattern: "xyz",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "empty items",
			items:   []string{},
			pattern: "test",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "invalid regex returns original items",
			items:   []string{"apple", "banana"},
			pattern: "[invalid",
			want:    []string{"apple", "banana"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterStrings(tt.items, tt.pattern)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FilterStrings() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("FilterStrings() unexpected error: %v", err)
				}
			}

			if !stringSliceEqual(got, tt.want) {
				t.Errorf("FilterStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}


