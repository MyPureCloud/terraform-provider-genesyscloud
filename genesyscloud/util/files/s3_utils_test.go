package files

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestIsS3Path(t *testing.T) {
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

func TestParseS3URI(t *testing.T) {
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
			name:        "Not an S3 URI",
			uri:         "http://example.com/file.yaml",
			expectError: true,
		},
		{
			name:        "Empty URI",
			uri:         "",
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

func TestGetS3FileReader_LocalFile(t *testing.T) {
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

func TestGetS3FileReader_NonExistentLocalFile(t *testing.T) {
	ctx := context.Background()
	_, _, err := GetS3FileReader(ctx, "/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestHashS3FileContent_LocalFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write some test content
	testContent := "test content for hashing"
	_, err = tempFile.WriteString(testContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test hashing local file
	ctx := context.Background()
	hash, err := HashS3FileContent(ctx, tempFile.Name())
	if err != nil {
		t.Fatalf("HashS3FileContent failed: %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Hash should be consistent for same content
	hash2, err := HashS3FileContent(ctx, tempFile.Name())
	if err != nil {
		t.Fatalf("HashS3FileContent failed on second call: %v", err)
	}

	if hash != hash2 {
		t.Errorf("Expected consistent hash, got %q and %q", hash, hash2)
	}
}

func TestCopyS3FileToLocal_InvalidS3Path(t *testing.T) {
	ctx := context.Background()
	err := CopyS3FileToLocal(ctx, "/local/path/file.txt", "/tmp/test.txt")
	if err == nil {
		t.Error("Expected error for non-S3 path")
	}
}

func TestValidateS3Path_InvalidPath(t *testing.T) {
	ctx := context.Background()
	err := ValidateS3Path(ctx, "/local/path/file.txt")
	if err == nil {
		t.Error("Expected error for non-S3 path")
	}
}

// Mock S3 client for testing
type mockS3Client struct {
	shouldError bool
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock S3 error")
	}

	// Return a mock response with test content
	content := "test S3 content"
	reader := strings.NewReader(content)

	return &s3.GetObjectOutput{
		Body: io.NopCloser(reader),
	}, nil
}

func (m *mockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock S3 error")
	}

	return &s3.HeadObjectOutput{}, nil
}

// Test helper function to create a mock S3 client
func createMockS3Client(shouldError bool) *mockS3Client {
	return &mockS3Client{shouldError: shouldError}
}

// Note: These tests would require more sophisticated mocking of the AWS SDK
// In a real implementation, you would use a proper mocking framework like testify/mock
// or create interfaces for the S3 client to make it more testable
