package feature_toggles

import "os"

const conditionalGroupActivationEnvToggle = "ENABLE_STANDALONE_CGA"

func CGAToggleName() string {
	return conditionalGroupActivationEnvToggle
}

func CGAToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(conditionalGroupActivationEnvToggle)
	return exists
}
