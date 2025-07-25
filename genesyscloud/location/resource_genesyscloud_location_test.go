package location

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourceLocationBasic(t *testing.T) {
	var (
		locResourceLabel1 = "test-location1"
		locResourceLabel2 = "test-location2"
		locName1          = "Terraform location" + uuid.NewString()
		locName2          = "Terraform location" + uuid.NewString()
		locName3          = "Terraform location" + uuid.NewString()
		locNotes1         = "HQ1"
		locNotes2         = "HQ2"
		emergencyNum1     = "+13173124756"
		emergencyNum2     = "+17654182735"
		locNumberDefault  = "default"
		locNumberElin     = "elin"

		street1  = "7601 Interactive Way"
		city1    = "Indianapolis"
		state1   = "IN"
		country1 = "US"
		zip1     = "46278"
		street2  = "2001 Junipero Serra Blvd"
		city2    = "Daly City"
		state2   = "CA"
		zip2     = "94014"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateLocationResource(
					locResourceLabel1,
					locName1,
					locNotes1,
					[]string{}, // no paths or emergency number
					GenerateLocationAddress(street1, city1, state1, country1, zip1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "name", locName1),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "notes", locNotes1),
					resource.TestCheckNoResourceAttr("genesyscloud_location."+locResourceLabel1, "path.%"),
					resource.TestCheckNoResourceAttr("genesyscloud_location."+locResourceLabel1, "emergency_number.%"),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "address.0.street1", street1),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "address.0.city", city1),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "address.0.state", state1),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "address.0.country", country1),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "address.0.zip_code", zip1),
				),
			},
			{
				// Update with new location path and number
				Config: GenerateLocationResource(
					locResourceLabel1,
					locName2,
					locNotes2,
					[]string{"genesyscloud_location." + locResourceLabel2 + ".id"},
					GenerateLocationEmergencyNum(
						emergencyNum1,
						util.NullValue, // Default number type
					),
					GenerateLocationAddress(street1, city1, state1, country1, zip1),
				) + GenerateLocationResourceBasic(locResourceLabel2, locName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "name", locName2),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "notes", locNotes2),
					resource.TestCheckResourceAttrPair("genesyscloud_location."+locResourceLabel1, "path.0", "genesyscloud_location."+locResourceLabel2, "id"),
					testCheckEmergencyNumber("genesyscloud_location."+locResourceLabel1, emergencyNum1),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "emergency_number.0.type", locNumberDefault),
				),
			},
			{
				// Update with new number and no path
				Config: GenerateLocationResource(
					locResourceLabel1,
					locName2,
					util.NullValue,
					[]string{},
					GenerateLocationEmergencyNum(
						emergencyNum2,
						strconv.Quote(locNumberElin),
					),
					GenerateLocationAddress(street1, city1, state1, country1, zip1),
				) + GenerateLocationResourceBasic(locResourceLabel2, locName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "name", locName2),
					resource.TestCheckNoResourceAttr("genesyscloud_location."+locResourceLabel1, "path.%"),
					testCheckEmergencyNumber("genesyscloud_location."+locResourceLabel1, emergencyNum2),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "emergency_number.0.type", locNumberElin),
				),
			},
			{
				// Remove number (cannot change address when emergency number is assigned)
				Config: GenerateLocationResource(
					locResourceLabel1,
					locName2,
					util.NullValue,
					[]string{},
					GenerateLocationAddress(street1, city1, state1, country1, zip1),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "name", locName2),
					resource.TestCheckNoResourceAttr("genesyscloud_location."+locResourceLabel1, "path.%"),
					resource.TestCheckNoResourceAttr("genesyscloud_location."+locResourceLabel1, "emergency_number.%"),
				),
			},
			{
				// Update address
				Config: GenerateLocationResource(
					locResourceLabel1,
					locName2,
					util.NullValue,
					[]string{},
					GenerateLocationAddress(street2, city2, state2, country1, zip2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "name", locName2),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "address.0.street1", street2),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "address.0.city", city2),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "address.0.state", state2),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "address.0.country", country1),
					resource.TestCheckResourceAttr("genesyscloud_location."+locResourceLabel1, "address.0.zip_code", zip2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_location." + locResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyLocationsDestroyed,
	})
}

func testVerifyLocationsDestroyed(state *terraform.State) error {
	locationsAPI := platformclientv2.NewLocationsApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_location" {
			continue
		}

		location, resp, err := locationsAPI.GetLocation(rs.Primary.ID, nil)
		if location != nil {
			if location.State != nil && *location.State == "deleted" {
				// Location deleted
				continue
			}
			return fmt.Errorf("Location (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Location not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All Locations destroyed
	return nil
}

func testCheckEmergencyNumber(resourceLabel string, expectedNumber string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		locResource, ok := state.RootModule().Resources[resourceLabel]
		if !ok {
			return fmt.Errorf("Failed to find location %s in state", resourceLabel)
		}
		locID := locResource.Primary.ID

		numMembersAttr, ok := locResource.Primary.Attributes["emergency_number.#"]
		if !ok || numMembersAttr != "1" {
			return fmt.Errorf("No emergency number found for location %s in state", locID)
		}

		stateNum := locResource.Primary.Attributes["emergency_number.0.number"]
		if !comparePhoneNumbers("", expectedNumber, stateNum, nil) {
			return fmt.Errorf("State emergency number %s does not match expected number %s", stateNum, locID)
		}
		return nil
	}
}
