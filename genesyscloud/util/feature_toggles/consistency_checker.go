package feature_toggles

import "os"

const consistencyCheckerEnvToggle = "BYPASS_CONSISTENCY_CHECKER"

func CCToggleName() string {
	return consistencyCheckerEnvToggle
}

func CCToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(consistencyCheckerEnvToggle)
	return exists
}
