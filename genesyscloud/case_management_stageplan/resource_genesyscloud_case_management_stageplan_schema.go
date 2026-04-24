package case_management_stageplan

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_case_management_stageplan"

// ResourceType is the Terraform type name for this resource.
const ResourceType = "genesyscloud_case_management_stageplan"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceCaseManagementStageplan())
	regInstance.RegisterDataSource(ResourceType, DataSourceCaseManagementStageplan())
	regInstance.RegisterExporter(ResourceType, CaseManagementStageplanExporter())
}

func ResourceCaseManagementStageplan() *schema.Resource {
	caseplanComputed := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: `Caseplan UUID from the API.`,
				Computed:    true,
				Type:        schema.TypeString,
			},
			"name": {
				Description: `Caseplan name from the API.`,
				Computed:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Binds to a stageplan auto-created with a caseplan (exactly 3 per caseplan). Uses version "latest" for list/read/PATCH. Create/delete are no-ops on the API; destroy removes Terraform state only.`,

		CreateContext: provider.CreateWithPooledClient(createCaseManagementStageplan),
		ReadContext:   provider.ReadWithPooledClient(readCaseManagementStageplan),
		UpdateContext: provider.UpdateWithPooledClient(updateCaseManagementStageplan),
		DeleteContext: provider.DeleteWithPooledClient(deleteCaseManagementStageplan),
		Importer: &schema.ResourceImporter{
			StateContext: importCaseManagementStageplan,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"caseplan_id": {
				Description: `Caseplan UUID. Stageplans are listed under GET .../caseplans/{caseplanId}/versions/latest/stageplans.`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"stage_number": {
				Description:  `Which auto-created stage to manage: 1, 2, or 3. Stages are sorted by name after list (e.g. Stage 1…Stage 3).`,
				Required:     true,
				ForceNew:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntInSlice([]int{1, 2, 3}),
			},
			"stageplan_id": {
				Description: `Resolved stageplan UUID.`,
				Computed:    true,
				Type:        schema.TypeString,
			},
			"name": {
				Description: `Patched name (optional).`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			"description": {
				Description: `Patched description (optional).`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			"caseplan": {
				Description: `Caseplan reference from the API read.`,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        caseplanComputed,
			},
		},
	}
}

func CaseManagementStageplanExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthCaseManagementStageplans),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}

func DataSourceCaseManagementStageplan() *schema.Resource {
	return &schema.Resource{
		Description: `Looks up the composite Terraform id for a stageplan by caseplan_id and stage_number (same ordering as the resource).`,
		ReadContext: provider.ReadWithPooledClient(dataSourceCaseManagementStageplanRead),
		Schema: map[string]*schema.Schema{
			"caseplan_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stage_number": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntInSlice([]int{1, 2, 3}),
			},
		},
	}
}

func importCaseManagementStageplan(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	_ = ctx
	_ = meta
	id := d.Id()
	parts := strings.Split(id, stageplanResourceIDSeparator)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid id %q: expected caseplan_id|stage_number|stageplan_id", id)
	}
	sn, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	if err := d.Set("caseplan_id", parts[0]); err != nil {
		return nil, err
	}
	if err := d.Set("stage_number", sn); err != nil {
		return nil, err
	}
	if err := d.Set("stageplan_id", parts[2]); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
