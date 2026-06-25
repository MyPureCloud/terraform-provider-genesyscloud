package feature_toggles

import "os"

const skipInactiveUserExportEnvToggle = "GENESYSCLOUD_SKIP_INACTIVE_USER_EXPORT"

// SkipInactiveUserExportToggleName returns the name of the environment variable that controls the skip inactive user export feature toggle.
func SkipInactiveUserExportToggleName() string {
	return skipInactiveUserExportEnvToggle
}

// SkipInactiveUserExportToggleExists returns true if the skip inactive user export feature toggle is enabled.
func SkipInactiveUserExportToggleExists() bool {
	_, exists := os.LookupEnv(skipInactiveUserExportEnvToggle)
	return exists
}
