package localstack

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	utilAws "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws"
)

// SetupS3Bucket creates an S3 bucket and uploads test data using AWS SDK
func (l *localStackManager) SetupS3Bucket(bucketName, filePath, objectKey string) error {
	// Wait a moment for LocalStack to be fully ready
	time.Sleep(5 * time.Second)

	// Create S3 client with LocalStack endpoint
	s3Client, err := l.createS3Client()
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %v", err)
	}

	// Create bucket
	err = l.createBucket(s3Client, bucketName)
	if err != nil {
		return fmt.Errorf("failed to create bucket: %v", err)
	}

	// Upload file
	err = l.uploadFile(s3Client, bucketName, filePath, objectKey)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	log.Printf("Uploaded file %s to s3://%s/%s", filePath, bucketName, objectKey)
	return nil
}

// CleanupS3Bucket removes the S3 bucket and its contents using AWS SDK
func (l *localStackManager) CleanupS3Bucket(bucketName string) error {
	// Create S3 client with LocalStack endpoint
	s3Client, err := l.createS3Client()
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %v", err)
	}

	// Remove all objects in bucket
	err = l.deleteAllObjects(s3Client, bucketName)
	if err != nil {
		log.Printf("Warning: failed to remove objects from bucket: %v", err)
	}

	// Remove bucket
	err = l.deleteBucket(s3Client, bucketName)
	if err != nil {
		log.Printf("Warning: failed to delete bucket: %v", err)
	}

	return nil
}

// createS3Client creates an S3 client configured for LocalStack
func (l *localStackManager) createS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	return utilAws.NewAWSS3Client(cfg).Client(), nil
}

// createBucket creates an S3 bucket
func (l *localStackManager) createBucket(s3Client *s3.Client, bucketName string) error {
	_, err := s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// Check if bucket already exists
		if strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") ||
			strings.Contains(err.Error(), "BucketAlreadyExists") {
			log.Printf("Bucket %s already exists", bucketName)
			return nil
		}
		return fmt.Errorf("failed to create bucket: %v", err)
	}

	log.Printf("Created bucket: %s", bucketName)
	return nil
}

// uploadFile uploads a file to S3
func (l *localStackManager) uploadFile(s3Client *s3.Client, bucketName, filePath, objectKey string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	// Upload the file
	_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

// deleteAllObjects deletes all objects in a bucket
func (l *localStackManager) deleteAllObjects(s3Client *s3.Client, bucketName string) error {
	// List all objects in the bucket
	listOutput, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to list objects: %v", err)
	}

	// Delete each object
	for _, object := range listOutput.Contents {
		_, err := s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    object.Key,
		})
		if err != nil {
			log.Printf("Warning: failed to delete object %s: %v", *object.Key, err)
		}
	}

	return nil
}

// deleteBucket deletes an S3 bucket
func (l *localStackManager) deleteBucket(s3Client *s3.Client, bucketName string) error {
	_, err := s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %v", err)
	}

	return nil
}
