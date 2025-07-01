package feature_toggles

import "os"

const enableBCPMode = "BCP_MODE_ENABLED"

func BcpModeEnabledName() string {
	return enableBCPMode
}

func BcpModeEnabledExists() bool {
	_, enabled := os.LookupEnv(enableBCPMode)
	return enabled
}
