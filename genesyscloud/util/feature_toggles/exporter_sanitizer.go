package feature_toggles

import "os"

const enableBCPOptimizedSanitizer = "ENABLE_BCP_OPTIMIZED_SANITIZER"

func ExporterSanitizerBCPOptimizedName() string {
	return enableBCPOptimizedSanitizer
}

func ExporterSanitizerBCPOptimizedToggleExists() bool {
	_, enabled := os.LookupEnv(enableBCPOptimizedSanitizer)
	return enabled
}
