package responsemanagement_response

import (
	"testing"

	respManagementRespAsset "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_responseasset"
)

// TestUnitResponseExporterResolvesAssetReferences guards against a regression of DEVTOOLING-1802.
// The asset_ids attribute must be registered as a reference to the fully-qualified
// genesyscloud_responsemanagement_responseasset resource type. If the RefType is missing the
// genesyscloud_ prefix (or is otherwise wrong), the exporter's exporters[RefType] lookup returns
// nil, so exported configs hard-code raw asset GUIDs instead of references and the referenced
// response assets are never pulled into the export.
func TestUnitResponseExporterResolvesAssetReferences(t *testing.T) {
	ref := ResponsemanagementResponseExporter().RefAttrs["asset_ids"]
	if ref == nil {
		t.Fatal("asset_ids is not a reference attribute; exported responses would hard-code asset GUIDs")
	}
	if ref.RefType != respManagementRespAsset.ResourceType {
		t.Errorf("asset_ids RefType = %q, want %q", ref.RefType, respManagementRespAsset.ResourceType)
	}
}
