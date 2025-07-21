package files

import (
	"context"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOS3Client struct {
	client *minio.Client
}

func NewMinIOS3Client(endpoint string) (*MinIOS3Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4("Q3AM3UQ867SPQQA43P2F", "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG", ""),
		Secure: true,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOS3Client{client: client}, nil
}

func (m *MinIOS3Client) Client() *minio.Client {
	return m.client
}

func (m *MinIOS3Client) GetObject(ctx context.Context, bucket, key string) (io.Reader, error) {
	log.Printf("Getting object from MinIO: s3://%s/%s", bucket, key)
	obj, err := m.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Error getting object from MinIO: s3://%s/%s: %v", bucket, key, err)
		return nil, err
	}
	return obj, nil
}

func (m *MinIOS3Client) PutObject(ctx context.Context, bucket, key string, reader io.Reader) error {
	_, err := m.client.PutObject(ctx, bucket, key, reader, int64(-1), minio.PutObjectOptions{})
	return err
}

func (m *MinIOS3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	return m.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}
