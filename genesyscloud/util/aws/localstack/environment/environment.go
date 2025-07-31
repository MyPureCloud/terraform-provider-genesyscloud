package environment

import "os"

// Environment variables for localstack (set in jenkins)
const (
	defaultLocalStackPort    = "4566"
	localStackPortEnvVar     = "LOCAL_STACK_PORT"
	LocalStackImageUriEnvVar = "LOCAL_STACK_IMAGE_URI"
	UseLocalStackEnvVar      = "USE_LOCAL_STACK"
)

func GetLocalStackPort() string {
	if port, ok := os.LookupEnv(localStackPortEnvVar); ok {
		return port
	}
	return defaultLocalStackPort
}

// LocalStackIsActive checks if the localstack should be used
func LocalStackIsActive() bool {
	v, ok := os.LookupEnv(UseLocalStackEnvVar)
	if !ok {
		return false
	}
	return v == "true"
}
