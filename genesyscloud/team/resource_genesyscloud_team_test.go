package team

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v129/platformclientv2"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_team_test.go contains all of the test cases for running the resource
tests for team.
*/

func TestAccResourceMembers(t *testing.T) {

	var (
		resourceId = "Teams" + uuid.NewString()
		name1      = "Test Teams " + uuid.NewString()

		divResource = "test-division"
		divName     = "terraform-" + uuid.NewString()

		testUserResource1 = "user_resource_1"
		testUserName1     = "nameUser1" + uuid.NewString()
		testUserEmail1    = fmt.Sprintf(randString(5) + "@" + randString(5) + ".com")

		testUserResource2 = "user_resource_2"
		testUserName2     = "nameUser2" + uuid.NewString()
		testUserEmail2    = fmt.Sprintf(randString(5) + "@" + randString(5) + ".com")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Team
				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) +
					GenerateUserWithDivisionId(testUserResource1, testUserName1, testUserEmail1, "genesyscloud_auth_division."+divResource+".id") +
					GenerateUserWithDivisionId(testUserResource2, testUserName2, testUserEmail2, "genesyscloud_auth_division."+divResource+".id") +
					generateTeamsWithMemberResource(
						resourceId,
						name1,
						[]string{"genesyscloud_user." + testUserResource1 + ".id", "genesyscloud_user." + testUserResource2 + ".id"},
						"genesyscloud_auth_division."+divResource+".id",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "name", name1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceId, "division_id", "genesyscloud_auth_division."+divResource, "id"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_team."+resourceId, "member_ids.0",
						"genesyscloud_user."+testUserResource1, "id"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_team."+resourceId, "member_ids.1",
						"genesyscloud_user."+testUserResource2, "id"),
				),
			},
			{
				// Update Team
				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) +
					GenerateUserWithDivisionId(testUserResource1, testUserName1, testUserEmail1, "genesyscloud_auth_division."+divResource+".id") +
					generateTeamsWithMemberResource(
						resourceId,
						name1,
						[]string{"genesyscloud_user." + testUserResource1 + ".id"},
						"genesyscloud_auth_division."+divResource+".id",
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "name", name1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceId, "division_id", "genesyscloud_auth_division."+divResource, "id"),
				),
				PreConfig: func() {
					// Wait for a specified duration - to avoid getting non empty plan
					time.Sleep(15 * time.Second)
				},
			},
			{
				// Read
				ResourceName:            "genesyscloud_team." + resourceId,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"member_ids"},
			},
		},
		CheckDestroy: testVerifyTeamDestroyed,
	})
}

func TestAccResourceTeam(t *testing.T) {
	var (
		resourceId   = "Teams" + uuid.NewString()
		name1        = "Test Teams " + uuid.NewString()
		description1 = "Test description"
		name2        = "Test Teams " + uuid.NewString()
		description2 = "A new description"

		divResource = "test-division"
		divName     = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Team
				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) + generateTeamResource(
					resourceId,
					name1,
					"genesyscloud_auth_division."+divResource+".id",
					description1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "name", name1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceId, "division_id", "genesyscloud_auth_division."+divResource, "id"),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "description", description1),
				),
			},
			{
				// Update Team
				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) + generateTeamResource(
					resourceId,
					name2,
					"genesyscloud_auth_division."+divResource+".id",
					description2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "name", name2),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceId, "division_id", "genesyscloud_auth_division."+divResource, "id"),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "description", description2),
				),
			},
			{
				// Read
				ResourceName:      "genesyscloud_team." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTeamDestroyed,
	})
}

func testVerifyTeamDestroyed(state *terraform.State) error {
	teamsAPI := platformclientv2.NewTeamsApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_team" {
			continue
		}

		team, resp, err := teamsAPI.GetTeam(rs.Primary.ID)
		if team != nil {
			return fmt.Errorf("team (%s) still exists", rs.Primary.ID)
		}
		if util.IsStatus404(resp) {
			continue
		}
		return fmt.Errorf("Unexpected error: %s", err)

	}

	return nil
}
