package feature_toggles

import "os"

const sanitizerOptimized = "GENESYS_SANITIZER_OPTIMIZED"
const enableOptimizedSanitizer = "ENABLE_OPTIMIZED_SANITIZER"

func ExporterSanitizerOptimizedName() string {
	return enableOptimizedSanitizer
}

func ExporterSanitizerOptimizedToggleExists() bool {
	var exists bool
	_, sanitizerOptimizedExists := os.LookupEnv(sanitizerOptimized)
	_, enableOptimizedSanitizer := os.LookupEnv(enableOptimizedSanitizer)
	if sanitizerOptimizedExists || enableOptimizedSanitizer {
		exists = true
	}
	return exists
}

const enableBCPOptimizedSanitizer = "ENABLE_BCP_OPTIMIZED_SANITIZER"

func ExporterSanitizerBCPOptimizedName() string {
	return enableBCPOptimizedSanitizer
}

func ExporterSanitizerBCPOptimizedToggleExists() bool {
	var exists bool
	_, enableBCPOptimizedSanitizer := os.LookupEnv(enableBCPOptimizedSanitizer)
	if enableBCPOptimizedSanitizer {
		exists = true
	}
	return exists
}
