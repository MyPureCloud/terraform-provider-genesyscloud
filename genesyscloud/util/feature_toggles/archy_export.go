package feature_toggles

import (
	"log"
	"os"
)

const ArchyExportToggle = "ENABLE_ARCHY_EXPORT"

var usingLegacyExport = true

func SetUsingLegacyExport(val bool) {
	log.Println("Setting usingLegacyExport")
	usingLegacyExport = val
}

func GetUsingLegacyExport() bool {
	return usingLegacyExport
}

func ArchyExportToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(ArchyExportToggle)
	return exists
}
