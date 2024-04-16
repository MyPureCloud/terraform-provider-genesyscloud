package team

import (
	"fmt"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the team Data Source
*/

func TestAccDataSourceResourceTeam(t *testing.T) {
	var (
		teamResource = "team-resource"
		teamData     = "team-data"
		name         = "team" + uuid.NewString()
		description  = "Sample description"
		divResource  = "test-division"
		divName      = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{

				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) + generateTeamResource(
					teamResource,
					name,
					"genesyscloud_auth_division."+divResource+".id",
					description,
				) + generateTeamDataSource(
					teamData,
					name,
					"genesyscloud_team."+teamResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_team."+teamData, "id", "genesyscloud_team."+teamResource, "id"),
				),
			},
		},
	})
}

func generateTeamDataSource(resourceID string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_team" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
