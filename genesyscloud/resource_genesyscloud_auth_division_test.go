package genesyscloud

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

func TestAccResourceAuthDivisionBasic(t *testing.T) {
	var (
		divResource1 = "auth-division1"
		divName1     = "Terraform Div-" + uuid.NewString()
		divName2     = "Terraform Div-" + uuid.NewString()
		divDesc1     = "Terraform test division"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateAuthDivisionResource(
					divResource1,
					divName1,
					util.NullValue, // No description
					util.NullValue, // Not home division
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divResource1, "name", divName1),
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divResource1, "description", ""),
				),
			},
			{
				// Update with a new name and description
				Config: GenerateAuthDivisionResource(
					divResource1,
					divName2,
					strconv.Quote(divDesc1),
					util.NullValue, // Not home division
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
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Set home division description
				Config: GenerateAuthDivisionResource(
					divHomeRes,
					divHomeName,
					strconv.Quote(homeDesc),
					util.TrueValue, // Home division
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divHomeRes, "name", divHomeName),
					resource.TestCheckResourceAttr("genesyscloud_auth_division."+divHomeRes, "description", homeDesc),
					validateHomeDivisionID("genesyscloud_auth_division."+divHomeRes),
				),
			},
			{
				// Set home division description again (applying twice to allow for desc to update)
				Config: GenerateAuthDivisionResource(
					divHomeRes,
					divHomeName,
					strconv.Quote(homeDesc2),
					util.TrueValue, // Home division
				),
			},
			{
				// Set home division description again
				Config: GenerateAuthDivisionResource(
					divHomeRes,
					divHomeName,
					strconv.Quote(homeDesc2),
					util.TrueValue, // Home division
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
		} else if util.IsStatus404(resp) {
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
		homeDivID, err := util.GetHomeDivisionID()
		if err != nil {
			return fmt.Errorf("%v", err)
		}

		if divID != homeDivID {
			return fmt.Errorf("Resource %s division ID %s not equal to home division ID %s", divResourceName, divID, homeDivID)
		}
		return nil
	}
}
