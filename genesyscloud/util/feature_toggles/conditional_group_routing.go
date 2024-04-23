package feature_toggles

import "os"

const conditionalGroupRoutingEnvToggle = "ENABLE_STANDALONE_CGR"

func CSGToggleName() string {
	return conditionalGroupRoutingEnvToggle
}

func CSGToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(conditionalGroupRoutingEnvToggle)
	return exists
}
