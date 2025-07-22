package files

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSS3Client struct {
	client *s3.Client
}

func NewAWSS3Client(cfg aws.Config) *AWSS3Client {
	return &AWSS3Client{
		client: s3.NewFromConfig(cfg),
	}
}

func (a *AWSS3Client) GetObject(ctx context.Context, bucket, key string) (io.Reader, error) {
	result, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (a *AWSS3Client) PutObject(ctx context.Context, bucket, key string, reader io.Reader) error {
	_, err := a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	return err
}

func (a *AWSS3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}
