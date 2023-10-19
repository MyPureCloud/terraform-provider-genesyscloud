package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceAuthDivisionBasic(t *testing.T) {
	var (
		divResource1 = "auth-division1"
		divName1     = "Terraform Div-" + uuid.NewString()
		divName2     = "Terraform Div-" + uuid.NewString()
		divDesc1     = "Terraform test division"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateAuthDivisionResource(
					divResource1,
					divName1,
					nullValue, // No description
					nullValue, // Not home division
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divResource1, "name", divName1),
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divResource1, "description", ""),
				),
			},
			{
				// Update with a new name and description
				Config: generateAuthDivisionResource(
					divResource1,
					divName2,
					strconv.Quote(divDesc1),
					nullValue, // Not home division
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divResource1, "name", divName2),
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divResource1, "description", divDesc1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_auth_division." + divResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyDivisionsDestroyed,
	})
}

func TestAccResourceAuthDivisionHome(t *testing.T) {
	var (
		divHomeRes  = "auth-division-home"
		divHomeName = "New Home"
		homeDesc    = "Home"
		homeDesc2   = "Home Division"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Set home division description
				Config: generateAuthDivisionResource(
					divHomeRes,
					divHomeName,
					strconv.Quote(homeDesc),
					trueValue, // Home division
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divHomeRes, "name", divHomeName),
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divHomeRes, "description", homeDesc),
					validateHomeDivisionID("genesyscloud_auth_division."+divHomeRes),
				),
			},
			{
				// Set home division description again (applying twice to allow for desc to update)
				Config: generateAuthDivisionResource(
					divHomeRes,
					divHomeName,
					strconv.Quote(homeDesc2),
					trueValue, // Home division
				),
			},
			{
				// Set home division description again
				Config: generateAuthDivisionResource(
					divHomeRes,
					divHomeName,
					strconv.Quote(homeDesc2),
					trueValue, // Home division
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divHomeRes, "name", divHomeName),
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divHomeRes, "description", homeDesc2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_auth_division." + divHomeRes,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyDivisionsDestroyed,
	})
}

func generateAuthDivisionBasic(resourceID string, name string) string {
	return generateAuthDivisionResource(resourceID, name, nullValue, falseValue)
}

func generateAuthDivisionResource(
	resourceID string,
	name string,
	description string,
	home string) string {
	return fmt.Sprintf(`resource "genesyscloud_auth_division" "%s" {
		name = "%s"
		description = %s
		home = %s
	}
	`, resourceID, name, description, home)
}

func testVerifyDivisionsDestroyed(state *terraform.State) error {
	authAPI := platformclientv2.NewAuthorizationApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_auth_division" {
			continue
		}

		if rs.Primary.Attributes["home"] == "true" {
			// We do not delete home divisions
			continue
		}

		division, resp, err := authAPI.GetAuthorizationDivision(rs.Primary.ID, false)
		if division != nil {
			return fmt.Errorf("Division (%s) still exists", rs.Primary.ID)
		} else if IsStatus404(resp) {
			// Division not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All divisions destroyed
	return nil
}

func validateHomeDivisionID(divResourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		divResource, ok := state.RootModule().Resources[divResourceName]
		if !ok {
			return fmt.Errorf("Failed to find division %s in state", divResourceName)
		}
		divID := divResource.Primary.ID
		homeDivID, err := getHomeDivisionID()
		if err != nil {
			return fmt.Errorf("%v", err)
		}

		if divID != homeDivID {
			return fmt.Errorf("Resource %s division ID %s not equal to home division ID %s", divResourceName, divID, homeDivID)
		}
		return nil
	}
}
