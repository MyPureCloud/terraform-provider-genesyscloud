package feature_toggles

import "os"

// envVarIsSet will look up the env var specified by the name param and return true if it exists, false if it does not.
func envVarIsSet(name string) bool {
	var exists bool
	_, exists = os.LookupEnv(name)
	return exists
}
