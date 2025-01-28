package panic_recovery_logger

import (
	"fmt"
	"testing"
)

func TestUnitGetPanicRecoveryLoggerInstance(t *testing.T) {
	// restore after all
	panicRecoverLoggerCopy := panicRecoverLogger
	defer func() {
		panicRecoverLogger = panicRecoverLoggerCopy

		fmt.Println()
	}()

	// LoggerEnabled should return false when instance has not been initialised
	panicRecoverLogger = nil
	instance := GetPanicRecoveryLoggerInstance()
	if instance.LoggerEnabled {
		t.Error("Expected LoggerEnabled to be false, but got true")
	}

	// LoggerEnabled should return true when instance has been initialised as such
	InitPanicRecoveryLoggerInstance(true, "test")
	instance = GetPanicRecoveryLoggerInstance()
	if !instance.LoggerEnabled {
		t.Error("Expected LoggerEnabled to be true, but got false")
	}

	// LoggerEnabled should return false when instance has been initialised as such
	InitPanicRecoveryLoggerInstance(false, "test")
	instance = GetPanicRecoveryLoggerInstance()
	if instance.LoggerEnabled {
		t.Error("Expected LoggerEnabled to be false, but got true")
	}
}
