package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/platformclientv2"
)

func TestAccResourceUserBasic(t *testing.T) {
	var (
		userResource1 = "test-user1"
		userResource2 = "test-user2"
		email1        = "terraform-" + uuid.NewString() + "@example.com"
		email2        = "terraform-" + uuid.NewString() + "@example.com"
		email3        = "terraform-" + uuid.NewString() + "@example.com"
		userName1     = "John Terraform"
		userName2     = "Jim Terraform"
		stateActive   = "active"
		stateInactive = "inactive"
		title1        = "Senior Director"
		title2        = "Project Manager"
		department1   = "Development"
		department2   = "Project Management"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserResource(
					userResource1,
					email1,
					userName1,
					nullValue, // Defaults to active
					strconv.Quote(title1),
					strconv.Quote(department1),
					nullValue, // No manager
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "state", stateActive),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "title", title1),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "department", department1),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "password"),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "manager"),
					testDefaultHomeDivision("genesyscloud_user."+userResource1),
				),
			},
			{
				// Update
				Config: generateUserResource(
					userResource1,
					email2,
					userName2,
					strconv.Quote(stateInactive),
					strconv.Quote(title2),
					strconv.Quote(department2),
					nullValue, // No manager
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "email", email2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "name", userName2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "state", stateInactive),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "title", title2),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource1, "department", department2),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "manager"),
					testDefaultHomeDivision("genesyscloud_user."+userResource1),
				),
			},
			{
				// Create another user and set manager as existing user
				Config: generateUserResource(
					userResource1,
					email2,
					userName2,
					strconv.Quote(stateInactive),
					strconv.Quote(title2),
					strconv.Quote(department2),
					nullValue, // No manager
				) + generateUserResource(
					userResource2,
					email3,
					userName1,
					nullValue, // Active
					strconv.Quote(title1),
					strconv.Quote(department1),
					"genesyscloud_user."+userResource1+".id",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource2, "email", email3),
					resource.TestCheckResourceAttr("genesyscloud_user."+userResource2, "name", userName1),
					resource.TestCheckResourceAttrPair("genesyscloud_user."+userResource2, "manager", "genesyscloud_user."+userResource1, "id"),
					resource.TestCheckNoResourceAttr("genesyscloud_user."+userResource1, "manager"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_user." + userResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_user." + userResource2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserAddresses(t *testing.T) {
	var (
		addrUserResource1 = "test-user-addr"
		addrUserName      = "Nancy Terraform"
		addrEmail1        = "terraform-" + uuid.NewString() + "@example.com"
		addrEmail2        = "terraform-" + uuid.NewString() + "@example.com"
		addrEmail3        = "terraform-" + uuid.NewString() + "@example.com"
		addrPhone1        = "3174269078"
		addrPhone2        = "+441434634996"
		addrPhoneExt      = "1234"
		phoneMediaType    = "PHONE"
		smsMediaType      = "SMS"
		addrTypeWork      = "WORK"
		addrTypeHome      = "HOME"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateUserWithCustomAttrs(
					addrUserResource1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone1),
							nullValue, // Default to type PHONE
							nullValue, // Default to type WORK
							nullValue, // No extension
						),
						generateUserEmailAddress(
							strconv.Quote(addrEmail2),
							strconv.Quote(addrTypeHome),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "email", addrEmail1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "name", addrUserName),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.number", addrPhone1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.other_emails.0.address", addrEmail2),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.other_emails.0.type", addrTypeHome),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_user." + addrUserResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update phone number and other email attributes
				Config: generateUserWithCustomAttrs(
					addrUserResource1,
					addrEmail1,
					addrUserName,
					generateUserAddresses(
						generateUserPhoneAddress(
							strconv.Quote(addrPhone2),
							strconv.Quote(smsMediaType),
							strconv.Quote(addrTypeHome),
							strconv.Quote(addrPhoneExt),
						),
						generateUserEmailAddress(
							strconv.Quote(addrEmail3),
							strconv.Quote(addrTypeWork),
						),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "email", addrEmail1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "name", addrUserName),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.number", addrPhone2),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.media_type", smsMediaType),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.type", addrTypeHome),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.phone_numbers.0.extension", addrPhoneExt),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.other_emails.0.address", addrEmail3),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "addresses.0.other_emails.0.type", addrTypeWork),
				),
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserSkills(t *testing.T) {
	var (
		userResource1  = "test-user"
		email1         = "terraform-" + uuid.NewString() + "@example.com"
		userName1      = "Skill Terraform"
		skillResource1 = "test-skill-1"
		skillResource2 = "test-skill-2"
		skillName1     = "skill1-" + uuid.NewString()
		skillName2     = "skill2-" + uuid.NewString()
		proficiency1   = "1.5"
		proficiency2   = "2.5"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create user with 1 skill
				Config: generateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResource1+".id", proficiency1),
				) + generateRoutingSkillResource(skillResource1, skillName1),
				Check: resource.ComposeTestCheckFunc(
					validateUserSkill("genesyscloud_user."+userResource1, "genesyscloud_routing_skill."+skillResource1, proficiency1),
				),
			},
			{
				// Create another skill and add to the user
				Config: generateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResource1+".id", proficiency1),
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResource2+".id", proficiency2),
				) + generateRoutingSkillResource(
					skillResource1,
					skillName1,
				) + generateRoutingSkillResource(
					skillResource2,
					skillName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserSkill("genesyscloud_user."+userResource1, "genesyscloud_routing_skill."+skillResource1, proficiency1),
					validateUserSkill("genesyscloud_user."+userResource1, "genesyscloud_routing_skill."+skillResource2, proficiency2),
				),
			},
			{
				// Remove a skill from the user and modify proficiency
				Config: generateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoutingSkill("genesyscloud_routing_skill."+skillResource2+".id", proficiency1),
				) + generateRoutingSkillResource(
					skillResource1,
					skillName1,
				) + generateRoutingSkillResource(
					skillResource2,
					skillName2,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserSkill("genesyscloud_user."+userResource1, "genesyscloud_routing_skill."+skillResource2, proficiency1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_user." + userResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func TestAccResourceUserRoles(t *testing.T) {
	var (
		userResource1 = "test-user"
		email1        = "terraform-" + uuid.NewString() + "@example.com"
		userName1     = "Role Terraform"
		roleResource1 = "test-role-1"
		roleResource2 = "test-role-2"
		roleName1     = "Terraform User Role Test1"
		roleName2     = "Terraform User Role Test2"
		roleDesc      = "Terraform user test role"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create user with 1 role in default division
				Config: generateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoles("genesyscloud_auth_role."+roleResource1+".id"),
				) + generateAuthRoleResource(roleResource1, roleName1, roleDesc),
				Check: resource.ComposeTestCheckFunc(
					validateUserRole("genesyscloud_user."+userResource1, "genesyscloud_auth_role."+roleResource1),
				),
			},
			{
				// Create another role and add to the user
				Config: generateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoles("genesyscloud_auth_role."+roleResource1+".id"),
					generateUserRoles("genesyscloud_auth_role."+roleResource2+".id"),
				) + generateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateAuthRoleResource(
					roleResource2,
					roleName2,
					roleDesc,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserRole("genesyscloud_user."+userResource1, "genesyscloud_auth_role."+roleResource1),
					validateUserRole("genesyscloud_user."+userResource1, "genesyscloud_auth_role."+roleResource2),
				),
			},
			{
				// Remove a role from the user and modify division
				Config: generateUserWithCustomAttrs(
					userResource1,
					email1,
					userName1,
					generateUserRoles("genesyscloud_auth_role."+roleResource2+".id", strconv.Quote("*")),
				) + generateAuthRoleResource(
					roleResource1,
					roleName1,
					roleDesc,
				) + generateAuthRoleResource(
					roleResource2,
					roleName2,
					roleDesc,
				),
				Check: resource.ComposeTestCheckFunc(
					validateUserRole("genesyscloud_user."+userResource1, "genesyscloud_auth_role."+roleResource2, "*"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_user." + userResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyUsersDestroyed,
	})
}

