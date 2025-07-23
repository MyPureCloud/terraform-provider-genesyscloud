package files

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	LocalStackContainerName = "terraform-provider-genesyscloud-localstack"
	LocalStackImage         = "localstack/localstack:latest"
	LocalStackPort          = "4566"
)

// LocalStackManager manages a LocalStack Docker container for testing
type LocalStackManager struct{}

// NewLocalStackManager creates a new LocalStack manager
func NewLocalStackManager() (*LocalStackManager, error) {
	return &LocalStackManager{}, nil
}

// StartLocalStack starts a LocalStack container using docker commands
func (l *LocalStackManager) StartLocalStack() error {
	// Check if container already exists and remove it
	cmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", LocalStackContainerName), "--format", "{{.ID}}")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) != "" {
		containerID := strings.TrimSpace(string(output))
		log.Printf("Removing existing container: %s", containerID)
		removeCmd := exec.Command("docker", "rm", "-f", containerID)
		if err := removeCmd.Run(); err != nil {
			log.Printf("Warning: failed to remove existing container: %v", err)
		}
	}

	// Pull the LocalStack image
	log.Printf("Pulling LocalStack image: %s", LocalStackImage)
	pullCmd := exec.Command("docker", "pull", LocalStackImage)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull image: %v", err)
	}

	// Start container
	log.Printf("Starting LocalStack container")
	startCmd := exec.Command("docker", "run", "-d",
		"--name", LocalStackContainerName,
		"-p", fmt.Sprintf("%s:%s", LocalStackPort, LocalStackPort),
		"-e", "SERVICES=s3",
		"-e", "DEBUG=1",
		LocalStackImage)

	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr
	if err := startCmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	// Wait for LocalStack to be ready
	log.Printf("Waiting for LocalStack to be ready...")
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)

		// Test if LocalStack is responding
		healthCmd := exec.Command("curl", "-f", GetLocalStackEndpoint()+"/_localstack/health")
		if healthCmd.Run() == nil {
			log.Printf("LocalStack is ready!")
			return nil
		}
	}

	return fmt.Errorf("LocalStack failed to start within 60 seconds")
}

// StopLocalStack stops and removes the LocalStack container
func (l *LocalStackManager) StopLocalStack() error {
	// Stop container
	stopCmd := exec.Command("docker", "stop", LocalStackContainerName)
	if err := stopCmd.Run(); err != nil {
		log.Printf("Warning: failed to stop container: %v", err)
	}

	// Remove container
	removeCmd := exec.Command("docker", "rm", LocalStackContainerName)
	if err := removeCmd.Run(); err != nil {
		log.Printf("Warning: failed to remove container: %v", err)
	}

	return nil
}

// SetupS3Bucket creates an S3 bucket and uploads test data
func (l *LocalStackManager) SetupS3Bucket(bucketName, filePath, objectKey string) error {
	// Wait a moment for LocalStack to be fully ready
	time.Sleep(5 * time.Second)

	// Create bucket
	createBucketCmd := exec.Command("aws", "s3api", "create-bucket",
		"--bucket", bucketName,
		"--region", "us-east-1",
		"--endpoint-url", GetLocalStackEndpoint())

	output, err := createBucketCmd.CombinedOutput()
	if err != nil {
		// Check if bucket already exists
		if !strings.Contains(string(output), "BucketAlreadyOwnedByYou") {
			return fmt.Errorf("failed to create bucket: %v, output: %s", err, string(output))
		}
		log.Printf("Bucket %s already exists", bucketName)
	} else {
		log.Printf("Created bucket: %s", bucketName)
	}

	// Upload file
	uploadCmd := exec.Command("aws", "s3", "cp",
		filePath,
		fmt.Sprintf("s3://%s/%s", bucketName, objectKey),
		"--endpoint-url", GetLocalStackEndpoint())

	output, err = uploadCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to upload file: %v, output: %s", err, string(output))
	}

	log.Printf("Uploaded file %s to s3://%s/%s", filePath, bucketName, objectKey)
	return nil
}

// CleanupS3Bucket removes the S3 bucket and its contents
func (l *LocalStackManager) CleanupS3Bucket(bucketName string) error {
	// Remove all objects in bucket
	removeObjectsCmd := exec.Command("aws", "s3", "rm",
		fmt.Sprintf("s3://%s", bucketName),
		"--recursive",
		"--endpoint-url", GetLocalStackEndpoint())

	output, err := removeObjectsCmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: failed to remove objects from bucket: %v, output: %s", err, string(output))
	}

	// Remove bucket
	removeBucketCmd := exec.Command("aws", "s3api", "delete-bucket",
		"--bucket", bucketName,
		"--endpoint-url", GetLocalStackEndpoint())

	output, err = removeBucketCmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: failed to delete bucket: %v, output: %s", err, string(output))
	}

	return nil
}

// Close closes the LocalStack manager (no-op for shell-based approach)
func (l *LocalStackManager) Close() error {
	return nil
}

// SkipIfLocalStackUnavailable skips the test if environment is not set up for localstack
func SkipIfLocalStackUnavailable(t *testing.T) {
	// Skip if Docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available, skipping test")
	}

	// Skip if AWS CLI is not available
	if _, err := exec.LookPath("aws"); err != nil {
		t.Skip("AWS CLI not available, skipping test")
	}
}
