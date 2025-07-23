# AWS Utilities for Testing

This directory contains utilities for testing AWS S3 integration with LocalStack.

## Docker-based LocalStack Testing

The `docker_utils.go` file provides utilities for managing LocalStack Docker containers for testing S3 integration.

### Prerequisites

1. **Docker**: Must be installed and running on your system
2. **AWS CLI**: Must be installed and configured (for S3 operations)
3. **curl**: Used for health checks

### Usage

The `LocalStackManager` provides a simple interface for managing LocalStack containers:

```go
// Create a LocalStack manager
manager, err := NewLocalStackManager()
if err != nil {
    log.Fatal(err)
}
defer manager.Close()

// Start LocalStack
err = manager.StartLocalStack()
if err != nil {
    log.Fatal(err)
}

// Setup S3 bucket and upload test file
err = manager.SetupS3Bucket("testbucket", "/path/to/file.yml", "flow.yml")
if err != nil {
    log.Fatal(err)
}

// Cleanup
defer func() {
    manager.CleanupS3Bucket("testbucket")
    manager.StopLocalStack()
}()
```

### Environment Variables

Set the `LOCALSTACK_ENDPOINT` environment variable to point to your LocalStack instance:

```bash
export LOCALSTACK_ENDPOINT=http://localhost:4566
```

### Testing

Run the Docker utilities test:

```bash
go test ./genesyscloud/util/aws -v -run TestLocalStackManager
```

This test will:
1. Start a LocalStack container
2. Create an S3 bucket
3. Upload a test file
4. Clean up the bucket
5. Stop the container

### Troubleshooting

1. **Docker not running**: Ensure Docker daemon is running
2. **Port conflicts**: LocalStack uses port 4566 by default
3. **AWS CLI not found**: Install AWS CLI and ensure it's in your PATH
4. **Container already exists**: The manager will automatically remove existing containers

### Integration with Tests

The `TestAccResourceArchFlowWithLocalStack` test demonstrates how to use the LocalStack manager in acceptance tests. The test:

1. Creates a LocalStack manager
2. Starts LocalStack container
3. Sets up S3 bucket with test data
4. Runs the actual test
5. Cleans up all resources

This approach ensures that tests are isolated and don't depend on external services or manual setup. 