package provider

import (
	"os"
	"testing"
	"time"
)

func TestUnitGetCustomRetryTimeout_Default(t *testing.T) {
	// Clear any existing environment variable
	originalEnv := os.Getenv(customRetryTimeoutEnvVar)
	os.Unsetenv(customRetryTimeoutEnvVar)
	defer func() {
		if originalEnv != "" {
			os.Setenv(customRetryTimeoutEnvVar, originalEnv)
		}
	}()

	// Clear provider meta/config
	originalMeta := providerMeta
	providerMeta = nil
	defer func() { providerMeta = originalMeta }()

	originalConfig := providerConfig
	providerConfig = nil
	defer func() { providerConfig = originalConfig }()

	timeout := GetCustomRetryTimeout()
	expectedTimeout := 5 * time.Minute

	if timeout != expectedTimeout {
		t.Errorf("Expected default timeout %v, got %v", expectedTimeout, timeout)
	}
}

func TestUnitGetCustomRetryTimeout_EnvVar(t *testing.T) {
	// Set environment variable
	originalEnv := os.Getenv(customRetryTimeoutEnvVar)
	os.Setenv(customRetryTimeoutEnvVar, "30s")
	defer func() {
		if originalEnv != "" {
			os.Setenv(customRetryTimeoutEnvVar, originalEnv)
		} else {
			os.Unsetenv(customRetryTimeoutEnvVar)
		}
	}()

	// Clear provider meta/config to test env var fallback
	originalMeta := providerMeta
	providerMeta = nil
	defer func() { providerMeta = originalMeta }()

	originalConfig := providerConfig
	providerConfig = nil
	defer func() { providerConfig = originalConfig }()

	timeout := GetCustomRetryTimeout()
	expectedTimeout := 30 * time.Second

	if timeout != expectedTimeout {
		t.Errorf("Expected timeout from env var %v, got %v", expectedTimeout, timeout)
	}
}

func TestUnitGetCustomRetryTimeout_ZeroEnvVar(t *testing.T) {
	// Set environment variable to zero for fail-fast
	originalEnv := os.Getenv(customRetryTimeoutEnvVar)
	os.Setenv(customRetryTimeoutEnvVar, "0s")
	defer func() {
		if originalEnv != "" {
			os.Setenv(customRetryTimeoutEnvVar, originalEnv)
		} else {
			os.Unsetenv(customRetryTimeoutEnvVar)
		}
	}()

	// Clear provider meta/config
	originalMeta := providerMeta
	providerMeta = nil
	defer func() { providerMeta = originalMeta }()

	originalConfig := providerConfig
	providerConfig = nil
	defer func() { providerConfig = originalConfig }()

	timeout := GetCustomRetryTimeout()

	if timeout != 0 {
		t.Errorf("Expected zero timeout for fail-fast, got %v", timeout)
	}
}

func TestUnitGetCustomRetryTimeout_InvalidEnvVar(t *testing.T) {
	// Set invalid environment variable - should fall back to default
	originalEnv := os.Getenv(customRetryTimeoutEnvVar)
	os.Setenv(customRetryTimeoutEnvVar, "invalid")
	defer func() {
		if originalEnv != "" {
			os.Setenv(customRetryTimeoutEnvVar, originalEnv)
		} else {
			os.Unsetenv(customRetryTimeoutEnvVar)
		}
	}()

	// Clear provider meta/config
	originalMeta := providerMeta
	providerMeta = nil
	defer func() { providerMeta = originalMeta }()

	originalConfig := providerConfig
	providerConfig = nil
	defer func() { providerConfig = originalConfig }()

	timeout := GetCustomRetryTimeout()
	expectedTimeout := 5 * time.Minute

	if timeout != expectedTimeout {
		t.Errorf("Expected default timeout %v for invalid env var, got %v", expectedTimeout, timeout)
	}
}

func TestUnitGetCustomRetryTimeout_UsesProviderMetaValue(t *testing.T) {
	originalEnv := os.Getenv(customRetryTimeoutEnvVar)
	os.Setenv(customRetryTimeoutEnvVar, "30s")
	defer func() {
		if originalEnv != "" {
			os.Setenv(customRetryTimeoutEnvVar, originalEnv)
		} else {
			os.Unsetenv(customRetryTimeoutEnvVar)
		}
	}()

	originalMeta := providerMeta
	setProviderMeta(&ProviderMeta{CustomRetryTimeout: 1 * time.Second})
	defer setProviderMeta(originalMeta)

	if got := GetCustomRetryTimeout(); got != 1*time.Second {
		t.Fatalf("expected provider meta timeout 1s, got %v", got)
	}
}

func TestUnitValidateLogFilePath(t *testing.T) {
	testCases := []struct {
		name        string
		path        interface{}
		expectError bool
	}{
		{
			name:        "Valid log file path",
			path:        "logs/application.log",
			expectError: false,
		},
		{
			name:        "Empty path",
			path:        "",
			expectError: true,
		},
		{
			name:        "Non-string value",
			path:        123,
			expectError: true,
		},
		{
			name:        "Relative path with directory",
			path:        "./logs/currentTestCase.log",
			expectError: false,
		},
		{
			name:        "Absolute path",
			path:        "/var/logs/currentTestCase.log",
			expectError: false,
		},
		{
			name:        "Path with spaces",
			path:        "logs/current TestCase.log",
			expectError: true,
		},
		{
			name:        "Incorrect file extension (.tfstate)",
			path:        "terraform.tfstate",
			expectError: true,
		},
		{
			name:        "Incorrect file extension (.go)",
			path:        "main.go",
			expectError: true,
		},
	}

	for _, currentTestCase := range testCases {
		t.Run(currentTestCase.name, func(t *testing.T) {
			diagErr := validateLogFilePath(currentTestCase.path, nil)
			if currentTestCase.expectError && diagErr == nil {
				t.Fatalf("Expected an error, but got none")
			}
			if !currentTestCase.expectError && diagErr != nil {
				t.Fatalf("Unexpected error: %v", diagErr)
			}
		})
	}
}
