package provider

import (
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
