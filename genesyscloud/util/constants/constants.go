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

	DefaultOutboundScriptID = "476c2b71-7429-11e4-9a5b-3f91746bffa3"
	DefaultCallbackScriptID = "ffde0662-8395-9b04-7dcb-b90172109065"
	DefaultInboundScriptID  = "766f1221-047a-11e5-bba2-db8c0964d007"
)

// DefaultScriptMap can be used to get a script name by its ID
var DefaultScriptMap = map[string]string{
	DefaultCallbackScriptID: DefaultCallbackScriptName,
	DefaultInboundScriptID:  DefaultInboundScriptName,
	DefaultOutboundScriptID: DefaultOutboundScriptName,
}

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

type CRUDOperation int

const (
	Create CRUDOperation = iota
	Read
	Update
	Delete
)

func (o CRUDOperation) String() string {
	switch o {
	case Create:
		return "Create"
	case Read:
		return "Read"
	case Update:
		return "Update"
	case Delete:
		return "Delete"
	default:
		return "Unknown"
	}
}
