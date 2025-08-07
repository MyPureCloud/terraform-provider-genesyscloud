package aws

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
)

// DownloadFile downloads a file from S3 and returns a reader
func DownloadFile(ctx context.Context, bucket, key string) (io.Reader, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return NewAWSS3Client(cfg).GetObject(ctx, bucket, key)
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
	if IsS3Path(path) {
		bucket, key, err := ParseS3URI(path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse S3 URI: %w", err)
		}

		reader, err := DownloadFile(ctx, bucket, key)
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
