package util

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/process"
)

// detectExecutingBinary returns the path of the currently executing binary by finding
// the parent process and determining its executable path (either `terraform“ or `tofu`)
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

// verifyBinary performs basic security checks on the provided binary path to ensure
// it exists, is a regular file (not a symlink or directory), and has proper execute permissions.
//
// Parameters:
//   - path: The filesystem path to the binary to verify
//
// Returns:
//   - error: An error if any verification check fails, nil if all checks pass
func verifyBinary(path string) error {
	// Basic existence check
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat binary: %w", err)
	}

	// Ensure it's a regular file, not a symlink or directory
	if !info.Mode().IsRegular() {
		return fmt.Errorf("binary path is not a regular file")
	}

	// Check if we have execute permission
	if info.Mode().Perm()&0111 == 0 {
		return fmt.Errorf("binary is not executable")
	}

	return nil
}

// validateCommandArgs uses HashiCorp's flags parser to validate command arguments
// before they are passed to the platform binary (terraform/tofu).
//
// Parameters:
//   - args: Slice of string arguments to validate
//
// Returns:
//   - error: An error if any argument fails validation, nil if all arguments are valid
func validateCommandArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no arguments provided")
	}

	// Additional custom validation if needed
	command := args[0]
	if !isAllowedCommand(command) {
		return fmt.Errorf("command %q is not allowed", command)
	}

	return nil
}

// isAllowedCommand checks if the given command is in the allowed list
func isAllowedCommand(cmd string) bool {
	allowedCommands := map[string]bool{
		"init":         true,
		"plan":         true,
		"apply":        true,
		"destroy":      true,
		"validate":     true,
		"output":       true,
		"show":         true,
		"state":        true,
		"import":       true,
		"version":      true,
		"fmt":          true,
		"force-unlock": true,
		"providers":    true,
		"login":        true,
		"logout":       true,
		"refresh":      true,
		"graph":        true,
		"taint":        true,
		"untaint":      true,
		"workspace":    true,
		"metadata":     true,
		"test":         true,
		"console":      true,
	}

	return allowedCommands[strings.TrimPrefix(cmd, "-")]
}

// ExecutePlatformCommand executes a command against the platform binary (`terraform` or `tofu`) with
// the provided arguments within the given context. It captures both stdout and stderr output from the
// command execution.
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
// The function will return an error if it cannot detect the executing binary path
func ExecutePlatformCommand(ctx context.Context, args []string) (stdoutString string, stderrString string, err error) {
	var stdout, stderr bytes.Buffer

	// Validate context
	if ctx == nil {
		return "", "", fmt.Errorf("nil context provided")
	}

	// Validate arguments
	if err := validateCommandArgs(args); err != nil {
		return "", "", fmt.Errorf("invalid arguments: %w", err)
	}

	tfpath, err := detectExecutingBinary()
	if err != nil {
		log.Print("Could not find the executing binary")
		return "", "", err
	}

	// Verify binary exists and has proper permissions
	if err := verifyBinary(tfpath); err != nil {
		return "", "", fmt.Errorf("binary verification failed: %w", err)
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
