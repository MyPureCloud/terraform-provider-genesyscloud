package files

import "os"

const (
	defaultLocalStackEndpoint = "http://localhost:4566"
	LocalStackEndpointEnvVar  = "LOCALSTACK_ENDPOINT"
)

// SetLocalStackEndpoint sets the localstack endpoint to the default value
func SetLocalStackEndpoint() error {
	return os.Setenv(LocalStackEndpointEnvVar, defaultLocalStackEndpoint)
}

// SetLocalStackEndpointWithCustomEndpoint sets the localstack endpoint to the custom value
func SetLocalStackEndpointWithCustomEndpoint(endpoint string) error {
	return os.Setenv(LocalStackEndpointEnvVar, endpoint)
}

// UnsetLocalStackEndpoint unsets the localstack endpoint from the environment variable
func UnsetLocalStackEndpoint() error {
	return os.Unsetenv(LocalStackEndpointEnvVar)
}

// GetLocalStackEndpoint gets the localstack endpoint from the environment variable or the default value
func GetLocalStackEndpoint() string {
	if localStackEndpoint, ok := os.LookupEnv(LocalStackEndpointEnvVar); ok && localStackEndpoint != "" {
		return localStackEndpoint
	}
	return defaultLocalStackEndpoint
}

// IsLocalStackEndpointSet checks if the localstack endpoint is set
func IsLocalStackEndpointSet() bool {
	return GetLocalStackEndpoint() != ""
}
