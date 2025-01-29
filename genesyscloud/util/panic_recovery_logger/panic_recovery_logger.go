package panic_recovery_logger

import (
	"fmt"
	"os"
	"runtime/debug"
)

type PanicRecoveryLogger struct {
	LoggerEnabled bool
	filePath      string
}

var panicRecoverLogger *PanicRecoveryLogger

func InitPanicRecoveryLoggerInstance(enabled bool, filepath string) {
	panicRecoverLogger = &PanicRecoveryLogger{
		LoggerEnabled: enabled,
		filePath:      filepath,
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

func (p *PanicRecoveryLogger) WriteStackTracesToFile(r any) error {
	tracesToWrite := fmt.Sprintf("\nStacktrace recovered: %v. %s", r, string(debug.Stack()))
	if err := appendToFile(p.filePath, []byte(tracesToWrite)); err != nil {
		return fmt.Errorf("WriteStackTracesToFile: failed to write to %s: %w", p.filePath, err)
	}
	return nil
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
