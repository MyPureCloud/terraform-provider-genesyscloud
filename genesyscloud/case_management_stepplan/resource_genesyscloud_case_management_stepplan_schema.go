package case_management_stepplan

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

const resourceName = "genesyscloud_case_management_stepplan"

// ResourceType is the Terraform type name for this resource.
const ResourceType = "genesyscloud_case_management_stepplan"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceCaseManagementStepplan())
	regInstance.RegisterDataSource(ResourceType, DataSourceCaseManagementStepplan())
	regInstance.RegisterExporter(ResourceType, CaseManagementStepplanExporter())
}

func ResourceCaseManagementStepplan() *schema.Resource {
	caseplanComputed := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"name": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
	stageplanComputed := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"name": {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
	workitemSettings := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"worktype_id": {
				Description: `UUID of the worktype (PATCH uses workitemSettings.worktypeId).`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			"worktype_name": {
				Description: `Worktype name from the API after read.`,
				Computed:    true,
				Type:        schema.TypeString,
			},
		},
	}

	return &schema.Resource{
		Description: `Binds to the single stepplan under a stage (auto-created with the caseplan). Uses version "latest" for list/read/PATCH. Create/delete are no-ops on the API; destroy removes Terraform state only.`,

		CreateContext: provider.CreateWithPooledClient(createCaseManagementStepplan),
		ReadContext:   provider.ReadWithPooledClient(readCaseManagementStepplan),
		UpdateContext: provider.UpdateWithPooledClient(updateCaseManagementStepplan),
		DeleteContext: provider.DeleteWithPooledClient(deleteCaseManagementStepplan),
		Importer: &schema.ResourceImporter{
			StateContext: importCaseManagementStepplan,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"caseplan_id": {
				Description: `Caseplan UUID.`,
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeString,
			},
			"stage_number": {
				Description:  `Stage ordinal 1–3 (same as genesyscloud_case_management_stageplan). The stepplan under that stage is resolved via list; exactly one stepplan per stage is required.`,
				Required:     true,
				ForceNew:     true,
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntInSlice([]int{1, 2, 3}),
			},
			"stageplan_id": {
				Description: `Resolved parent stageplan UUID.`,
				Computed:    true,
				Type:        schema.TypeString,
			},
			"stepplan_id": {
				Description: `Resolved stepplan UUID.`,
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
				Description: `Caseplan reference from the API.`,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        caseplanComputed,
			},
			"stageplan": {
				Description: `Stageplan reference from the API.`,
				Computed:    true,
				Type:        schema.TypeList,
				Elem:        stageplanComputed,
			},
			"activity_type": {
				Description: `e.g. workitem — passed to PATCH as activityType.`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			"workitem_settings": {
				Description: `Maps to workitemSettings on PATCH; use worktype_id for Workitem settings.`,
				Optional:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        workitemSettings,
			},
		},
	}
}

func CaseManagementStepplanExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthCaseManagementStepplans),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{},
	}
}

func DataSourceCaseManagementStepplan() *schema.Resource {
	return &schema.Resource{
		Description: `Looks up the composite Terraform id for a stepplan by caseplan_id and stage_number.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceCaseManagementStepplanRead),
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

func importCaseManagementStepplan(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	_ = ctx
	_ = meta
	id := d.Id()
	parts := strings.Split(id, stepplanResourceIDSeparator)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid id %q: expected caseplan_id|stage_number|stepplan_id", id)
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
	if err := d.Set("stepplan_id", parts[2]); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
