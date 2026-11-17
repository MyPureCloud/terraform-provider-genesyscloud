package workforcemanagement_businessunits_activitycode

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesyscloud_workforcemanagement_businessunits_activitycode_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the workforcemanagement_businessunits_activitycode resource.
3.  The datasource schema definitions for the workforcemanagement_businessunits_activitycode datasource.
4.  The resource exporter configuration for the workforcemanagement_businessunits_activitycode exporter.
*/
const ResourceType = "genesyscloud_workforcemanagement_businessunits_activitycode"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceBusinessUnitActivityCode())
	regInstance.RegisterDataSource(ResourceType, DataSourceBusinessUnitActivityCode())
	regInstance.RegisterExporter(ResourceType, BusinessUnitActivityCodeExporter())
}

// ResourceBusinessUnitActivityCode registers the genesyscloud_workforcemanagement_businessunits_activitycode resource with Terraform
func ResourceBusinessUnitActivityCode() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud workforce management business unit activity code`,

		CreateContext: provider.CreateWithPooledClient(createBusinessUnitActivityCode),
		ReadContext:   provider.ReadWithPooledClient(readBusinessUnitActivityCode),
		UpdateContext: provider.UpdateWithPooledClient(updateBusinessUnitActivityCode),
		DeleteContext: provider.DeleteWithPooledClient(deleteBusinessUnitActivityCode),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"business_unit_id": {
				Description: `The id of the business unit this activity code belongs to. Changing this attribute will cause the activity code to be dropped and recreated.`,
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: `The name of the activity code. A default activity code can have an empty name, in which case Genesys Cloud shows the name of its category.`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"category": {
				Description: `The category of the activity code. Attempting to change the category of a default activity code will return an error. Valid values: OnQueueWork, Break, Meal, Meeting, OffQueueWork, TimeOff, Training, Unavailable.`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"length_in_minutes": {
				Description: `The default length of the activity in minutes.`,
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"counts_as_paid_time": {
				Description: `Whether an agent is paid while performing this activity.`,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"counts_as_work_time": {
				Description: `Whether the activity should be counted as contiguous work time for calculating daily constraints.`,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"agent_time_off_selectable": {
				Description: `Whether an agent can select this activity code when creating or editing a time off request. Only applicable when the category is TimeOff.`,
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"counts_toward_shrinkage": {
				Description: `Whether this activity code counts toward shrinkage calculations.`,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"planned_shrinkage": {
				Description: `Whether this activity code is considered planned (true) or unplanned (false) shrinkage.`,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"interruptible": {
				Description: `Whether this activity code is considered interruptible.`,
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"secondary_presence_ids": {
				Description: `The ids of the secondary presences of this activity code.`,
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"active": {
				Description: `Whether the activity code is active. Deleting an activity code in Genesys Cloud makes it inactive instead of removing it, and an inactive activity code can still be on old schedules and work plans. An inactive activity code is exported, but applying that configuration elsewhere creates an active one, because Genesys Cloud only creates active activity codes.`,
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"default_code": {
				Description: `Whether this is a default activity code. Default activity codes are created by Genesys Cloud, so they can only be managed by importing them, and they cannot be deleted. Destroying one only stops Terraform managing it.`,
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

// BusinessUnitActivityCodeExporter returns the resourceExporter object used to hold the genesyscloud_workforcemanagement_businessunits_activitycode exporter's config
func BusinessUnitActivityCodeExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllBusinessUnitActivityCodes),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"business_unit_id": {RefType: "genesyscloud_workforcemanagement_businessunits"},
		},
	}
}

// DataSourceBusinessUnitActivityCode registers the genesyscloud_workforcemanagement_businessunits_activitycode data source
func DataSourceBusinessUnitActivityCode() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud workforce management business unit activity code data source. Select a business unit activity code by name`,
		ReadContext: provider.ReadWithPooledClient(dataSourceBusinessUnitActivityCodeRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"business_unit_id": {
				Description: `The id of the business unit the activity code belongs to`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description:  `Business unit activity code name. Exactly one of name and category is required.`,
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "category"},
			},
			"category": {
				Description:  `The category of a default business unit activity code, which has one default activity code per category. Use it to select a default activity code, which has an empty name unless it was renamed. Exactly one of name and category is required.`,
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "category"},
			},
		},
	}
}
