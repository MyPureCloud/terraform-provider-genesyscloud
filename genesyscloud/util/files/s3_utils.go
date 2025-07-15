package files

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

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

// getS3Client creates an S3 client using AWS credential chain
func getS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return s3.NewFromConfig(cfg), nil
}

// DownloadS3File downloads a file from S3 and returns a reader
func DownloadS3File(ctx context.Context, bucket, key string) (io.Reader, error) {
	client, err := getS3Client(ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("Downloading S3 file: s3://%s/%s", bucket, key)

	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}

	return result.Body, nil
}

// UploadS3File uploads a file to S3
func UploadS3File(ctx context.Context, bucket, key string, reader io.Reader) error {
	client, err := getS3Client(ctx)
	if err != nil {
		return err
	}

	log.Printf("Uploading file to S3: s3://%s/%s", bucket, key)

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}

// GetS3FileReader returns a reader for a file from S3 or local filesystem
func GetS3FileReader(ctx context.Context, path string) (io.Reader, *os.File, error) {
	if IsS3Path(path) {
		bucket, key, err := ParseS3URI(path)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse S3 URI: %w", err)
		}

		reader, err := DownloadS3File(ctx, bucket, key)
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

// HashS3FileContent calculates SHA256 hash of S3 file content
func HashS3FileContent(ctx context.Context, path string) (string, error) {
	reader, file, err := GetS3FileReader(ctx, path)
	if err != nil {
		return "", fmt.Errorf("failed to get file reader: %w", err)
	}

	if file != nil {
		defer file.Close()
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("failed to hash file content: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CopyS3FileToLocal copies an S3 file to a local temporary file
func CopyS3FileToLocal(ctx context.Context, s3Path, localPath string) error {
	if !IsS3Path(s3Path) {
		return fmt.Errorf("not an S3 path: %s", s3Path)
	}

	reader, _, err := GetS3FileReader(ctx, s3Path)
	if err != nil {
		return fmt.Errorf("failed to get S3 file reader: %w", err)
	}

	// Create local directory if it doesn't exist
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy content from S3 to local file
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to copy S3 file to local: %w", err)
	}

	log.Printf("Successfully copied S3 file %s to local path %s", s3Path, localPath)
	return nil
}

// ValidateS3Path validates if an S3 path is accessible
func ValidateS3Path(ctx context.Context, path string) error {
	if !IsS3Path(path) {
		return fmt.Errorf("not a valid S3 path: %s", path)
	}

	bucket, key, err := ParseS3URI(path)
	if err != nil {
		return fmt.Errorf("failed to parse S3 URI: %w", err)
	}

	client, err := getS3Client(ctx)
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Check if object exists
	_, err = client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("S3 object does not exist or is not accessible: %w", err)
	}

	return nil
}
