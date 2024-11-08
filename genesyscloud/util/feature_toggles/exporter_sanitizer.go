package feature_toggles

import "os"

const sanitizerLegacy = "GENESYS_SANITIZER_LEGACY"
const sanitizerOptimized = "GENESYS_SANITIZER_OPTIMIZED"

func ExporterSanitizerLegacyName() string {
	return sanitizerLegacy
}

func ExporterSanitizerLegacyToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(sanitizerLegacy)
	return exists
}

func ExporterSanitizerOptimizedName() string {
	return sanitizerOptimized
}

func ExporterSanitizerOptimizedToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(sanitizerOptimized)
	return exists
}
