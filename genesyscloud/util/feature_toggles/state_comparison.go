package feature_toggles

import "os"

const enableStateComparison = "ENABLE_EXPORTER_STATE_COMPARISON"

func StateComparison() string {
	return enableStateComparison
}

func StateComparisonTrue() bool {
	var exists bool
	val, exists := os.LookupEnv(enableStateComparison)
	if exists && val == "true" {
		return true
	} else {
		return false
	}
}
