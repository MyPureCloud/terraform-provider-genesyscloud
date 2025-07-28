package localstack

import (
	"context"
	"fmt"

	localStackEnv "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack/environment"
)

// LocalStackManager performs S3 operations using LocalStack (for testing)
type LocalStackManager struct {
	ctx      context.Context
	endpoint string
}

func NewLocalStackManager(ctx context.Context) (*LocalStackManager, error) {
	if !localStackEnv.LocalStackIsActive() {
		return nil, fmt.Errorf("cannot initiate local stack manager because %s is not set", localStackEnv.UseLocalStackEnvVar)
	}

	return &LocalStackManager{
		ctx:      ctx,
		endpoint: "http://localhost:" + localStackEnv.GetLocalStackPort(),
	}, nil
}
