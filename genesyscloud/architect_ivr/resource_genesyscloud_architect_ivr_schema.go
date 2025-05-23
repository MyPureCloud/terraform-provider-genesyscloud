package architect_ivr

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ResourceType      = "genesyscloud_architect_ivr"
	maxDnisPerRequest = 50
)

// SetRegistrar registers all resources, data sources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceArchitectIvr())
	l.RegisterResource(ResourceType, ResourceArchitectIvrConfig())
	l.RegisterExporter(ResourceType, ArchitectIvrExporter())
}

// ArchitectIvrExporter returns the resourceExporter object used to hold the genesyscloud_architect_ivr exporter's config
func ArchitectIvrExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllIvrConfigs),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"open_hours_flow_id":    {RefType: "genesyscloud_flow"},
			"closed_hours_flow_id":  {RefType: "genesyscloud_flow"},
			"holiday_hours_flow_id": {RefType: "genesyscloud_flow"},
			"schedule_group_id":     {RefType: "genesyscloud_architect_schedulegroups"},
			"division_id":           {RefType: "genesyscloud_auth_division"},
		},
	}
}

// ResourceArchitectIvrConfig registers the genesyscloud_architect_ivr resource with Terraform
func ResourceArchitectIvrConfig() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud IVR config",

		CreateContext: provider.CreateWithPooledClient(createIvrConfig),
		ReadContext:   provider.ReadWithPooledClient(readIvrConfig),
		UpdateContext: provider.UpdateWithPooledClient(updateIvrConfig),
		DeleteContext: provider.DeleteWithPooledClient(deleteIvrConfig),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the IVR config. Note: If the name changes, the existing Genesys Cloud IVR config will be dropped and recreated with a new ID. This can cause an Architect Flow to become invalid if the old flow is reference in the flow.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "IVR Config description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"dnis": {
				Description: fmt.Sprintf("The phone number(s) to contact the IVR by. Each phone number in the array must be in an E.164 number format. (Note: An array with a length greater than %v will be broken into chunks and uploaded in subsequent PUT requests.)", maxDnisPerRequest),
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    false,
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateDiagFunc: validators.ValidatePhoneNumber},
			},
			"open_hours_flow_id": {
				Description: "ID of inbound call flow for open hours.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"closed_hours_flow_id": {
				Description: "ID of inbound call flow for closed hours.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"holiday_hours_flow_id": {
				Description: "ID of inbound call flow for holidays.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"schedule_group_id": {
				Description: "Schedule group ID.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"division_id": {
				Description: "Division ID.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// DataSourceArchitectIvr registers the genesyscloud_architect_ivr data source
func DataSourceArchitectIvr() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud IVRs. Select an IVR by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceIvrRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "IVR name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