func testVerifyUsersDestroyed(state *terraform.State) error {
	usersAPI := platformclientv2.NewUsersApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_user" {
			continue
		}

		user, resp, err := usersAPI.GetUser(rs.Primary.ID, nil, "", "")
		if user != nil {
			return fmt.Errorf("User (%s) still exists", rs.Primary.ID)
		} else if resp != nil && resp.StatusCode == 404 {
			// User not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All users destroyed
	return nil
}

func validateUserSkill(userResourceName string, skillResourceName string, proficiency string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		userResource, ok := state.RootModule().Resources[userResourceName]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourceName)
		}
		userID := userResource.Primary.ID

		skillResource, ok := state.RootModule().Resources[skillResourceName]
		if !ok {
			return fmt.Errorf("Failed to find skill %s in state", skillResourceName)
		}
		skillID := skillResource.Primary.ID

		numSkillsAttr, ok := userResource.Primary.Attributes["routing_skills.#"]
		if !ok {
			return fmt.Errorf("No skills found for user %s in state", userID)
		}

		numSkills, _ := strconv.Atoi(numSkillsAttr)
		for i := 0; i < numSkills; i++ {
			if userResource.Primary.Attributes["routing_skills."+strconv.Itoa(i)+".skill_id"] == skillID {
				if userResource.Primary.Attributes["routing_skills."+strconv.Itoa(i)+".proficiency"] == proficiency {
					// Found skill with correct proficiency
					return nil
				}
				return fmt.Errorf("Skill %s found for user %s with incorrect proficiency", skillID, userID)
			}
		}

		return fmt.Errorf("Skill %s not found for user %s in state", skillID, userID)
	}
}

