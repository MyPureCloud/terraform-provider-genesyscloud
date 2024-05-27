package feature_toggles

import "os"

const outboundEmailAddressEnvToggle = "ENABLE_STANDALONE_EMAIL_ADDRESS"

func OEAToggleName() string {
	return outboundEmailAddressEnvToggle
}

func OEAToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(outboundEmailAddressEnvToggle)
	return exists
}
