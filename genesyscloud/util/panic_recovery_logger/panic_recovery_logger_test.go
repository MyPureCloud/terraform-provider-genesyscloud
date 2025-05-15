package panic_recovery_logger

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"strings"
	"testing"
)

func TestUnitGetPanicRecoveryLoggerInstance(t *testing.T) {
	// restore after all
	panicRecoverLoggerCopy := panicRecoverLogger
	defer func() {
		panicRecoverLogger = panicRecoverLoggerCopy
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

func TestUnitHandleRecovery(t *testing.T) {
	// restore after all
	panicRecoverLoggerCopy := panicRecoverLogger
	defer func() {
		panicRecoverLogger = panicRecoverLoggerCopy
	}()

	const mockWriteErrorMessage = "mock error"

	InitPanicRecoveryLoggerInstance(true, "example/path.log")

	panicRecoverLogger.writeStackTracesToFileAttr = func(logger *PanicRecoveryLogger, a any) error {
		return nil
	}

	// 1. returns error if exporter is active
	panicRecoverLogger.isExporterActiveAttr = func() bool {
		return true
	}

	err := panicRecoverLogger.HandleRecovery(nil, constants.Read)
	if err == nil {
		t.Error("Expected error, but got nil")
	}

	// 2. returns nil if exporter not active, and file write successful
	panicRecoverLogger.isExporterActiveAttr = func() bool {
		return false
	}
	err = panicRecoverLogger.HandleRecovery(nil, constants.Read)
	if err != nil {
		t.Errorf("Expected nil error, got '%s'", err.Error())
	}

	// 3. returns error if operation exporter not active, but file write unsuccessful
	panicRecoverLogger.writeStackTracesToFileAttr = func(logger *PanicRecoveryLogger, a any) error {
		return fmt.Errorf(mockWriteErrorMessage)
	}

	err = panicRecoverLogger.HandleRecovery(nil, constants.Read)
	if err == nil {
		t.Errorf("Expected error '%s', got nil", mockWriteErrorMessage)
	} else if err.Error() != mockWriteErrorMessage {
		t.Errorf("Expected error '%s', got '%s'", mockWriteErrorMessage, err.Error())
	}

	// 4. returns error if #1 is true and file write unsuccessful
	panicRecoverLogger.isExporterActiveAttr = func() bool {
		return true
	}

	err = panicRecoverLogger.HandleRecovery(nil, constants.Read)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// verify that details of create issue and write are both in the error message
	if err != nil {
		const snippetOfExportErrorMessage = "failed to export resource"
		if !strings.Contains(err.Error(), mockWriteErrorMessage) || !strings.Contains(err.Error(), snippetOfExportErrorMessage) {
			t.Errorf("Expected error '%s' to contain '%s' and '%s'", err.Error(), mockWriteErrorMessage, snippetOfExportErrorMessage)
		}
	}
}
