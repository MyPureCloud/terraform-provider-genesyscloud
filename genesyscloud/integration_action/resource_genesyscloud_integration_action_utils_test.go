package integration_action

import (
	"context"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

func strPtr(s string) *string { return &s }

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

// TestBuildIntegrationActionBlockLabel verifies the export block label format:
//   - custom actions keep the legacy "<category>_<name>" format so existing exports
//     remain stable;
//   - static actions are prefixed with the parent integration name and a three-underscore
//     delimiter ("<integrationName>___<category>_<name>") to disambiguate copies that
//     share a name across integration instances and keep the integration name visually
//     separable from the category/name pair;
//   - missing/unknown integration metadata falls back to the legacy format.
func TestBuildIntegrationActionBlockLabel(t *testing.T) {
	const customId = "9b1d8c50-cafe-4b1a-b0c0-feeddeadbeef"
	const staticId = "static_e7b86b86-abcd-4242-9999-1234567890ab"

	tests := []struct {
		name                 string
		action               platformclientv2.Action
		integrationNamesById map[string]string
		want                 string
	}{
		{
			name: "custom action keeps legacy label",
			action: platformclientv2.Action{
				Id:            strPtr(customId),
				Name:          strPtr("My Action"),
				Category:      strPtr("My Category"),
				IntegrationId: strPtr("integ-1"),
			},
			integrationNamesById: map[string]string{"integ-1": "Primary Integration"},
			want:                 "My Category_My Action",
		},
		{
			name: "static action is prefixed with parent integration name",
			action: platformclientv2.Action{
				Id:            strPtr(staticId),
				Name:          strPtr("Get User"),
				Category:      strPtr("Genesys Cloud Data Actions"),
				IntegrationId: strPtr("integ-1"),
			},
			integrationNamesById: map[string]string{"integ-1": "Primary Integration"},
			want:                 "Primary Integration___Genesys Cloud Data Actions_Get User",
		},
		{
			name: "static action without lookup entry falls back",
			action: platformclientv2.Action{
				Id:            strPtr(staticId),
				Name:          strPtr("Get User"),
				Category:      strPtr("Genesys Cloud Data Actions"),
				IntegrationId: strPtr("integ-unknown"),
			},
			integrationNamesById: map[string]string{"integ-1": "Primary Integration"},
			want:                 "Genesys Cloud Data Actions_Get User",
		},
		{
			name: "static action with nil integration id falls back",
			action: platformclientv2.Action{
				Id:       strPtr(staticId),
				Name:     strPtr("Get User"),
				Category: strPtr("Genesys Cloud Data Actions"),
			},
			integrationNamesById: nil,
			want:                 "Genesys Cloud Data Actions_Get User",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildIntegrationActionBlockLabel(tc.action, tc.integrationNamesById)
			if got != tc.want {
				t.Fatalf("buildIntegrationActionBlockLabel() = %q, want %q", got, tc.want)
			}
		})
	}
}
