package responsemanagement_responseasset

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
)

/*
The resource_genesyscloud_responsemanagement_responseasset.go contains all of the methods that perform the core logic for a resource.
*/

// createResponsemanagementResponseasset is used by the responsemanagement_responseasset resource to create Genesys cloud responsemanagement responseasset
func createResponsemanagementResponseasset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// readResponsemanagementResponseasset is used by the responsemanagement_responseasset resource to read an responsemanagement responseasset from genesys cloud
func readResponsemanagementResponseasset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// updateResponsemanagementResponseasset is used by the responsemanagement_responseasset resource to update an responsemanagement responseasset in Genesys Cloud
func updateResponsemanagementResponseasset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

// deleteResponsemanagementResponseasset is used by the responsemanagement_responseasset resource to delete an responsemanagement responseasset from Genesys cloud
func deleteResponsemanagementResponseasset(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
