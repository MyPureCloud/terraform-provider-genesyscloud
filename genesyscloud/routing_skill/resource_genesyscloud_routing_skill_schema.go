package routing_skill

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_routing_skill"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceRoutingSkill())
	regInstance.RegisterExporter(ResourceType, RoutingSkillExporter())
	regInstance.RegisterDataSource(ResourceType, DataSourceRoutingSkill())
}

// The context is now added without Timeout ,
// since the warming up of cache will take place for the first Datasource registered during a Terraform Apply.
func DataSourceRoutingSkill() *schema.Resource {
	return &schema.Resource{
		Description:        "Data source for Genesys Cloud Routing Skills. Select a skill by name.",
		ReadWithoutTimeout: provider.ReadWithPooledClient(dataSourceRoutingSkillRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Skill name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func ResourceRoutingSkill() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Skill",

		CreateContext: provider.CreateWithPooledClient(createRoutingSkill),
		ReadContext:   provider.ReadWithPooledClient(readRoutingSkill),
		DeleteContext: provider.DeleteWithPooledClient(deleteRoutingSkill),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Skill name. Changing the name attribute will cause the skill object object to dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func RoutingSkillExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(GetAllRoutingSkills),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}