func validateUserRole(userResourceName string, roleResourceName string, divisions ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		userResource, ok := state.RootModule().Resources[userResourceName]
		if !ok {
			return fmt.Errorf("Failed to find user %s in state", userResourceName)
		}
		userID := userResource.Primary.ID

		roleResource, ok := state.RootModule().Resources[roleResourceName]
		if !ok {
			return fmt.Errorf("Failed to find role %s in state", roleResourceName)
		}
		roleID := roleResource.Primary.ID

		if len(divisions) == 0 {
			// If no division specified, role should be in the home division
			homeDiv, err := getHomeDivisionID()
			if err != nil {
				return fmt.Errorf("Failed to query home div: %v", err)
			}
			divisions = []string{homeDiv}
		}

		userAttrs := userResource.Primary.Attributes
		numRolesAttr, _ := userAttrs["roles.#"]
		numRoles, _ := strconv.Atoi(numRolesAttr)
		for i := 0; i < numRoles; i++ {
			if userAttrs["roles."+strconv.Itoa(i)+".role_id"] == roleID {
				numDivsAttr, _ := userAttrs["roles."+strconv.Itoa(i)+".division_ids.#"]
				numDivs, _ := strconv.Atoi(numDivsAttr)
				stateDivs := make([]string, numDivs)
				for j := 0; j < numDivs; j++ {
					stateDivs[j] = userAttrs["roles."+strconv.Itoa(i)+".division_ids."+strconv.Itoa(j)]
				}

				extraDivs := sliceDifference(stateDivs, divisions)
				if len(extraDivs) > 0 {
					return fmt.Errorf("Unexpected divisions found for role %s in state: %v", roleID, extraDivs)
				}

				missingDivs := sliceDifference(divisions, stateDivs)
				if len(missingDivs) > 0 {
					return fmt.Errorf("Missing expected divisions for role %s in state: %v", roleID, missingDivs)
				}

				// Found expected role and divisions
				return nil
			}
		}
		return fmt.Errorf("Missing expected role for user %s in state: %s", userID, roleID)
	}
}

// Basic user with minimum required fields
func generateBasicUserResource(resourceID string, email string, name string) string {
	return generateUserResource(resourceID, email, name, nullValue, nullValue, nullValue, nullValue)
}

func generateUserResource(
	resourceID string,
	email string,
	name string,
	state string,
	title string,
	department string,
	manager string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		state = %s
		title = %s
		department = %s
		manager = %s
	}
	`, resourceID, email, name, state, title, department, manager)
}

func generateUserWithCustomAttrs(resourceID string, email string, name string, attrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_user" "%s" {
		email = "%s"
		name = "%s"
		%s
	}
	`, resourceID, email, name, strings.Join(attrs, "\n"))
}

func generateUserAddresses(nestedBlocks ...string) string {
	return fmt.Sprintf(`addresses {
		%s
	}
	`, strings.Join(nestedBlocks, "\n"))
}

func generateUserPhoneAddress(phoneNum string, phoneMediaType string, phoneType string, extension string) string {
	return fmt.Sprintf(`phone_numbers {
				number = %s
				media_type = %s
				type = %s
				extension = %s
			}
			`, phoneNum, phoneMediaType, phoneType, extension)
}

func generateUserEmailAddress(emailAddress string, emailType string) string {
	return fmt.Sprintf(`other_emails {
				address = %s
				type = %s
			}
			`, emailAddress, emailType)
}

func generateUserRoutingSkill(skillID string, proficiency string) string {
	return fmt.Sprintf(`routing_skills {
		skill_id = %s
		proficiency = %s
	}
	`, skillID, proficiency)
}

func generateUserRoles(skillID string, divisionIds ...string) string {
	var divAttr string
	if len(divisionIds) > 0 {
		divAttr = "division_ids = [" + strings.Join(divisionIds, ",") + "]"
	}
	return fmt.Sprintf(`roles {
		role_id = %s
		%s
	}
	`, skillID, divAttr)
}
