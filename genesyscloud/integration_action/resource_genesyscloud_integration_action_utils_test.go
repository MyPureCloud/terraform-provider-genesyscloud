package integration_action

import (
	"context"
	"testing"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

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

// TestSameNameDifferentCategoryProducesStableDistinctLabels is a regression test for
// the bug where two integration actions with the same name but different categories
// would get non-deterministic labels on export, causing destructive plan changes.
//
// The fix includes the category in the BlockLabel so that same-named actions in
// different categories never collide in the sanitizer.
//
// This test runs the full pipeline: buildIntegrationActionBlockLabel -> sanitizer,
// and verifies the resulting labels are distinct AND stable across multiple iterations.
func TestSameNameDifferentCategoryProducesStableDistinctLabels(t *testing.T) {
	// Simulate two custom actions with the same name in different categories
	actionProd := platformclientv2.Action{
		Id:            strPtr("custom_-_aaaa-1111-bbbb-2222"),
		Name:          strPtr("Log Call"),
		Category:      strPtr("Navigator Data Actions - Production"),
		IntegrationId: strPtr("integ-prod"),
	}
	actionStag := platformclientv2.Action{
		Id:            strPtr("custom_-_cccc-3333-dddd-4444"),
		Name:          strPtr("Log Call"),
		Category:      strPtr("Navigator Data Actions - Staging"),
		IntegrationId: strPtr("integ-stag"),
	}

	// Run 100 iterations to confirm labels never flip (the old bug was non-deterministic)
	var firstLabelProd, firstLabelStag string

	for i := 0; i < 100; i++ {
		// Build labels the same way getAllIntegrationActions does
		idMetaMap := resourceExporter.ResourceIDMetaMap{
			*actionProd.Id: &resourceExporter.ResourceMeta{
				BlockLabel: buildIntegrationActionBlockLabel(actionProd, nil),
			},
			*actionStag.Id: &resourceExporter.ResourceMeta{
				BlockLabel: buildIntegrationActionBlockLabel(actionStag, nil),
			},
		}

		// Run through the sanitizer (same as the exporter does)
		sanitizer := resourceExporter.NewSanitizerProvider()
		sanitizer.S.Sanitize(idMetaMap)

		labelProd := idMetaMap[*actionProd.Id].BlockLabel
		labelStag := idMetaMap[*actionStag.Id].BlockLabel

		// Labels must be different
		if labelProd == labelStag {
			t.Fatalf("iteration %d: both actions got the same label %q — category not differentiating", i, labelProd)
		}

		// Labels must be stable across iterations
		if i == 0 {
			firstLabelProd = labelProd
			firstLabelStag = labelStag
		} else {
			if labelProd != firstLabelProd {
				t.Fatalf("iteration %d: Production label changed from %q to %q — labels are non-deterministic", i, firstLabelProd, labelProd)
			}
			if labelStag != firstLabelStag {
				t.Fatalf("iteration %d: Staging label changed from %q to %q — labels are non-deterministic", i, firstLabelStag, labelStag)
			}
		}
	}

	t.Logf("Production label (stable across 100 runs): %s", firstLabelProd)
	t.Logf("Staging label (stable across 100 runs):    %s", firstLabelStag)
}
