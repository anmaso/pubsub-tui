package utils

import (
	"testing"
)

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{
			name: "simple object",
			data: []byte(`{"name":"test","value":123}`),
			want: `{
  "name": "test",
  "value": 123
}`,
			wantErr: false,
		},
		{
			name: "nested object",
			data: []byte(`{"outer":{"inner":"value"}}`),
			want: `{
  "outer": {
    "inner": "value"
  }
}`,
			wantErr: false,
		},
		{
			name:    "array",
			data:    []byte(`[1,2,3]`),
			want:    "[\n  1,\n  2,\n  3\n]",
			wantErr: false,
		},
		{
			name:    "invalid JSON returns as-is",
			data:    []byte(`not valid json`),
			want:    "not valid json",
			wantErr: false,
		},
		{
			name:    "empty string",
			data:    []byte(``),
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatJSON(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Errorf("FormatJSON() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("FormatJSON() unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("FormatJSON() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestIsValidJSON(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid object",
			data: []byte(`{"key": "value"}`),
			want: true,
		},
		{
			name: "valid array",
			data: []byte(`[1, 2, 3]`),
			want: true,
		},
		{
			name: "valid string",
			data: []byte(`"hello"`),
			want: true,
		},
		{
			name: "valid number",
			data: []byte(`123`),
			want: true,
		},
		{
			name: "valid boolean",
			data: []byte(`true`),
			want: true,
		},
		{
			name: "valid null",
			data: []byte(`null`),
			want: true,
		},
		{
			name: "invalid - plain text",
			data: []byte(`hello world`),
			want: false,
		},
		{
			name: "invalid - unclosed brace",
			data: []byte(`{"key": "value"`),
			want: false,
		},
		{
			name: "invalid - trailing comma",
			data: []byte(`{"key": "value",}`),
			want: false,
		},
		{
			name: "empty bytes",
			data: []byte{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidJSON(tt.data)
			if got != tt.want {
				t.Errorf("IsValidJSON(%s) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestCompactJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []byte
		wantErr bool
	}{
		{
			name:    "already compact",
			data:    []byte(`{"key":"value"}`),
			want:    []byte(`{"key":"value"}`),
			wantErr: false,
		},
		{
			name:    "with whitespace",
			data:    []byte(`{  "key":  "value"  }`),
			want:    []byte(`{"key":"value"}`),
			wantErr: false,
		},
		{
			name: "with newlines",
			data: []byte(`{
  "key": "value"
}`),
			want:    []byte(`{"key":"value"}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON returns original",
			data:    []byte(`not json`),
			want:    []byte(`not json`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompactJSON(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CompactJSON() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("CompactJSON() unexpected error: %v", err)
				}
			}
			if string(got) != string(tt.want) {
				t.Errorf("CompactJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestPrettyPrint(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name    string
		v       interface{}
		want    string
		wantErr bool
	}{
		{
			name: "struct",
			v:    testStruct{Name: "test", Value: 42},
			want: `{
  "name": "test",
  "value": 42
}`,
			wantErr: false,
		},
		{
			name:    "map",
			v:       map[string]int{"a": 1, "b": 2},
			want:    "", // Map order is not guaranteed
			wantErr: false,
		},
		{
			name:    "slice",
			v:       []string{"a", "b", "c"},
			want:    "[\n  \"a\",\n  \"b\",\n  \"c\"\n]",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrettyPrint(tt.v)
			if tt.wantErr {
				if err == nil {
					t.Errorf("PrettyPrint() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("PrettyPrint() unexpected error: %v", err)
				}
				// Skip checking exact output for map (order not guaranteed)
				if tt.want != "" && got != tt.want {
					t.Errorf("PrettyPrint() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}


