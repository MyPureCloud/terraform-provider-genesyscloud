package routing_skill_group

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingSkillGroup(t *testing.T) {
	t.Parallel()
	var (
		skillGroupResourceLabel   = "routing-skill-groups"
		skillGroupDataSourceLabel = "routing-skill-groups-data"
		skillGroupName            = "Skillgroup" + uuid.NewString()
		skillGroupDescription     = "description-" + uuid.NewString()
	)

	config := GenerateRoutingSkillGroupResourceBasic(
		skillGroupResourceLabel,
		skillGroupName,
		skillGroupDescription,
	) + generateRoutingSkillGroupDataSource(skillGroupDataSourceLabel, "genesyscloud_routing_skill_group."+skillGroupResourceLabel+".name", "genesyscloud_routing_skill_group."+skillGroupResourceLabel)

	resource.Test(t, resource.TestCase{

		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_skill_group."+skillGroupDataSourceLabel, "id", "genesyscloud_routing_skill_group."+skillGroupResourceLabel, "id"),
				),
			},
		},
	})

}

func generateRoutingSkillGroupDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, ResourceType, resourceLabel, name, dependsOnResource)
}
