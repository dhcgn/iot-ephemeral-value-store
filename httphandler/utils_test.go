package httphandler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func Test_sanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain string unchanged",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "escapes angle brackets",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "escapes ampersand",
			input:    "a & b",
			expected: "a &amp; b",
		},
		{
			name:     "escapes double quotes",
			input:    `say "hello"`,
			expected: "say &#34;hello&#34;",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeInput(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeInput(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func Test_jsonResponse(t *testing.T) {
	t.Run("sets content-type and encodes data", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"key": "value"}

		jsonResponse(w, data)

		if ct := w.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		var got map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal response body: %v", err)
		}
		if got["key"] != "value" {
			t.Errorf("expected key=value, got %v", got)
		}
	})
}

func Test_collectParams(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string][]string
		expected map[string]string
	}{
		{
			name:     "empty params",
			params:   map[string][]string{},
			expected: map[string]string{},
		},
		{
			name: "single param",
			params: map[string][]string{
				"temp": {"23.5"},
			},
			expected: map[string]string{
				"temp": "23.5",
			},
		},
		{
			name: "multiple values for key uses first",
			params: map[string][]string{
				"temp": {"23.5", "99.9"},
			},
			expected: map[string]string{
				"temp": "23.5",
			},
		},
		{
			name: "multiple keys",
			params: map[string][]string{
				"temp":     {"23.5"},
				"humidity": {"45"},
			},
			expected: map[string]string{
				"temp":     "23.5",
				"humidity": "45",
			},
		},
		{
			name: "values are sanitized",
			params: map[string][]string{
				"field": {"<b>bold</b>"},
			},
			expected: map[string]string{
				"field": "&lt;b&gt;bold&lt;/b&gt;",
			},
		},
		{
			name: "key with empty values slice is skipped",
			params: map[string][]string{
				"empty": {},
				"keep":  {"ok"},
			},
			expected: map[string]string{
				"keep": "ok",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectParams(tt.params)
			if len(got) != len(tt.expected) {
				t.Errorf("collectParams returned %d entries, want %d", len(got), len(tt.expected))
				return
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("collectParams[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
