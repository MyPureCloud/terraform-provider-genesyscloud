package feature_toggles

import "os"

const ArchyExportToggle = "ENABLE_ARCHY_EXPORT"

func ArchyExportToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(ArchyExportToggle)
	return exists
}
