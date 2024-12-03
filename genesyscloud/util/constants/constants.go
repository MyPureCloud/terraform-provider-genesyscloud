package constants

import (
	"log"
	"os"
	"strconv"
)

const (
	DefaultOutboundScriptName = "Default Outbound Script"
	DefaultInboundScriptName  = "Default Inbound Script"
	DefaultCallbackScriptName = "Default Callback Script"
)

// ConsistencyChecks will return the number of times the consistency checker should retry.
// The user can specify this by setting the env variable CONSISTENCY_CHECKS
// This env variable will only be used when BYPASS_CONSISTENCY_CHECKER is set
func ConsistencyChecks() int {
	defaultChecks := 5

	if checks, exists := os.LookupEnv("CONSISTENCY_CHECKS"); exists {
		if value, err := strconv.Atoi(checks); err == nil {
			return value
		}

		log.Printf("CONSISTENCY_CHECKS set to invalid value, using default value %d", defaultChecks)
		return defaultChecks
	}

	return defaultChecks
}
