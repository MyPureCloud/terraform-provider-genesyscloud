package platform

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUnitPlatformString(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     string
	}{
		{
			name:     "terraform platform",
			platform: PlatformTerraform,
			want:     "terraform",
		},
		{
			name:     "opentofu platform",
			platform: PlatformOpenTofu,
			want:     "tofu",
		},
		{
			name:     "debug server platform",
			platform: PlatformDebugServer,
			want:     "debug-server",
		},
		{
			name:     "go lang platform",
			platform: PlatformGoLang,
			want:     "go",
		},
		{
			name:     "unknown platform",
			platform: Platform(99),
			want:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.String(); got != tt.want {
				t.Errorf("Platform.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnitPlatformValidate(t *testing.T) {
	tests := []struct {
		name        string
		platform    Platform
		setBinPath  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid terraform platform",
			platform:   PlatformTerraform,
			setBinPath: "/usr/local/bin/terraform",
			wantErr:    false,
		},
		{
			name:        "invalid platform",
			platform:    Platform(99),
			setBinPath:  "/usr/local/bin/terraform",
			wantErr:     true,
			errContains: "invalid platform value",
		},
		{
			name:        "empty binary path",
			platform:    PlatformTerraform,
			setBinPath:  "",
			wantErr:     false,
			errContains: "Unable to determine provider binary path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original and restore after test
			if platformConfigSingleton != nil {
				origPath := platformConfigSingleton.binaryPath
				defer func() { platformConfigSingleton.binaryPath = origPath }()
			} else {
				platformConfigSingleton = &platformConfig{}
			}

			platformConfigSingleton.binaryPath = tt.setBinPath
			platformConfigSingleton.platform = tt.platform

			err := platformConfigSingleton.platform.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Platform.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("error message '%v' does not contain '%v'", err.Error(), tt.errContains)
				}
			}
		})
	}
}

func TestUnitGetProviderRegistry(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     string
	}{
		{
			name:     "terraform registry",
			platform: PlatformTerraform,
			want:     "registry.terraform.io",
		},
		{
			name:     "opentofu registry",
			platform: PlatformOpenTofu,
			want:     "registry.opentofu.org",
		},
		{
			name:     "debug server registry",
			platform: PlatformDebugServer,
			want:     "registry.terraform.io",
		},
		{
			name:     "go lang registry",
			platform: PlatformGoLang,
			want:     "registry.terraform.io",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.GetProviderRegistry(); got != tt.want {
				t.Errorf("Platform.GetProviderRegistry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnitExecuteCommand(t *testing.T) {
	// Create a test binary
	tmpDir := t.TempDir()
	testBinary := filepath.Join(tmpDir, "test-binary")

	// Create a simple shell script that echoes its arguments
	script := `#!/bin/sh
echo "stdout output"
echo "stderr output" >&2
exit 0
`
	if err := os.WriteFile(testBinary, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		ctx          context.Context
		args         []string
		wantStdout   string
		wantStderr   string
		wantExitCode int
		wantErr      bool
	}{
		{
			name:         "successful command",
			ctx:          context.Background(),
			args:         []string{"version"},
			wantStdout:   "stdout output\n",
			wantStderr:   "stderr output\n",
			wantExitCode: 0,
			wantErr:      false,
		},
		{
			name:         "timeout context",
			ctx:          timeoutContext(t),
			args:         []string{"version"},
			wantExitCode: -1,
			wantErr:      true,
		},
		{
			name:    "invalid command",
			ctx:     context.Background(),
			args:    []string{"invalid-command"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original and restore after test
			origPath := platformConfigSingleton.binaryPath
			defer func() { platformConfigSingleton.binaryPath = origPath }()

			platformConfigSingleton.binaryPath = testBinary

			output, err := executePlatformCommand(tt.ctx, platformConfigSingleton.binaryPath, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if output.Stdout != tt.wantStdout {
					t.Errorf("stdout = %v, want %v", output.Stdout, tt.wantStdout)
				}
				if output.Stderr != tt.wantStderr {
					t.Errorf("stderr = %v, want %v", output.Stderr, tt.wantStderr)
				}
				if output.ExitCode != tt.wantExitCode {
					t.Errorf("exit code = %v, want %v", output.ExitCode, tt.wantExitCode)
				}
			}
		})
	}
}

func TestUnitIsDevelopmentPlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     bool
	}{
		{
			name:     "debug server",
			platform: PlatformDebugServer,
			want:     true,
		},
		{
			name:     "go lang",
			platform: PlatformGoLang,
			want:     true,
		},
		{
			name:     "terraform",
			platform: PlatformTerraform,
			want:     false,
		},
		{
			name:     "opentofu",
			platform: PlatformOpenTofu,
			want:     false,
		},
		{
			name:     "test2json",
			platform: PlatformDebugServer,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.IsDevelopmentPlatform(); got != tt.want {
				t.Errorf("Platform.IsDebugServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}

func timeoutContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	t.Cleanup(cancel)
	time.Sleep(time.Millisecond) // Ensure timeout
	return ctx
}
