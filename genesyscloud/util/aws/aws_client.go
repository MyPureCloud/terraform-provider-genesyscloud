package aws

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack"
)

type AWSS3Client struct {
	client *s3.Client
}

func NewAWSS3Client(cfg aws.Config) *AWSS3Client {
	return &AWSS3Client{
		client: s3.NewFromConfig(cfg, func(o *s3.Options) {
			if shouldUseLocalStack() {
				log.Println("Using localstack port: ", localstack.GetLocalStackPort())
				o.BaseEndpoint = aws.String(fmt.Sprintf("http://localhost:%s", localstack.GetLocalStackPort()))
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

const UseLocalStackEnvVar = "USE_LOCAL_STACK"

// shouldUseLocalStack checks if the localstack should be used
func shouldUseLocalStack() bool {
	v, ok := os.LookupEnv(UseLocalStackEnvVar)
	if !ok {
		return false
	}
	return v == "true"
}
