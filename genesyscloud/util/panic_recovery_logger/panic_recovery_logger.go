package panic_recovery_logger

import (
	"fmt"
	tfExporterState "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"log"
	"os"
	"runtime/debug"
)

type PanicRecoveryLogger struct {
	LoggerEnabled bool
	FilePath      string

	writeStackTracesToFileAttr func(*PanicRecoveryLogger, any) error
	isExporterActiveAttr       func() bool
}

var panicRecoverLogger *PanicRecoveryLogger

func InitPanicRecoveryLoggerInstance(enabled bool, filepath string) {
	if err := clearFileContentsIfExists(filepath); err != nil {
		log.Println(err)
	}
	panicRecoverLogger = &PanicRecoveryLogger{
		LoggerEnabled: enabled,
		FilePath:      filepath,

		writeStackTracesToFileAttr: writeStackTracesToFileFn,
		isExporterActiveAttr:       isExporterActiveFn,
	}
}

func GetPanicRecoveryLoggerInstance() *PanicRecoveryLogger {
	if panicRecoverLogger == nil {
		InitPanicRecoveryLoggerInstance(false, "")
	}
	return panicRecoverLogger
}

// HandleRecovery
// In the case of an export, return an error to avoid exporting an invalid configuration.
// Next write the stack trace info to the log file. If the file writing is unsuccessful, return an error (or append
// to the existing error regarding export if not nil.)
func (p *PanicRecoveryLogger) HandleRecovery(r any, operation constants.CRUDOperation) (err error) {
	if operation == constants.Read && p.isExporterActive() {
		err = fmt.Errorf("failed to export resource because of stack trace: %s", r)
	}

	log.Printf("Writing stack traces to file %s", p.FilePath)
	writeErr := p.WriteStackTracesToFile(r)
	if writeErr == nil {
		return
	}

	// WriteStackTracesToFile failed - append error info
	if err != nil {
		err = fmt.Errorf("%w.\n%w", err, writeErr)
	} else {
		err = writeErr
	}

	return err
}

func (p *PanicRecoveryLogger) WriteStackTracesToFile(r any) error {
	return p.writeStackTracesToFileAttr(p, r)
}

func (p *PanicRecoveryLogger) isExporterActive() bool {
	return p.isExporterActiveAttr()
}

// appendToFile appends data to a file. If the file does not exist, it will be created.
func appendToFile(filename string, data []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("appendToFile: %w", err)
		}
	}()

	// Open file with append mode (O_APPEND), create if it doesn't exist (O_CREATE),
	// and set write permission (O_WRONLY)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Write the data to the file
	_, err = file.Write(data)
	if err != nil {
		_ = file.Close()
		return err
	}
	_ = file.Close()
	return err
}

// clearFileContentsIfExists deletes file at filepath if it exists and does nothing if it doesn't
func clearFileContentsIfExists(filepath string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("clearFileContentsIfExists failed: %w (%s may contain stack traces from the previous deployment)", err, filepath)
		}
	}()

	if filepath == "" {
		return nil
	}

	// Check if file exists
	_, err = os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist
			return nil
		}
		// Return other errors (permission issues, etc.)
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	// Open file with truncate flag to clear contents
	file, err := os.OpenFile(filepath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to clear contents of %s: %w", filepath, err)
	}
	_ = file.Close()

	return err
}

func writeStackTracesToFileFn(p *PanicRecoveryLogger, r any) error {
	tracesToWrite := fmt.Sprintf("\nStacktrace recovered: %v. %s", r, string(debug.Stack()))
	if err := appendToFile(p.FilePath, []byte(tracesToWrite)); err != nil {
		return fmt.Errorf("WriteStackTracesToFile: failed to write to %s: %w", p.FilePath, err)
	}
	return nil
}

func isExporterActiveFn() bool {
	return tfExporterState.IsExporterActive()
}
