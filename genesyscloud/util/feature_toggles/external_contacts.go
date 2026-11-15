package feature_toggles

import "os"

const simpleExternalContactBlockLabelEnvToggle = "GENESYSCLOUD_EXTERNAL_CONTACTS_SIMPLE_BLOCK_LABEL"

// SimpleExternalContactBlockLabelToggleName returns the environment variable
// that disables organization names in exported external contact block labels.
func SimpleExternalContactBlockLabelToggleName() string {
	return simpleExternalContactBlockLabelEnvToggle
}

// SimpleExternalContactBlockLabelToggleExists returns true when external contact
// exports should avoid the organization lookup and use a simpler block label.
func SimpleExternalContactBlockLabelToggleExists() bool {
	_, exists := os.LookupEnv(simpleExternalContactBlockLabelEnvToggle)
	return exists
}
