package util

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/process"
)

// detectExecutingBinary returns the path of the currently executing binary by finding
// the parent process and determining its executable path. This is particularly useful
// for identifying if the code is running under a debug environment.
//
// Returns:
//   - string: The path to the executing binary
//   - error: An error if the process cannot be found or if the executable path cannot be determined
func detectExecutingBinary() (string, error) {
	ppid, err := os.FindProcess(os.Getppid())
	if err != nil {
		return "", err
	}
	tfProcess, err := process.NewProcess(int32(ppid.Pid))
	if err != nil {
		return "", err
	}

	exe, err := tfProcess.Exe()
	if err != nil {
		return "", err
	}

	return exe, nil
}

// ExecutePlatformCommand executes a command against the platform binary with the provided arguments
// within the given context. It captures both stdout and stderr output from the command execution.
//
// Parameters:
//   - ctx: Context for command execution and timeout control
//   - args: Slice of string arguments to pass to the command
//
// Returns:
//   - stdoutString: The stdout output from the command execution
//   - stderrString: The stderr output from the command execution
//   - error: An error if the command fails, times out, or if the platform binary cannot be detected
//
// The function will panic if it cannot detect the executing binary path
func ExecutePlatformCommand(ctx context.Context, args []string) (stdoutString string, stderrString string, err error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	tfpath, err := detectExecutingBinary()
	if err != nil {
		log.Print("Could not find the executing binary")
		return "", "", err
	}

	cmd := exec.CommandContext(ctx, tfpath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Args = append(cmd.Args, args...)

	log.Printf("Running command against platform binary: %s", cmd.String())
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", stderr.String(), ctx.Err()
		}
		return "", stderr.String(), err
	}

	return stdout.String(), stderr.String(), nil

}

// IsDebugServerExecution determines if the current process is running under a debug server
// by examining the executable path for common debug binary patterns.
//
// Returns:
//   - bool: true if running under a debug server (e.g., delve, debug binary), false otherwise
//
// Debug patterns checked:
//   - "__debug_bin"
//   - "dlv" (Delve debugger)
//   - "debug-server"
func IsDebugServerExecution() bool {
	exe, err := detectExecutingBinary()
	if err != nil {
		return false
	}

	debugPatterns := []string{
		"__debug_bin",
		"dlv", // Delve debugger
		"debug-server",
	}

	for _, pattern := range debugPatterns {
		if strings.Contains(exe, pattern) {
			return true
		}
	}

	return false
}
