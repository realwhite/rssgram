package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEllipsisString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "Hello World",
			maxLen:   20,
			expected: "Hello World",
		},
		{
			name:     "exact length",
			input:    "Hello World",
			maxLen:   11,
			expected: "Hello World",
		},
		{
			name:     "long string with ellipsis",
			input:    "This is a very long string that needs to be truncated",
			maxLen:   20,
			expected: "This is a very long...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "maxLen 0",
			input:    "Hello",
			maxLen:   0,
			expected: "",
		},
		{
			name:     "maxLen negative",
			input:    "Hello",
			maxLen:   -5,
			expected: "",
		},
		{
			name:     "unicode string",
			input:    "Привет мир",
			maxLen:   8,
			expected: "Прив...",
		},
		{
			name:     "unicode string exact",
			input:    "Привет",
			maxLen:   6,
			expected: "При...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EllipsisString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEllipsisString_EdgeCases(t *testing.T) {
	longString := "This is a very long string that contains many characters and should be truncated properly when the maximum length is exceeded"
	result := EllipsisString(longString, 50)
	// Ожидаем, что результат заканчивается на "..."
	assert.True(t, len(result) <= 50+3)
	assert.Equal(t, "...", result[len(result)-3:])

	result = EllipsisString("Hello", 2)
	assert.Equal(t, "He...", result)
}

func TestEllipsisString_Benchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping benchmark in short mode")
	}

	longString := "This is a benchmark test string that will be used to measure the performance of the EllipsisString function"

	t.Run("benchmark", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			EllipsisString(longString, 50)
		}
	})
}
