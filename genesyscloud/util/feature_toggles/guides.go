package feature_toggles

import "os"

const guideEnvToggle = "ENABLE_GUIDE_RESOURCE"

func GuideToggleName() string {
	return guideEnvToggle
}

func GuideToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(guideEnvToggle)
	return exists
}
