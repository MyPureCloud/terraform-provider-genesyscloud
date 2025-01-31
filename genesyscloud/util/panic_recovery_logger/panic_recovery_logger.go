package panic_recovery_logger

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	tfExporterState "terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
)

type PanicRecoveryLogger struct {
	LoggerEnabled bool
	FilePath      string

	writeStackTracesToFileAttr func(*PanicRecoveryLogger, any) error
	isExporterActiveAttr       func() bool
}

var panicRecoverLogger *PanicRecoveryLogger

func InitPanicRecoveryLoggerInstance(enabled bool, filepath string) {
	if err := deleteFileIfExists(filepath); err != nil {
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
		return &PanicRecoveryLogger{
			LoggerEnabled: false,
		}
	}
	return panicRecoverLogger
}

// HandleRecovery — In the case of a Create: return an error object with stack trace info and warn of potential dangling resources.
// In the case of any export, return an error to avoid exporting an invalid configuration.
// Next and in any case, write the stack trace info to the log file. If the file writing is unsuccessful, we will fail to avoid the loss of data.
func (p *PanicRecoveryLogger) HandleRecovery(r any, operation constants.CRUDOperation) (err error) {
	if operation == constants.Create {
		err = fmt.Errorf("creation failed becasue of stack trace: %s. There may be dangling resource left in your org", r)
	} else if operation == constants.Read && p.isExporterActiveAttr() {
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

// deleteFileIfExists deletes file at filepath if it exists and does nothing if it doesn't
func deleteFileIfExists(filepath string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("deleteFileIfExists failed: %w (%s may contain stack traces from the previous deployment)", err, filepath)
		}
	}()

	// Check if file exists
	_, err = os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return nil as this is not an error condition
			return nil
		}
		// Return other errors (permission issues, etc.)
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	// File exists, try to remove it
	err = os.Remove(filepath)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", filepath, err)
	}

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
