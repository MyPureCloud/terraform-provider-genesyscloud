package team

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"math/rand"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_team_test.go contains all of the test cases for running the resource
tests for team.
*/

func TestAccResourceTeam(t *testing.T) {
	var (
		resourceLabel = "Teams" + uuid.NewString()
		name1         = "Test Teams " + uuid.NewString()
		description1  = "Test description"
		name2         = "Test Teams " + uuid.NewString()
		description2  = "A new description"

		divResourceLabel = "test-division"
		divName          = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Team
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + GenerateTeamResource(
					resourceLabel,
					name1,
					"genesyscloud_auth_division."+divResourceLabel+".id",
					description1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "description", description1),
				),
			},
			{
				// Update Team
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + GenerateTeamResource(
					resourceLabel,
					name2,
					"genesyscloud_auth_division."+divResourceLabel+".id",
					description2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "description", description2),
				),
			},
			{
				// Read
				ResourceName:      "genesyscloud_team." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTeamDestroyed,
	})
}

func TestAccResourceTeamAddMembers(t *testing.T) {
	var (
		resourceLabel = "Team" + uuid.NewString()
		name1         = "Test Team " + uuid.NewString()
		description1  = "Test description"

		divResourceLabel = "test-division"
		divName          = "terraform-" + uuid.NewString()

		testUserResourceLabel1 = "user_resource_1"
		testUserName1          = "nameUser1" + uuid.NewString()
		testUserEmail1         = fmt.Sprintf(randString(5) + "@" + randString(5) + ".com")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Team
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					GenerateTeamResource(
						resourceLabel,
						name1,
						"genesyscloud_auth_division."+divResourceLabel+".id",
						description1,
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
				),
			},
			{
				// Update Team with one member
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					generateUserWithDivisionId(testUserResourceLabel1, testUserName1, testUserEmail1, "genesyscloud_auth_division."+divResourceLabel+".id") +
					GenerateTeamResource(
						resourceLabel,
						name1,
						"genesyscloud_auth_division."+divResourceLabel+".id",
						description1,
						generateMemberIdsArray([]string{"genesyscloud_user." + testUserResourceLabel1 + ".id"}),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "member_ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_team."+resourceLabel, "member_ids.0",
						"genesyscloud_user."+testUserResourceLabel1, "id"),
				),
			},
			{
				// Read
				ResourceName:      "genesyscloud_team." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTeamDestroyed,
	})
}

func TestAccResourceTeamRemoveMembers(t *testing.T) {
	var (
		resourceLabel = "Team" + uuid.NewString()
		name1         = "Test Team " + uuid.NewString()
		description1  = "Test description"

		divResourceLabel = "test-division"
		divName          = "terraform-" + uuid.NewString()

		testUserResourceLabel1 = "user_resource_1"
		testUserName1          = "nameUser1" + uuid.NewString()
		testUserEmail1         = fmt.Sprintf(randString(5) + "@" + randString(5) + ".com")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create Team with member
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					generateUserWithDivisionId(testUserResourceLabel1, testUserName1, testUserEmail1, "genesyscloud_auth_division."+divResourceLabel+".id") +
					GenerateTeamResource(
						resourceLabel,
						name1,
						"genesyscloud_auth_division."+divResourceLabel+".id",
						description1,
						generateMemberIdsArray([]string{"genesyscloud_user." + testUserResourceLabel1 + ".id"}),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "member_ids.#", "1"),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_team."+resourceLabel, "member_ids.0",
						"genesyscloud_user."+testUserResourceLabel1, "id"),
				),
			},
			{
				// Update Team with no members
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					GenerateTeamResource(
						resourceLabel,
						name1,
						"genesyscloud_auth_division."+divResourceLabel+".id",
						description1,
						generateMemberIdsArray([]string{}),
					),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "description", description1),
					resource.TestCheckResourceAttrPair("genesyscloud_team."+resourceLabel, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
					resource.TestCheckResourceAttr("genesyscloud_team."+resourceLabel, "member_ids.#", "0"),
				),
			},
			{
				// Read
				ResourceName:      "genesyscloud_team." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyTeamDestroyed,
	})
}

func generateMemberIdsArray(memberIds []string) string {
	return fmt.Sprintf(`member_ids = [%s]`, strings.Join(memberIds, ", "))
}

func generateUserWithDivisionId(resourceLabel string, name string, email string, divisionId string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		name = "%s"
		email = "%s"
		division_id = %s
	}
	`, resourceLabel, name, email, divisionId)
}

func testVerifyTeamDestroyed(state *terraform.State) error {
	teamsAPI := platformclientv2.NewTeamsApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_team" {
			continue
		}

		team, resp, err := teamsAPI.GetTeam(rs.Primary.ID, "")
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
