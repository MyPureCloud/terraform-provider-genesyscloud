package localstack

import (
	"context"
	"os"
	"testing"

	localStackEnv "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/aws/localstack/environment"
)

func TestUnitNewLocalStackManager(t *testing.T) {
	tests := []struct {
		name          string
		useLocalStack string
		expectError   bool
	}{
		{
			name:          "returns error when USE_LOCAL_STACK is not set",
			useLocalStack: "",
			expectError:   true,
		},
		{
			name:          "returns error when USE_LOCAL_STACK is 'false'",
			useLocalStack: "false",
			expectError:   true,
		},
		{
			name:          "succeeds when USE_LOCAL_STACK is 'true'",
			useLocalStack: "true",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment after test
			defer os.Unsetenv(localStackEnv.UseLocalStackEnvVar)

			if tt.useLocalStack != "" {
				os.Setenv(localStackEnv.UseLocalStackEnvVar, tt.useLocalStack)
			} else {
				os.Unsetenv(localStackEnv.UseLocalStackEnvVar)
			}

			manager, err := NewLocalStackManager(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if manager != nil {
					t.Error("Expected nil manager but got one")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if manager == nil {
					t.Fatalf("Expected manager but got nil")
				}
				if manager.endpoint != "http://localhost:"+localStackEnv.GetLocalStackPort() {
					t.Errorf("Expected endpoint %s but got %s", "http://localhost:"+localStackEnv.GetLocalStackPort(), manager.endpoint)
				}
			}
		})
	}
}
