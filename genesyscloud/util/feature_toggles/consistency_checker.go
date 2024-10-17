package feature_toggles

import "os"

const bypassConsistencyCheckerEnvToggle = "BYPASS_CONSISTENCY_CHECKER"
const disableConsistencyCheckerEnvToggle = "DISABLE_CONSISTENCY_CHECKER"

func BypassCCToggleName() string {
	return bypassConsistencyCheckerEnvToggle
}

func BypassCCToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(bypassConsistencyCheckerEnvToggle)
	return exists
}

func DisableCCToggleName() string {
	return disableConsistencyCheckerEnvToggle
}

func DisableCCToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(disableConsistencyCheckerEnvToggle)
	return exists
}
