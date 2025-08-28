package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with spaces",
			input:    "HeLlo World",
			expected: "hello-world",
		},
		{
			name:     "with special characters",
			input:    "Hello@#$%world!",
			expected: "helloworld",
		},
		{
			name:     "with numbers",
			input:    "Test 123 test",
			expected: "test-123-test",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "already slug",
			input:    "hello-world",
			expected: "hello-world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CreateSlug(tt.input)

			assert.Equal(t, tt.expected, result, "CreateSlug(%q) should return %q, got %q", tt.input, tt.expected, result)
		})
	}
}
