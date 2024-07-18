package team

import (
	"fmt"
	"math/rand"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_team_test.go contains all of the test cases for running the resource
tests for team.
*/

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

func TestAccResourceTeamAddMembers(t *testing.T) {
	var (
		resourceId   = "Team" + uuid.NewString()
		name1        = "Test Team " + uuid.NewString()
		description1 = "Test description"

		divResource = "test-division"
		divName     = "terraform-" + uuid.NewString()

		testUserResource1 = "user_resource_1"
		testUserName1     = "nameUser1" + uuid.NewString()
		testUserEmail1    = fmt.Sprintf(randString(5) + "@" + randString(5) + ".com")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Team
				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) +
					generateTeamResource(
						resourceId,
						name1,
						"genesyscloud_auth_division."+divResource+".id",
						description1,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "description", description1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceId, "division_id", "genesyscloud_auth_division."+divResource, "id"),
				),
			},
			{
				// Update Team with one member
				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) +
					generateUserWithDivisionId(testUserResource1, testUserName1, testUserEmail1, "genesyscloud_auth_division."+divResource+".id") +
					generateTeamResource(
						resourceId,
						name1,
						"genesyscloud_auth_division."+divResource+".id",
						description1,
						generateMemberIdsArray([]string{"genesyscloud_user." + testUserResource1 + ".id"}),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "description", description1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceId, "division_id", "genesyscloud_auth_division."+divResource, "id"),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "member_ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_team."+resourceId, "member_ids.0",
						"genesyscloud_user."+testUserResource1, "id"),
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

func TestAccResourceTeamRemoveMembers(t *testing.T) {
	var (
		resourceId   = "Team" + uuid.NewString()
		name1        = "Test Team " + uuid.NewString()
		description1 = "Test description"

		divResource = "test-division"
		divName     = "terraform-" + uuid.NewString()

		testUserResource1 = "user_resource_1"
		testUserName1     = "nameUser1" + uuid.NewString()
		testUserEmail1    = fmt.Sprintf(randString(5) + "@" + randString(5) + ".com")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Team with member
				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) +
					generateUserWithDivisionId(testUserResource1, testUserName1, testUserEmail1, "genesyscloud_auth_division."+divResource+".id") +
					generateTeamResource(
						resourceId,
						name1,
						"genesyscloud_auth_division."+divResource+".id",
						description1,
						generateMemberIdsArray([]string{"genesyscloud_user." + testUserResource1 + ".id"}),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "description", description1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceId, "division_id", "genesyscloud_auth_division."+divResource, "id"),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "member_ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_team."+resourceId, "member_ids.0",
						"genesyscloud_user."+testUserResource1, "id"),
				),
			},
			{
				// Update Team with no members
				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) +
					generateTeamResource(
						resourceId,
						name1,
						"genesyscloud_auth_division."+divResource+".id",
						description1,
						generateMemberIdsArray([]string{}),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "description", description1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceId, "division_id", "genesyscloud_auth_division."+divResource, "id"),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceId, "member_ids.#", "0"),
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

func generateTeamResource(
	teamResource string,
	name string,
	divisionId string,
	description string,
	memberIds ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_team" "%s" {
		name = "%s"
		division_id = %s
		description = "%s"
		%s
	}
	`, teamResource, name, divisionId, description, strings.Join(memberIds, "\n"))
}

func generateMemberIdsArray(memberIds []string) string {
	return fmt.Sprintf(`member_ids = [%s]`, strings.Join(memberIds, ", "))
}

func generateUserWithDivisionId(resourceID string, name string, email string, divisionId string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		name = "%s"
		email = "%s"
		division_id = %s
	}
	`, resourceID, name, email, divisionId)
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

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	s := make([]byte, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
