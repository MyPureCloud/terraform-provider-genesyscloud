package aws

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestUnitIsS3Path(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "S3 path with s3://",
			path:     "s3://my-bucket/path/to/file.yaml",
			expected: true,
		},
		{
			name:     "S3 path with s3a://",
			path:     "s3a://my-bucket/path/to/file.yaml",
			expected: true,
		},
		{
			name:     "Local file path",
			path:     "/path/to/local/file.yaml",
			expected: false,
		},
		{
			name:     "HTTP URL",
			path:     "http://example.com/file.yaml",
			expected: false,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "Relative path",
			path:     "./file.yaml",
			expected: false,
		},
		{
			name:     "HTTPS URL",
			path:     "https://example.com/file.yaml",
			expected: false,
		},
		{
			name:     "File protocol",
			path:     "file:///path/to/file.yaml",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsS3Path(tt.path)
			if result != tt.expected {
				t.Errorf("IsS3Path(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestUnitParseS3URI(t *testing.T) {
	tests := []struct {
		name           string
		uri            string
		expectedBucket string
		expectedKey    string
		expectError    bool
	}{
		{
			name:           "Valid S3 URI with s3://",
			uri:            "s3://my-bucket/path/to/file.yaml",
			expectedBucket: "my-bucket",
			expectedKey:    "path/to/file.yaml",
			expectError:    false,
		},
		{
			name:           "Valid S3 URI with s3a://",
			uri:            "s3a://my-bucket/path/to/file.yaml",
			expectedBucket: "my-bucket",
			expectedKey:    "path/to/file.yaml",
			expectError:    false,
		},
		{
			name:           "S3 URI with nested paths",
			uri:            "s3://my-bucket/folder/subfolder/file.yaml",
			expectedBucket: "my-bucket",
			expectedKey:    "folder/subfolder/file.yaml",
			expectError:    false,
		},
		{
			name:           "S3 URI with special characters in key",
			uri:            "s3://my-bucket/path/with spaces and (parentheses)/file.yaml",
			expectedBucket: "my-bucket",
			expectedKey:    "path/with spaces and (parentheses)/file.yaml",
			expectError:    false,
		},
		{
			name:           "S3 URI with file extension",
			uri:            "s3://my-bucket/data/file.json",
			expectedBucket: "my-bucket",
			expectedKey:    "data/file.json",
			expectError:    false,
		},
		{
			name:        "Invalid S3 URI - missing key",
			uri:         "s3://my-bucket",
			expectError: true,
		},
		{
			name:        "Invalid S3 URI - empty bucket",
			uri:         "s3:///path/to/file.yaml",
			expectError: true,
		},
		{
			name:           "S3 URI with double slash (valid)",
			uri:            "s3://my-bucket//path/to/file.yaml",
			expectedBucket: "my-bucket",
			expectedKey:    "/path/to/file.yaml",
			expectError:    false,
		},
		{
			name:        "Not an S3 URI",
			uri:         "http://example.com/file.yaml",
			expectError: true,
		},
		{
			name:        "Empty URI",
			uri:         "",
			expectError: true,
		},
		{
			name:        "Invalid S3 URI - only protocol",
			uri:         "s3://",
			expectError: true,
		},
		{
			name:        "Invalid S3 URI - s3a with missing key",
			uri:         "s3a://my-bucket",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := ParseS3URI(tt.uri)
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseS3URI(%q) expected error but got none", tt.uri)
				}
			} else {
				if err != nil {
					t.Errorf("ParseS3URI(%q) unexpected error: %v", tt.uri, err)
				}
				if bucket != tt.expectedBucket {
					t.Errorf("ParseS3URI(%q) bucket = %q, expected %q", tt.uri, bucket, tt.expectedBucket)
				}
				if key != tt.expectedKey {
					t.Errorf("ParseS3URI(%q) key = %q, expected %q", tt.uri, key, tt.expectedKey)
				}
			}
		})
	}
}

func TestUnitGetS3FileReader_LocalFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write some test content
	testContent := "test content for local file"
	_, err = tempFile.WriteString(testContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test reading local file
	ctx := context.Background()
	reader, file, err := GetS3FileReader(ctx, tempFile.Name())
	if err != nil {
		t.Fatalf("GetS3FileReader failed: %v", err)
	}

	if file == nil {
		t.Error("Expected file to be returned for local file")
	}
	defer file.Close()

	// Read and verify content
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read content: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content %q, got %q", testContent, string(content))
	}
}

func TestUnitGetS3FileReader_NonExistentLocalFile(t *testing.T) {
	ctx := context.Background()
	_, _, err := GetS3FileReader(ctx, "/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "failed to open local file") {
		t.Errorf("Expected error about opening local file, got: %v", err)
	}
}

func TestUnitGetS3FileReader_InvalidS3Path(t *testing.T) {
	ctx := context.Background()
	_, _, err := GetS3FileReader(ctx, "s3://invalid-uri")
	if err == nil {
		t.Error("Expected error for invalid S3 URI")
	}
	if !strings.Contains(err.Error(), "failed to parse S3 URI") {
		t.Errorf("Expected error about parsing S3 URI, got: %v", err)
	}
}

func TestUnitGetS3FileReader_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Empty path",
			path:        "",
			expectError: true,
			errorMsg:    "failed to open local file",
		},
		{
			name:        "Path with spaces",
			path:        "/path/with spaces/file.txt",
			expectError: true,
			errorMsg:    "failed to open local file",
		},
		{
			name:        "Invalid S3 URI format",
			path:        "s3://",
			expectError: true,
			errorMsg:    "failed to parse S3 URI",
		},
		{
			name:        "S3 URI with missing key",
			path:        "s3://bucket",
			expectError: true,
			errorMsg:    "failed to parse S3 URI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, _, err := GetS3FileReader(ctx, tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for path %q", tt.path)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for path %q: %v", tt.path, err)
				}
			}
		})
	}
}
