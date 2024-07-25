package routing_skill_group

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const resourceName = "genesyscloud_routing_skill_group"

// SetRegistrar registers all the resources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(resourceName, ResourceRoutingSkillGroup())
	regInstance.RegisterDataSource(resourceName, DataSourceRoutingSkillGroup())
	regInstance.RegisterExporter(resourceName, ResourceSkillGroupExporter())
}

func ResourceRoutingSkillGroup() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud Skill Group`,

		CreateContext: provider.CreateWithPooledClient(createSkillGroups),
		ReadContext:   provider.ReadWithPooledClient(readSkillGroups),
		UpdateContext: provider.UpdateWithPooledClient(updateSkillGroups),
		DeleteContext: provider.DeleteWithPooledClient(deleteSkillGroups),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The group name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of the skill group",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"division_id": {
				Description: "The division to which this entity belongs",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"skill_conditions": {
				Description:      "JSON encoded array of rules that will be used to determine group membership.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: util.SuppressEquivalentJsonDiffs,
			},
			"member_division_ids": {
				Description: "The IDs of member divisions to add or remove for this skill group. An empty array means all divisions will be removed, '*' means all divisions will be added.",
				Type:        schema.TypeList,
				MaxItems:    50,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func DataSourceRoutingSkillGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Routing Skills Groups. Select a skill group by name.",
		ReadContext: provider.ReadWithPooledClient(dataSourceRoutingSkillGroupRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Skill group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func ResourceSkillGroupExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllRoutingSkillGroups),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id":      {RefType: "genesyscloud_auth_division"},
			"member_division_ids": {RefType: "genesyscloud_auth_division"},
		},
		RemoveIfMissing: map[string][]string{
			"division_id": {"division_id"},
		},
		JsonEncodeAttributes: []string{"skill_conditions"},
	}
}
