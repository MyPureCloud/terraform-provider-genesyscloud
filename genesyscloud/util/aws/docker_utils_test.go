package files

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestLocalStackManager(t *testing.T) {
	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available, skipping test")
	}

	// Skip if AWS CLI is not available
	if _, err := exec.LookPath("aws"); err != nil {
		t.Skip("AWS CLI not available, skipping test")
	}

	// Create LocalStack manager
	manager, err := NewLocalStackManager()
	if err != nil {
		t.Fatalf("Failed to create LocalStack manager: %v", err)
	}
	defer manager.Close()

	// Test starting LocalStack
	t.Log("Testing LocalStack start...")
	err = manager.StartLocalStack()
	if err != nil {
		t.Fatalf("Failed to start LocalStack: %v", err)
	}

	// Wait a moment for LocalStack to be fully ready
	time.Sleep(5 * time.Second)

	// Test S3 bucket operations
	bucketName := "test-bucket"
	objectKey := "test-file.txt"

	// Create a test file
	testContent := "This is a test file for LocalStack S3 testing"
	tempFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(testContent)
	if err != nil {
		t.Fatalf("Failed to write test content: %v", err)
	}
	tempFile.Close()

	// Test bucket setup
	t.Log("Testing S3 bucket setup...")
	err = manager.SetupS3Bucket(bucketName, tempFile.Name(), objectKey)
	if err != nil {
		t.Fatalf("Failed to setup S3 bucket: %v", err)
	}

	// Test bucket cleanup
	t.Log("Testing S3 bucket cleanup...")
	err = manager.CleanupS3Bucket(bucketName)
	if err != nil {
		t.Fatalf("Failed to cleanup S3 bucket: %v", err)
	}

	// Test stopping LocalStack
	t.Log("Testing LocalStack stop...")
	err = manager.StopLocalStack()
	if err != nil {
		t.Fatalf("Failed to stop LocalStack: %v", err)
	}

	t.Log("All LocalStack tests passed!")
}
