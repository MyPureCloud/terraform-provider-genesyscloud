package localstack

import (
	"context"
	"fmt"

	localStackEnv "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack/environment"
)

// localStackManager performs S3 operations using LocalStack (for testing)
type localStackManager struct {
	ctx      context.Context
	endpoint string
}

// NewLocalStackManager creates a new LocalStack manager for S3 operations during testing.
//
// This function is used to initialize a manager that performs S3 operations (such as bucket creation,
// file uploads, and bucket deletion) against a LocalStack instance instead of actual AWS services.
//
// The function requires the USE_LOCAL_STACK environment variable to be set to "true" before it can
// be instantiated. This safety measure ensures that the S3 client will point to the LocalStack
// endpoint (default: http://localhost:4566) instead of AWS services. This allows for safe, isolated
// testing of S3-dependent functionality without affecting real AWS resources.
//
// Returns:
//   - *localStackManager: A configured manager for LocalStack S3 operations
//   - error: An error if the USE_LOCAL_STACK environment variable is not set to "true"
func NewLocalStackManager(ctx context.Context) (*localStackManager, error) {
	if !localStackEnv.LocalStackIsActive() {
		return nil, fmt.Errorf("cannot initiate local stack manager because %s is not set", localStackEnv.UseLocalStackEnvVar)
	}

	return &localStackManager{
		ctx:      ctx,
		endpoint: "http://localhost:" + localStackEnv.GetLocalStackPort(),
	}, nil
}
