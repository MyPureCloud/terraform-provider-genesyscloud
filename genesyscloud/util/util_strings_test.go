package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripInvisibleUnicodeFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no invisible characters",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "non-breaking space replaced with regular space",
			input:    "hello\u00A0world",
			expected: "hello world",
		},
		{
			name:     "zero-width space removed",
			input:    "hello\u200Bworld",
			expected: "helloworld",
		},
		{
			name:     "zero-width non-joiner removed",
			input:    "hello\u200Cworld",
			expected: "helloworld",
		},
		{
			name:     "zero-width joiner removed",
			input:    "hello\u200Dworld",
			expected: "helloworld",
		},
		{
			name:     "word joiner removed",
			input:    "hello\u2060world",
			expected: "helloworld",
		},
		{
			name:     "byte order mark removed",
			input:    "\uFEFFhello world",
			expected: "hello world",
		},
		{
			name:     "multiple invisible characters",
			input:    "\uFEFFhello\u200B\u200C\u00A0world\u2060",
			expected: "hello world",
		},
		{
			name:     "only invisible characters",
			input:    "\u200B\u200C\u200D\u2060\uFEFF",
			expected: "",
		},
		{
			name:     "multiple non-breaking spaces",
			input:    "a\u00A0b\u00A0c",
			expected: "a b c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripInvisibleUnicodeFromString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CamelCase", "camel_case"},
		{"camelCase", "camel_case"},
		{"HTMLParser", "html_parser"},
		{"simpletest", "simpletest"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToSnakeCase(tt.input))
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"snake_case", "snakeCase"},
		{"one_two_three", "oneTwoThree"},
		{"nounderscores", "nounderscores"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToCamelCase(tt.input))
		})
	}
}

func TestStringExists(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.True(t, StringExists("b", slice))
	assert.False(t, StringExists("d", slice))
	assert.False(t, StringExists("a", nil))
}

func TestStringOrNil(t *testing.T) {
	s := "hello"
	assert.Equal(t, "hello", StringOrNil(&s))
	assert.Equal(t, "nil", StringOrNil(nil))
}
