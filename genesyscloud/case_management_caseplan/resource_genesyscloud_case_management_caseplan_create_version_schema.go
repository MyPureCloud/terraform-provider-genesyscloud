package case_management_caseplan

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// CreateVersionResourceType is the Terraform type for POST .../versions.
const CreateVersionResourceType = "genesyscloud_case_management_caseplan_create_version"

const createVersionResourceLogName = "genesyscloud_case_management_caseplan_create_version"

// ResourceCaseManagementCaseplanCreateVersion registers the create-version resource.
func ResourceCaseManagementCaseplanCreateVersion() *schema.Resource {
	return &schema.Resource{
		Description: `Calls POST /api/v2/casemanagement/caseplans/{caseplanId}/versions (no body). Use after a publish when the caseplan has no draft; creates a new draft (latest becomes published+1). Increment revision to create another draft later. Destroy only removes this from Terraform state (no delete-version API).`,

		CreateContext: provider.CreateWithPooledClient(createCaseManagementCaseplanCreateVersion),
		ReadContext:   provider.ReadWithPooledClient(readCaseManagementCaseplanCreateVersion),
		UpdateContext: provider.UpdateWithPooledClient(updateCaseManagementCaseplanCreateVersion),
		DeleteContext: provider.DeleteWithPooledClient(deleteCaseManagementCaseplanCreateVersion),
		Importer: &schema.ResourceImporter{
			StateContext: importCaseManagementCaseplanCreateVersion,
		},
		Schema: map[string]*schema.Schema{
			"caseplan_id": {
				Description: `Caseplan UUID.`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"revision": {
				Description: `Bump this integer to call POST .../versions again after a later publish (when there is no open draft).`,
				Optional:    true,
				Default:     0,
				Type:        schema.TypeInt,
			},
		},
	}
}

func importCaseManagementCaseplanCreateVersion(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	if err := d.Set("caseplan_id", d.Id()); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
