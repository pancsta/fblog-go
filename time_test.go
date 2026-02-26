package fblog

import (
	"testing"
)

func TestTryConvertTimestampToReadable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Millis",
			input:    "1716292213381",
			expected: "2024-05-21T11:50:13.381Z",
		},
		{
			name:     "Seconds",
			input:    "1716292213",
			expected: "2024-05-21T11:50:13.000Z",
		},
		{
			name:     "Invalid String",
			input:    "bla",
			expected: "bla",
		},
		{
			name:     "String with Numbers",
			input:    "1234bla",
			expected: "1234bla",
		},
		{
			name:     "Empty String",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TryConvertTimestampToReadable(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
