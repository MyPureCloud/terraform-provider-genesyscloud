package files

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
)

// S3ClientConfig holds configuration for S3 client creation
type S3ClientConfig struct {
	S3Client S3Client
}

// Common interface for S3 operations
type S3Client interface {
	GetObject(ctx context.Context, bucket, key string) (io.Reader, error)
	PutObject(ctx context.Context, bucket, key string, reader io.Reader) error
	DeleteObject(ctx context.Context, bucket, key string) error
}

func NewS3ClientConfig() *S3ClientConfig {
	return &S3ClientConfig{}
}

func (c *S3ClientConfig) WithS3Client(client S3Client) *S3ClientConfig {
	c.S3Client = client
	return c
}

// DownloadFile downloads a file from S3 and returns a reader
func DownloadFile(ctx context.Context, bucket, key string) (io.Reader, error) {
	return DownloadFileWithConfig(ctx, bucket, key, nil)
}

// DownloadFileWithConfig downloads a file from S3 using the provided configuration
func DownloadFileWithConfig(ctx context.Context, bucket, key string, s3Config *S3ClientConfig) (io.Reader, error) {
	log.Printf("Downloading S3 file: s3://%s/%s", bucket, key)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	defaultClient := NewAWSS3Client(cfg)
	return defaultClient.GetObject(ctx, bucket, key)
}

// IsS3Path checks if the given path is an S3 URI
func IsS3Path(path string) bool {
	return strings.HasPrefix(path, "s3://") || strings.HasPrefix(path, "s3a://")
}

// ParseS3URI parses an S3 URI and returns bucket and key
func ParseS3URI(uri string) (bucket, key string, err error) {
	if !IsS3Path(uri) {
		return "", "", fmt.Errorf("not a valid S3 URI: %s", uri)
	}

	// Remove the s3:// or s3a:// prefix
	cleanURI := strings.TrimPrefix(strings.TrimPrefix(uri, "s3://"), "s3a://")

	// Split by first slash to separate bucket from key
	parts := strings.SplitN(cleanURI, "/", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid S3 URI format: %s", uri)
	}

	bucket = parts[0]
	key = parts[1]

	if bucket == "" {
		return "", "", fmt.Errorf("empty bucket name in S3 URI: %s", uri)
	}

	return bucket, key, nil
}

// GetS3FileReader is a variable that holds the function for getting a file reader from S3 or local filesystem.
var GetS3FileReader = getS3FileReader

// getS3FileReader returns a reader for a file from S3 or local filesystem
func getS3FileReader(ctx context.Context, path string) (io.Reader, *os.File, error) {
	return GetS3FileReaderWithConfig(ctx, path, nil)
}

// GetS3FileReaderWithConfig returns a reader for a file from S3 or local filesystem using the provided configuration
func GetS3FileReaderWithConfig(ctx context.Context, path string, s3Config *S3ClientConfig) (io.Reader, *os.File, error) {
	if IsS3Path(path) {
		bucket, key, err := ParseS3URI(path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse S3 URI: %w", err)
		}

		reader, err := DownloadFileWithConfig(ctx, bucket, key, s3Config)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to download S3 file: %w", err)
		}

		return reader, nil, nil
	}

	// Fall back to local file handling
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open local file: %w", err)
	}

	return file, file, nil
}
