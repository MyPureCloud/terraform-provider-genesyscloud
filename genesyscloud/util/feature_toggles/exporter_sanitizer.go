package feature_toggles

import "os"

const sanitizerLegacy = "GENESYS_SANITIZER_LEGACY"
const sanitizerTimeOptimized = "GENESYS_SANITIZER_TIME_OPTIMIZED"

func ExporterSanitizerLegacyName() string {
	return sanitizerLegacy
}

func ExporterSanitizerLegacyToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(sanitizerLegacy)
	return exists
}

func ExporterSanitizerTimeOptimizedName() string {
	return sanitizerTimeOptimized
}

func ExporterSanitizerTimeOptimizedToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(sanitizerTimeOptimized)
	return exists
}
