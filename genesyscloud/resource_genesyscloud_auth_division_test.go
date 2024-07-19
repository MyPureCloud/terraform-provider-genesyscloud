package genesyscloud

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceAuthDivisionBasic(t *testing.T) {
	var (
		divResource1 = "auth-division1"
		divName1     = "Terraform Div-" + uuid.NewString()
		divName2     = "Terraform Div-" + uuid.NewString()
		divDesc1     = "Terraform test division"
		divisionID   string
	)
	cleanupAuthDivision("Terraform")

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
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper updation
						return nil
					},
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["genesyscloud_auth_division."+divResource1]
						if !ok {
							return fmt.Errorf("not found: %s", "genesyscloud_auth_division."+divResource1)
						}
						divisionID = rs.Primary.ID
						log.Printf("Division ID: %s\n", divisionID) // Print ID
						return nil
					},
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
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper updation
						return nil
					},
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
				Destroy:           true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyDivisionsDestroyed(state)
		},
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
			},
			{
				// Set home division description again
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
				// Set home division description
				Config: GenerateAuthDivisionResource(
					divHomeRes,
					divHomeName,
					strconv.Quote(homeDesc2),
					util.TrueValue, // Home division
				),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper updation
						return nil
					},
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
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						time.Sleep(30 * time.Second) // Wait for 30 seconds for proper deletion
						return nil
					},
				),
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
		err := checkDivisionDeleted(rs.Primary.ID)(state)
		if err != nil {
			return err
		}
		fmt.Printf("Check complete for division ID: %s\n", rs.Primary.ID)

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

func cleanupAuthDivision(idPrefix string) {
	authAPI := platformclientv2.NewAuthorizationApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		divisions, _, getErr := authAPI.GetAuthorizationDivisions(pageSize, pageNum, "", nil, "", "", false, nil, "")
		if getErr != nil {
			log.Printf("failed to get auth division %s", getErr)
			return
		}

		if divisions.Entities == nil || len(*divisions.Entities) == 0 {
			break
		}

		for _, div := range *divisions.Entities {
			if div.Name != nil && strings.HasPrefix(*div.Name, idPrefix) {
				_, delErr := authAPI.DeleteAuthorizationDivision(*div.Id, true)
				if delErr != nil {
					log.Printf("failed to delete Auth division %s", delErr)
					return
				}
				log.Printf("Deleted auth division %s (%s)", *div.Id, *div.Name)
			}
		}
	}
}

func checkDivisionDeleted(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Printf("Fetching division with ID: %s\n", id)
		maxAttempts := 24
		for i := 0; i < maxAttempts; i++ {
			deleted, err := isDivisionDeleted(id)
			if err != nil {
				return err
			}
			if deleted {
				return nil
			}
			time.Sleep(10 * time.Second)
		}
		return fmt.Errorf("division %s was not deleted properly", id)
	}
}

func isDivisionDeleted(id string) (bool, error) {
	mu.Lock()
	defer mu.Unlock()

	authAPI := platformclientv2.NewAuthorizationApi()
	// Attempt to get the division
	_, response, err := authAPI.GetAuthorizationDivision(id, false)

	// Check if the division is not found (deleted)
	if response != nil && response.StatusCode == 404 {
		return true, nil // division is deleted
	}

	// Handle other errors
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		return false, err
	}

	// If division is found, it means the division is not deleted
	return false, nil
}
