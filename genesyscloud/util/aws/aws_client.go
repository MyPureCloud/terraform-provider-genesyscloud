package files

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSS3Client struct {
	client *s3.Client
}

func NewAWSS3Client(cfg aws.Config) *AWSS3Client {
	return &AWSS3Client{
		client: s3.NewFromConfig(cfg, func(o *s3.Options) {
			if IsLocalStackEndpointSet() {
				log.Println("Using localstack endpoint: ", GetLocalStackEndpoint())
				o.BaseEndpoint = aws.String(GetLocalStackEndpoint())
			}
			o.UsePathStyle = true
		}),
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
