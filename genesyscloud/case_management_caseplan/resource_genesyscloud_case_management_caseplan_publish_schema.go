package case_management_caseplan

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// PublishResourceType is the Terraform type for the publish action resource.
const PublishResourceType = "genesyscloud_case_management_caseplan_publish"

const publishResourceLogName = "genesyscloud_case_management_caseplan_publish"

// ResourceCaseManagementCaseplanPublish registers the publish resource.
func ResourceCaseManagementCaseplanPublish() *schema.Resource {
	return &schema.Resource{
		Description: `Calls POST /api/v2/casemanagement/caseplans/{caseplanId}/publish. Use depends_on so this runs after stageplan and stepplan resources apply their PATCHes. Increment revision to publish again after later edits. Destroy only removes this from Terraform state (no unpublish API).`,

		CreateContext: provider.CreateWithPooledClient(createCaseManagementCaseplanPublish),
		ReadContext:   provider.ReadWithPooledClient(readCaseManagementCaseplanPublish),
		UpdateContext: provider.UpdateWithPooledClient(updateCaseManagementCaseplanPublish),
		DeleteContext: provider.DeleteWithPooledClient(deleteCaseManagementCaseplanPublish),
		Importer: &schema.ResourceImporter{
			StateContext: importCaseManagementCaseplanPublish,
		},
		Schema: map[string]*schema.Schema{
			"caseplan_id": {
				Description: `Caseplan UUID to publish.`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"revision": {
				Description: `Bump this integer to run publish again after changing stageplans or stepplans (or use terraform apply -replace on this resource).`,
				Optional:    true,
				Default:     0,
				Type:        schema.TypeInt,
			},
		},
	}
}

func importCaseManagementCaseplanPublish(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	if err := d.Set("caseplan_id", d.Id()); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
