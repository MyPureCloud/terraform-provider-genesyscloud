package feature_toggles

import "os"

const integrationConfigEnvToggle = "ENABLE_STANDALONE_INTEGRATION_CONFIG"

func ICToggleName() string {
	return integrationConfigEnvToggle
}

func ICToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(integrationConfigEnvToggle)
	return exists
}
