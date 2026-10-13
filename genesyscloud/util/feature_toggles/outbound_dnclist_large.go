package feature_toggles

import "os"

const outboundDnclistUnmanagedEntriesEnvToggle = "GENESYSCLOUD_OUTBOUND_DNCLIST_UNMANAGED_ENTRIES"

// OutboundDnclistUnmanagedEntriesToggleName returns the name of the environment
// variable that controls how genesyscloud_outbound_dnclist handles the `entries`
// block.
//
// When the variable is set, the provider stops managing `entries` through
// Terraform: it does not export/download the list on read, does not store
// entries in state, and never deletes entries based on an omitted/empty config
// block. This mode is intended for customers with very large DNC lists (where
// managing entries through Terraform causes plan to run out of memory) and who
// maintain their phone numbers through a separate process.
//
// When the variable is NOT set (default), the provider keeps the original
// behavior: entries are fetched into state and reconciled (delete + re-upload)
// whenever they change, including removal when the block is omitted.
func OutboundDnclistUnmanagedEntriesToggleName() string {
	return outboundDnclistUnmanagedEntriesEnvToggle
}

// OutboundDnclistUnmanagedEntriesToggleExists returns true if the unmanaged
// entries feature toggle is enabled.
func OutboundDnclistUnmanagedEntriesToggleExists() bool {
	_, exists := os.LookupEnv(outboundDnclistUnmanagedEntriesEnvToggle)
	return exists
}
