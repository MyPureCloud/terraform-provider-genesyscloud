package environment

import "os"

const (
	defaultLocalStackPort    = "4566"
	LocalStackImageUriEnvVar = "LOCAL_STACK_IMAGE_URI" // Set in jenkins
	localStackPortEnvVar     = "LOCALSTACK_PORT"
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
