package exporter

import (
	"os"
	"strings"
	"testing"
)

func TestMrmoValidateExportInput(t *testing.T) {
	tests := []struct {
		name        string
		input       ExportInput
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid input with all required fields",
			input: ExportInput{
				ResourceType:        "genesyscloud_flow",
				EntityId:            "test-id-123",
				GenerateOutputFiles: false,
				Directory:           "",
			},
			expectError: false,
		},
		{
			name: "Valid input with GenerateOutputFiles true and Directory set",
			input: ExportInput{
				ResourceType:        "genesyscloud_flow",
				EntityId:            "test-id-123",
				GenerateOutputFiles: true,
				Directory:           "/tmp/test",
			},
			expectError: false,
		},
		{
			name: "Missing ResourceType",
			input: ExportInput{
				ResourceType:        "",
				EntityId:            "test-id-123",
				GenerateOutputFiles: false,
				Directory:           "",
			},
			expectError: true,
			errorMsg:    "'ResourceType' is a required field",
		},
		{
			name: "Missing EntityId",
			input: ExportInput{
				ResourceType:        "genesyscloud_flow",
				EntityId:            "",
				GenerateOutputFiles: false,
				Directory:           "",
			},
			expectError: true,
			errorMsg:    "'EntityId' is a required field",
		},
		{
			name: "GenerateOutputFiles true but Directory empty",
			input: ExportInput{
				ResourceType:        "genesyscloud_flow",
				EntityId:            "test-id-123",
				GenerateOutputFiles: true,
				Directory:           "",
			},
			expectError: true,
			errorMsg:    "'Directory' is a required field when 'GenerateOutputFiles' is set to true",
		},
		{
			name: "Both ResourceType and EntityId missing",
			input: ExportInput{
				ResourceType:        "",
				EntityId:            "",
				GenerateOutputFiles: false,
				Directory:           "",
			},
			expectError: true,
			errorMsg:    "'ResourceType' is a required field", // Should return first error encountered
		},
		{
			name: "All fields missing",
			input: ExportInput{
				ResourceType:        "",
				EntityId:            "",
				GenerateOutputFiles: true,
				Directory:           "",
			},
			expectError: true,
			errorMsg:    "'ResourceType' is a required field", // Should return first error encountered
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExportInput(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestMrmoGenerateDefaults(t *testing.T) {
	tests := []struct {
		name           string
		input          ExportInput
		expectedChange bool
		validateDir    func(string) bool
	}{
		{
			name: "GenerateOutputFiles false and Directory empty - should set default",
			input: ExportInput{
				ResourceType:        "genesyscloud_flow",
				EntityId:            "test-id-123",
				GenerateOutputFiles: false,
				Directory:           "",
			},
			expectedChange: true,
			validateDir: func(dir string) bool {
				// Should be in temp directory and start with "mrmo_"
				return strings.HasPrefix(dir, os.TempDir()) && strings.Contains(dir, "mrmo_")
			},
		},
		{
			name: "GenerateOutputFiles false and Directory set - should not change",
			input: ExportInput{
				ResourceType:        "genesyscloud_flow",
				EntityId:            "test-id-123",
				GenerateOutputFiles: false,
				Directory:           "/existing/path",
			},
			expectedChange: false,
			validateDir: func(dir string) bool {
				return dir == "/existing/path"
			},
		},
		{
			name: "GenerateOutputFiles true and Directory empty - should not change",
			input: ExportInput{
				ResourceType:        "genesyscloud_flow",
				EntityId:            "test-id-123",
				GenerateOutputFiles: true,
				Directory:           "",
			},
			expectedChange: false,
			validateDir: func(dir string) bool {
				return dir == ""
			},
		},
		{
			name: "GenerateOutputFiles true and Directory set - should not change",
			input: ExportInput{
				ResourceType:        "genesyscloud_flow",
				EntityId:            "test-id-123",
				GenerateOutputFiles: true,
				Directory:           "/existing/path",
			},
			expectedChange: false,
			validateDir: func(dir string) bool {
				return dir == "/existing/path"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalDir := tt.input.Directory

			generateDefaults(&tt.input)

			if tt.expectedChange {
				if tt.input.Directory == originalDir {
					t.Errorf("Expected Directory to change but it remained '%s'", originalDir)
				}
			} else {
				if tt.input.Directory != originalDir {
					t.Errorf("Expected Directory to remain '%s' but it changed to '%s'", originalDir, tt.input.Directory)
				}
			}

			if !tt.validateDir(tt.input.Directory) {
				t.Errorf("Directory validation failed for value: '%s'", tt.input.Directory)
			}
		})
	}
}

func TestMrmoGenerateDefaults_UniqueDirectories(t *testing.T) {
	// Test that multiple calls generate unique directory names
	input1 := ExportInput{
		ResourceType:        "genesyscloud_flow",
		EntityId:            "test-id-1",
		GenerateOutputFiles: false,
		Directory:           "",
	}

	input2 := ExportInput{
		ResourceType:        "genesyscloud_flow",
		EntityId:            "test-id-2",
		GenerateOutputFiles: false,
		Directory:           "",
	}

	generateDefaults(&input1)
	generateDefaults(&input2)

	if input1.Directory == input2.Directory {
		t.Errorf("Expected unique directories but got same value: '%s'", input1.Directory)
	}

	// Verify both are valid temp directory paths
	if !strings.HasPrefix(input1.Directory, os.TempDir()) {
		t.Errorf("Expected input1 directory to be in temp dir, got: '%s'", input1.Directory)
	}

	if !strings.HasPrefix(input2.Directory, os.TempDir()) {
		t.Errorf("Expected input2 directory to be in temp dir, got: '%s'", input2.Directory)
	}
}
