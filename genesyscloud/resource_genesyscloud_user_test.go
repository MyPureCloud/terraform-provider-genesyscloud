package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/MyPureCloud/platform-client-sdk-go/platformclientv2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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

func TestAccResourceUserBasic(t *testing.T) {
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
	addrTypeMobile    = "MOBILE"
	addrTypeHome      = "HOME"
)

func TestAccResourceUserAddresses(t *testing.T) {
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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "email", addrEmail1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "name", addrUserName),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "phone_numbers.0.number", addrPhone1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "phone_numbers.0.media_type", phoneMediaType),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "phone_numbers.0.type", addrTypeWork),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "other_emails.0.address", addrEmail2),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "other_emails.0.type", addrTypeHome),
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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "email", addrEmail1),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "name", addrUserName),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "phone_numbers.0.number", addrPhone2),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "phone_numbers.0.media_type", smsMediaType),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "phone_numbers.0.type", addrTypeHome),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "phone_numbers.0.extension", addrPhoneExt),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "other_emails.0.address", addrEmail3),
					resource.TestCheckResourceAttr("genesyscloud_user."+addrUserResource1, "other_emails.0.type", addrTypeWork),
				),
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
