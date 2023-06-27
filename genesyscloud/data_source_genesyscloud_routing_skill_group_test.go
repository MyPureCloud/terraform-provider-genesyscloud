package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingSkillGroup(t *testing.T) {
	var (
		skillGroupResource    = "routing-skill-group"
		skillGroupDataSource  = "routing-skill-group-data"
		skillGroupName        = "Skillgroup" + uuid.NewString()
		skillGroupDescription = "description-" + uuid.NewString()
	)

	config := generateRoutingSkillGroupResourceBasic(
		skillGroupResource,
		skillGroupName,
		skillGroupDescription,
	) + generateRoutingSkillGroupDataSource(skillGroupDataSource, "genesyscloud_routing_skill_group."+skillGroupResource+".name", "genesyscloud_routing_skill_group."+skillGroupResource)

	resource.Test(t, resource.TestCase{

		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_skill_group."+skillGroupDataSource, "id", "genesyscloud_routing_skill_group."+skillGroupResource, "id"),
				),
			},
		},
	})

}

func generateRoutingSkillGroupDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_skill_group" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
