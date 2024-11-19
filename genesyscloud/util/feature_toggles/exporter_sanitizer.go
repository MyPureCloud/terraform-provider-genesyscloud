package feature_toggles

import "os"

const sanitizerOptimized = "GENESYS_SANITIZER_OPTIMIZED"

func ExporterSanitizerOptimizedName() string {
	return sanitizerOptimized
}

func ExporterSanitizerOptimizedToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(sanitizerOptimized)
	return exists
}
