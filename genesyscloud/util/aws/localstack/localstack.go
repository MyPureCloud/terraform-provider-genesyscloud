package localstack

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

const (
	defaultLocalStackImage         = "localstack/localstack:latest"
	defaultLocalStackPort          = "4566"
	defaultLocalStackContainerName = "terraform-provider-genesyscloud-localstack"
)

// LocalStackManager manages a LocalStack Docker container for testing
type LocalStackManager struct {
	ctx context.Context

	cfg aws.Config

	dockerClient *client.Client
	awsClient    *ecr.Client
	password     string

	containerName string
	imageURI      string
	port          string
	endpoint      string
}

// NewLocalStackManagerWithConfig creates a new LocalStack manager with the given configuration.
// LocalStack is a cloud service emulator that runs in a single container on your laptop or CI environment.
// It provides a testing environment for AWS services like S3, ECR, and others.
//
// Parameters:
//   - cfg: AWS configuration used for ECR authentication and other AWS operations
//   - containerName: name for the Docker container (if empty, defaults to "terraform-provider-genesyscloud-localstack")
//   - image: Docker image URI for LocalStack (if empty, defaults to "localstack/localstack:latest")
//   - port: port number for the LocalStack service (if empty, defaults to "4566")
//
// Returns:
//   - *LocalStackManager: configured manager instance for controlling LocalStack container
//   - error: any error that occurred during initialization (ECR authentication, Docker client creation, etc.)
//
// The function performs the following initialization steps:
//   - Authenticates with ECR to get credentials for pulling images
//   - Creates a Docker client for container management
//   - Sets default values for empty parameters using environment variables or constants
//   - Constructs the LocalStack endpoint URL
func NewLocalStackManagerWithConfig(cfg aws.Config, containerName, image, port string) (*LocalStackManager, error) {
	ctx := context.Background()

	ecrClient := ecr.NewFromConfig(cfg)
	password, err := getECRLoginPasswork(ctx, ecrClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get ECR login password: %v", err)
	}

	dockerClient, err := newDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %v", err)
	}

	lsm := &LocalStackManager{
		ctx:          ctx,
		cfg:          cfg,
		password:     password,
		dockerClient: dockerClient,
		awsClient:    ecrClient,
	}

	lsm.configureLocalStackSettings(containerName, image, port)

	// Set env variable so that the aws client can use it
	err = setLocalStackPort(lsm.port)
	if err != nil {
		log.Printf("failed to set env variable %s: %v", localStackPortEnvVar, err)
	}

	return lsm, nil
}

func NewLocalStackManager(containerName, image, port string) (*LocalStackManager, error) {

	var lsm LocalStackManager

	lsm.configureLocalStackSettings(containerName, image, port)

	lsm.ctx = context.Background()

	// Set env variable so that the aws client can use it
	err := setLocalStackPort(lsm.port)
	if err != nil {
		log.Printf("failed to set env variable %s: %v", localStackPortEnvVar, err)
	}

	return &lsm, nil
}

// StartLocalStack starts a LocalStack container using docker commands
func (l *LocalStackManager) StartLocalStack() error {
	err := pullImage(l.ctx, l.dockerClient, l.imageURI, l.password)
	if err != nil {
		return fmt.Errorf("failed to pull image: %v", err)
	}

	// Start container
	err = l.startLocalStack()
	if err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	// Wait for LocalStack to be ready
	log.Printf("Waiting for LocalStack to be ready...")
	for range 30 {
		time.Sleep(2 * time.Second)

		// Test if LocalStack is responding using native Go HTTP client
		resp, err := http.Get(l.endpoint + "/_localstack/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			log.Printf("LocalStack is ready!")
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	return errors.New("LocalStack failed to start within 60 seconds")
}

// StopLocalStack stops and removes the LocalStack container
func (l *LocalStackManager) StopLocalStack() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("StopLocalStack: %w", err)
		}
	}()

	err = l.dockerClient.ContainerStop(l.ctx, l.containerName, container.StopOptions{})
	if err != nil {
		log.Printf("Warning: failed to stop container: %v", err)
	}

	err = l.dockerClient.ContainerRemove(l.ctx, l.containerName, container.RemoveOptions{})
	if err != nil {
		return err
	}

	return
}

