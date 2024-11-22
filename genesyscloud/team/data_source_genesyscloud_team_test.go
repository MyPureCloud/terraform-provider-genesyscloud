package team

import (
	"fmt"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
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
		teamResourceLabel = "team-resource"
		teamDataLabel     = "team-data"
		name              = "team" + uuid.NewString()
		description       = "Sample description"
		divResourceLabel  = "test-division"
		divName           = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{

				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + generateTeamResource(
					teamResourceLabel,
					name,
					"genesyscloud_auth_division."+divResourceLabel+".id",
					description,
				) + generateTeamDataSource(
					teamDataLabel,
					name,
					"genesyscloud_team."+teamResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_team."+teamDataLabel, "id", "genesyscloud_team."+teamResourceLabel, "id"),
				),
			},
		},
	})
}

func generateTeamDataSource(resourceLabel string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_team" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
