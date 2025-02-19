package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"testing"
)

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

func TestUnitDetermineTokenPoolSize(t *testing.T) {
	// Save original env var value and restore after test
	originalEnvVar := os.Getenv(tokenPoolSizeEnvVar)
	defer func(key, value string) {
		err := os.Setenv(key, value)
		if err != nil {
			t.Logf("Failed to restore env var %s: %s", tokenPoolSizeEnvVar, err.Error())
		}
	}(tokenPoolSizeEnvVar, originalEnvVar)

	tests := []struct {
		name       string
		envVar     string
		mockData   map[string]interface{}
		wantResult int
	}{
		{
			name:       "resource_data_value",
			envVar:     "15",
			mockData:   map[string]interface{}{"token_pool_size": 10},
			wantResult: 10,
		},
		{
			name:       "env_var_value",
			envVar:     "20",
			mockData:   map[string]interface{}{},
			wantResult: 20,
		},
		{
			name:       "default_value_when_no_env_var",
			envVar:     "",
			mockData:   map[string]interface{}{},
			wantResult: int(tokenPoolSizeDefault),
		},
		{
			name:       "default_value_when_invalid_env_var",
			envVar:     "invalid",
			mockData:   map[string]interface{}{},
			wantResult: int(tokenPoolSizeDefault),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envVar != "" {
				_ = os.Setenv(tokenPoolSizeEnvVar, tt.envVar)
			} else {
				_ = os.Unsetenv(tokenPoolSizeEnvVar)
			}

			// Create mock ResourceData
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"token_pool_size": {
					Type:     schema.TypeInt,
					Optional: true,
				},
			}, tt.mockData)

			result := determineTokenPoolSize(d)

			if result != tt.wantResult {
				t.Errorf("determineTokenPoolSize() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}
