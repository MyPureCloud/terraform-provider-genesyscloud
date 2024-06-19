package feature_toggles

import "os"

const outboundRoutesEnvToggle = "ENABLE_STANDALONE_OUTBOUND_ROUTES"

func OutboundRoutesToggleName() string {
	return outboundRoutesEnvToggle
}

func OutboundRoutesToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(outboundRoutesEnvToggle)
	return exists
}
