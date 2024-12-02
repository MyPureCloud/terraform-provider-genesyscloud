package constants

import (
	"os"
	"strconv"
)

const (
	DefaultOutboundScriptName = "Default Outbound Script"
	DefaultInboundScriptName  = "Default Inbound Script"
	DefaultCallbackScriptName = "Default Callback Script"
)

// ConsistencyChecks will return the number of times the consistency checker should retry.
// The use can specify this by setting the env variable CONSISTENCY_CHECKS
func ConsistencyChecks() int {
	defaultChecks := 5

	if checks, exists := os.LookupEnv("CONSISTENCY_CHECKS"); exists {
		if value, err := strconv.Atoi(checks); err == nil {
			return value
		}

		return defaultChecks
	}

	return defaultChecks
}
