package delay

import (
	"os"
	"testing"
	"time"
)

// Test constants
const (
	testEnvVarName     = "TEST_DELAY_ENV_VAR"
	routingQueueEnvVar = "ROUTING_QUEUE_DELAY_MAX"
)

func TestUnitConfigurableDelay_NoEnvVar(t *testing.T) {
	// Ensure no environment variable is set
	os.Unsetenv(testEnvVarName)

	start := time.Now()
	ConfigurableDelay(testEnvVarName)
	duration := time.Since(start)

	// Should complete immediately (no delay)
	if duration > 100*time.Millisecond {
		t.Errorf("Expected no delay, but operation took %v", duration)
	}
}

func TestUnitConfigurableDelay_EnvVarPresentNoValue(t *testing.T) {
	// Set environment variable without value
	os.Setenv(testEnvVarName, "")
	defer os.Unsetenv(testEnvVarName)

	start := time.Now()
	ConfigurableDelay(testEnvVarName)
	duration := time.Since(start)

	// Should have some delay (0-7 seconds)
	if duration > 8*time.Second {
		t.Errorf("Expected delay within 0-7 seconds, but operation took %v", duration)
	}
}

func TestUnitConfigurableDelay_EnvVarWithValidValue(t *testing.T) {
	// Set environment variable with valid value
	os.Setenv(testEnvVarName, "3")
	defer os.Unsetenv(testEnvVarName)

	start := time.Now()
	ConfigurableDelay(testEnvVarName)
	duration := time.Since(start)

	// Should have some delay (0-3 seconds)
	if duration > 4*time.Second {
		t.Errorf("Expected delay within 0-3 seconds, but operation took %v", duration)
	}
}

func TestUnitConfigurableDelay_EnvVarWithInvalidValue(t *testing.T) {
	// Set environment variable with invalid value
	os.Setenv(testEnvVarName, "invalid")
	defer os.Unsetenv(testEnvVarName)

	start := time.Now()
	ConfigurableDelay(testEnvVarName)
	duration := time.Since(start)

	// Should fall back to default delay (0-7 seconds)
	if duration > 8*time.Second {
		t.Errorf("Expected delay within 0-7 seconds (default), but operation took %v", duration)
	}
}

func TestUnitConfigurableDelay_EnvVarWithZeroValue(t *testing.T) {
	// Set environment variable with zero value
	os.Setenv(testEnvVarName, "0")
	defer os.Unsetenv(testEnvVarName)

	start := time.Now()
	ConfigurableDelay(testEnvVarName)
	duration := time.Since(start)

	// Should complete immediately (0 seconds delay)
	if duration > 100*time.Millisecond {
		t.Errorf("Expected no delay, but operation took %v", duration)
	}
}

func TestUnitConfigurableDelay_EnvVarWithNegativeValue(t *testing.T) {
	// Set environment variable with negative value
	os.Setenv(testEnvVarName, "-5")
	defer os.Unsetenv(testEnvVarName)

	start := time.Now()
	ConfigurableDelay(testEnvVarName)
	duration := time.Since(start)

	// Should fall back to default delay (0-7 seconds)
	if duration > 8*time.Second {
		t.Errorf("Expected delay within 0-7 seconds (default), but operation took %v", duration)
	}
}

func TestUnitConfigurableDelay_EnvVarWithLargeValue(t *testing.T) {
	// Set environment variable with large value
	os.Setenv(testEnvVarName, "60")
	defer os.Unsetenv(testEnvVarName)

	start := time.Now()
	ConfigurableDelay(testEnvVarName)
	duration := time.Since(start)

	// Should have some delay (0-60 seconds)
	if duration > 61*time.Second {
		t.Errorf("Expected delay within 0-60 seconds, but operation took %v", duration)
	}
}

func TestUnitConfigurableDelay_DifferentEnvVarNames(t *testing.T) {
	// Test with different environment variable names
	testCases := []string{
		"USER_DELAY_MAX",
		"GROUP_DELAY_MAX",
		"CUSTOM_DELAY_MAX",
		"MY_RESOURCE_DELAY",
	}

	for _, envVarName := range testCases {
		t.Run(envVarName, func(t *testing.T) {
			// Set environment variable with value
			os.Setenv(envVarName, "2")
			defer os.Unsetenv(envVarName)

			start := time.Now()
			ConfigurableDelay(envVarName)
			duration := time.Since(start)

			// Should have some delay (0-2 seconds)
			if duration > 3*time.Second {
				t.Errorf("Expected delay within 0-2 seconds for %s, but operation took %v", envVarName, duration)
			}
		})
	}
}

func TestUnitConstants(t *testing.T) {
	// Test that constants are properly defined
	if DefaultMaxDelaySeconds != 7 {
		t.Errorf("Expected DefaultMaxDelaySeconds to be 7, got %d", DefaultMaxDelaySeconds)
	}
}
