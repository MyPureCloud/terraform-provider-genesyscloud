package page_size

import (
	"os"
	"testing"
)

func TestUnitEnvVarName(t *testing.T) {
	got := EnvVarName("genesyscloud_routing_wrapupcode")
	want := "GENESYSCLOUD_PAGE_SIZE_genesyscloud_routing_wrapupcode"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestUnitForResource(t *testing.T) {
	const (
		resourceType = "genesyscloud_routing_wrapupcode"
		defaultSize  = 500
	)
	envVar := EnvVarName(resourceType)

	t.Run("unset uses default", func(t *testing.T) {
		os.Unsetenv(envVar)
		if got := ForResource(resourceType, defaultSize); got != defaultSize {
			t.Fatalf("expected %d, got %d", defaultSize, got)
		}
	})

	t.Run("valid override", func(t *testing.T) {
		os.Setenv(envVar, "100")
		defer os.Unsetenv(envVar)

		if got := ForResource(resourceType, defaultSize); got != 100 {
			t.Fatalf("expected 100, got %d", got)
		}
	})

	t.Run("empty value uses default", func(t *testing.T) {
		os.Setenv(envVar, "")
		defer os.Unsetenv(envVar)

		if got := ForResource(resourceType, defaultSize); got != defaultSize {
			t.Fatalf("expected %d, got %d", defaultSize, got)
		}
	})

	t.Run("invalid value uses default", func(t *testing.T) {
		os.Setenv(envVar, "not-a-number")
		defer os.Unsetenv(envVar)

		if got := ForResource(resourceType, defaultSize); got != defaultSize {
			t.Fatalf("expected %d, got %d", defaultSize, got)
		}
	})

	t.Run("non-positive value uses default", func(t *testing.T) {
		os.Setenv(envVar, "0")
		defer os.Unsetenv(envVar)

		if got := ForResource(resourceType, defaultSize); got != defaultSize {
			t.Fatalf("expected %d, got %d", defaultSize, got)
		}
	})
}