// SetupS3Bucket creates an S3 bucket and uploads test data
func (l *LocalStackManager) SetupS3Bucket(bucketName, filePath, objectKey string) error {
	// Wait a moment for LocalStack to be fully ready
	time.Sleep(5 * time.Second)

	// Create bucket
	createBucketCmd := exec.Command("aws", "s3api", "create-bucket",
		"--bucket", bucketName,
		"--region", "us-east-1",
		"--endpoint-url", l.endpoint)

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
		"--endpoint-url", l.endpoint)

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
		"--endpoint-url", l.endpoint)

	output, err := removeObjectsCmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: failed to remove objects from bucket: %v, output: %s", err, string(output))
	}

	// Remove bucket
	removeBucketCmd := exec.Command("aws", "s3api", "delete-bucket",
		"--bucket", bucketName,
		"--endpoint-url", l.endpoint)

	output, err = removeBucketCmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: failed to delete bucket: %v, output: %s", err, string(output))
	}

	return nil
}

func (l *LocalStackManager) configureLocalStackSettings(containerName, image, port string) {
	if containerName == "" {
		containerName = defaultLocalStackContainerName
	}

	if image == "" {
		image = defaultLocalStackImage
	}

	if port == "" {
		port = defaultLocalStackPort
	}

	l.containerName = containerName
	l.imageURI = image
	l.port = port
	l.endpoint = fmt.Sprintf("http://localhost:%s", port)
}

func newDockerClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %v", err)
	}
	return cli, nil
}

func pullImage(ctx context.Context, cli *client.Client, img string, password string) error {
	auth := map[string]interface{}{
		"username": "AWS",
		"password": password,
	}

	authData, err := json.Marshal(auth)
	if err != nil {
		return err
	}

	auths := base64.URLEncoding.EncodeToString(authData)
	out, err := cli.ImagePull(
		ctx,
		img,
		image.PullOptions{
			RegistryAuth: auths,
		})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(os.Stdout, out)
	if err != nil {
		return fmt.Errorf("failed to read image logs: %w", err)
	}
	return nil
}

func (l *LocalStackManager) startLocalStack() error {
	// remove container if it exists
	log.Printf("Attempting to stop/remove container %s before starting", l.containerName)
	err := l.StopLocalStack()
	if err != nil {
		log.Printf("failed to remove container: %s", err.Error())
	}

	natPort := nat.Port(l.port + "/tcp")

	log.Printf("creating container %s", l.containerName)
	_, err = l.dockerClient.ContainerCreate(l.ctx, &container.Config{
		Image: l.imageURI,
		Env: []string{
			"SERVICES=s3",
			"DEBUG=1",
		},
		ExposedPorts: nat.PortSet{
			natPort: struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			natPort: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: l.port,
				},
			},
		},
	}, nil, nil, l.containerName)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	log.Printf("starting container %s", l.containerName)
	err = l.dockerClient.ContainerStart(l.ctx, l.containerName, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	return nil
}

// Close closes the LocalStack manager (no-op for shell-based approach)
func (l *LocalStackManager) Close() error {
	return nil
}

func getECRLoginPasswork(ctx context.Context, ecr *ecr.Client) (string, error) {
	if ecr == nil {
		return "", fmt.Errorf("ecr client is nil")
	}

	tkn, err := ecr.GetAuthorizationToken(ctx, nil)
	if err != nil {
		return "", err
	}

	if len(tkn.AuthorizationData) == 0 {
		return "", fmt.Errorf("ecr token is empty")
	}

	if len(tkn.AuthorizationData) > 1 {
		return "", fmt.Errorf("multiple ecr tokens: length: %d", len(tkn.AuthorizationData))
	}
	if tkn.AuthorizationData[0].AuthorizationToken == nil {
		return "", fmt.Errorf("ecr token is nil")
	}

	str := *tkn.AuthorizationData[0].AuthorizationToken
	dec, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		return "", fmt.Errorf("failed to decode ecr token: %w", err)
	}

	spl := strings.Split(string(dec), ":")
	if len(spl) != 2 {
		return "", fmt.Errorf("unexpected ecr token format")
	}

	return spl[1], nil
}

const localStackPortEnvVar = "LOCALSTACK_PORT"

func setLocalStackPort(port string) error {
	return os.Setenv(localStackPortEnvVar, port)
}

func GetLocalStackPort() string {
	if port, ok := os.LookupEnv(localStackPortEnvVar); ok {
		return port
	}
	return defaultLocalStackPort
}
