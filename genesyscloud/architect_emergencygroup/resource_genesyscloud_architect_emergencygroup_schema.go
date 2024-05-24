package architect_emergencygroup

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_architect_emergencygroup"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceArchitectEmergencyGroup())
	regInstance.RegisterDataSource(resourceName, DataSourceArchitectEmergencyGroup())
	regInstance.RegisterExporter(resourceName, ArchitectEmergencyGroupExporter())
}

func ResourceArchitectEmergencyGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Emergency Group",

		CreateContext: provider.CreateWithPooledClient(createEmergencyGroup),
		ReadContext:   provider.ReadWithPooledClient(readEmergencyGroup),
		UpdateContext: provider.UpdateWithPooledClient(updateEmergencyGroup),
		DeleteContext: provider.DeleteWithPooledClient(deleteEmergencyGroup),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the emergency group. Note:  If the name is changed, the emergency group is dropped and recreated with a new ID. This can cause an Architect flow to be invalid if it references the old emergency group",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"division_id": {
				Description: "The division to which this emergency group will belong. If not set, the home division will be used.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "Description of the emergency group.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "The state of the emergency group. Defaults to false/inactive.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"emergency_call_flows": {
				Description: "The emergency call flows for this emergency group.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"emergency_flow_id": {
							Description: "The ID of the connected call flow.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"ivr_ids": {
							Description: "The IDs of the connected IVRs.",
							Type:        schema.TypeSet,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func DataSourceArchitectEmergencyGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Emergency Groups. Select an emergency group by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceEmergencyGroupRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Emergency Group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func ArchitectEmergencyGroupExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllEmergencyGroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":                            {RefType: "genesyscloud_auth_division"},
			"emergency_call_flows.emergency_flow_id": {RefType: "genesyscloud_flow"},
			"emergency_call_flows.ivr_ids":           {RefType: "genesyscloud_architect_ivr"},
		},
	}
}
