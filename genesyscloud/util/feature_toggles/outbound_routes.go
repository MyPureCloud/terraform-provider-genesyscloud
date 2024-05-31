package feature_toggles

import "os"

const outboundRotesEnvToggle = "ENABLE_STANDALONE_OUTBOUND_ROUTES"

func OutboundRoutesToggleName() string {
	return outboundRotesEnvToggle
}

func OutboundRoutesToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(outboundRotesEnvToggle)
	return exists
}
