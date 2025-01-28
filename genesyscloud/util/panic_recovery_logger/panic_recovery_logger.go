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

func (p *PanicRecoveryLogger) WriteStackTracesToFile(r any) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to write stack traces to file %s: %w", p.filePath, err)
		}
	}()

	tracesToWrite := fmt.Sprintf("\nStacktrace recovered: %v. %s", r, string(debug.Stack()))
	err = os.WriteFile(p.filePath, []byte(tracesToWrite), os.ModePerm)
	return err
}
