package integration_action

import (
	"context"
	"testing"
)

// TestShouldExportIntegrationActionAsDataSource asserts that static (built-in)
// integration actions are flagged for data-source export while custom actions
// remain managed resources.
func TestShouldExportIntegrationActionAsDataSource(t *testing.T) {
	tests := []struct {
		name        string
		attributes  map[string]string
		expectAsDS  bool
		description string
	}{
		{
			name:        "static action is exported as data source",
			attributes:  map[string]string{"id": "static_e7b86b86-abcd-4242-9999-1234567890ab"},
			expectAsDS:  true,
			description: "IDs prefixed with 'static' represent built-in Genesys Cloud actions",
		},
		{
			name:        "static action with bare prefix is exported as data source",
			attributes:  map[string]string{"id": "static"},
			expectAsDS:  true,
			description: "any ID starting with the static prefix should be treated as built-in",
		},
		{
			name:        "custom action is exported as managed resource",
			attributes:  map[string]string{"id": "9b1d8c50-cafe-4b1a-b0c0-feeddeadbeef"},
			expectAsDS:  false,
			description: "custom action IDs are GUIDs and must remain managed resources",
		},
		{
			name:        "missing id falls back to managed resource",
			attributes:  map[string]string{},
			expectAsDS:  false,
			description: "if id is absent we cannot identify a static action, so default to managed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := shouldExportIntegrationActionAsDataSource(context.Background(), nil, tc.attributes)
			if err != nil {
				t.Fatalf("unexpected error for case %q: %v", tc.name, err)
			}
			if got != tc.expectAsDS {
				t.Fatalf("%s: shouldExportIntegrationActionAsDataSource() = %v, want %v", tc.description, got, tc.expectAsDS)
			}
		})
	}
}
