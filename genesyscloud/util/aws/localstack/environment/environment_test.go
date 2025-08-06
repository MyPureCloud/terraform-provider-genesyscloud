package environment

import (
	"os"
	"testing"
)

func TestUnitGetLocalStackPort(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "returns default when env var not set",
			envValue: "",
			expected: defaultLocalStackPort,
		},
		{
			name:     "returns env var value when set",
			envValue: "8080",
			expected: "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment after test
			defer os.Unsetenv(localStackPortEnvVar)

			if tt.envValue != "" {
				os.Setenv(localStackPortEnvVar, tt.envValue)
			} else {
				os.Unsetenv(localStackPortEnvVar)
			}

			result := GetLocalStackPort()
			if result != tt.expected {
				t.Errorf("GetLocalStackPort() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUnitLocalStackIsActive(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "returns false when env var not set",
			envValue: "",
			expected: false,
		},
		{
			name:     "returns true when env var is 'true'",
			envValue: "true",
			expected: true,
		},
		{
			name:     "returns false when env var is 'false'",
			envValue: "false",
			expected: false,
		},
		{
			name:     "returns false when env var is anything else",
			envValue: "other",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment after test
			defer os.Unsetenv(UseLocalStackEnvVar)

			if tt.envValue != "" {
				os.Setenv(UseLocalStackEnvVar, tt.envValue)
			} else {
				os.Unsetenv(UseLocalStackEnvVar)
			}

			result := LocalStackIsActive()
			if result != tt.expected {
				t.Errorf("LocalStackIsActive() = %v, want %v", result, tt.expected)
			}
		})
	}
}
