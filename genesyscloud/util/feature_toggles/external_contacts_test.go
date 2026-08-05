package feature_toggles

import (
	"os"
	"testing"
)

func TestUnitSimpleExternalContactBlockLabelToggleExists(t *testing.T) {
	toggleName := SimpleExternalContactBlockLabelToggleName()
	originalValue, originallySet := os.LookupEnv(toggleName)
	t.Cleanup(func() {
		if originallySet {
			os.Setenv(toggleName, originalValue)
		} else {
			os.Unsetenv(toggleName)
		}
	})

	t.Run("unset", func(t *testing.T) {
		os.Unsetenv(toggleName)
		if SimpleExternalContactBlockLabelToggleExists() {
			t.Fatal("expected simple block labels to be disabled by default")
		}
	})

	t.Run("set", func(t *testing.T) {
		t.Setenv(toggleName, "")
		if !SimpleExternalContactBlockLabelToggleExists() {
			t.Fatal("expected simple block labels when the environment variable is set")
		}
	})
}
