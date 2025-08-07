package aws

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	localStackEnv "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack/environment"
)

type AWSS3Client struct {
	client *s3.Client
}

func NewAWSS3Client(cfg aws.Config) *AWSS3Client {
	return &AWSS3Client{
		client: s3.NewFromConfig(cfg, func(o *s3.Options) {
			if localStackEnv.LocalStackIsActive() {
				log.Println("Using localstack port: ", localStackEnv.GetLocalStackPort())
				o.BaseEndpoint = aws.String(fmt.Sprintf("http://localhost:%s", localStackEnv.GetLocalStackPort()))
				region := os.Getenv("GENESYSCLOUD_REGION")
				if region == "" {
					region = "us-east-1" // default
				}
				o.Region = region
			}
			o.UsePathStyle = true
		}),
	}
}

func (a *AWSS3Client) Client() *s3.Client {
	return a.client
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
