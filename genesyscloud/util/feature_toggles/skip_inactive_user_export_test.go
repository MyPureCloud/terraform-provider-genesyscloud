package feature_toggles

import (
	"os"
	"testing"
)

func TestUnitSkipInactiveUserExportToggleExists(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		os.Unsetenv(SkipInactiveUserExportToggleName())
		if SkipInactiveUserExportToggleExists() {
			t.Fatal("expected toggle to be disabled when env var is unset")
		}
	})

	t.Run("set", func(t *testing.T) {
		os.Setenv(SkipInactiveUserExportToggleName(), "true")
		defer os.Unsetenv(SkipInactiveUserExportToggleName())

		if !SkipInactiveUserExportToggleExists() {
			t.Fatal("expected toggle to be enabled when env var is set")
		}
	})
}
