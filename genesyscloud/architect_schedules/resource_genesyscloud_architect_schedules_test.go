package architect_schedules

import (
	"fmt"
	"regexp"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

const errorMessageToMatch = "invalid start date."

func TestAccResourceArchitectSchedules(t *testing.T) {
	var (
		schedResourceLabel1 = "arch-sched1"
		name                = "CX as Code Schedule " + uuid.NewString()
		description         = "Sample Schedule by CX as Code"
		start               = "2021-08-04T08:00:00.000000"
		start2              = "2021-08-04T09:00:00.000000"
		end                 = "2021-08-04T17:00:00.000000"
		rrule               = "FREQ=DAILY;INTERVAL=1"

		schedResourceLabel2 = "arch-sched2"
		name2               = "CX as Code Schedule 2 " + uuid.NewString()

		divResourceLabel = "test-division"
		divName          = "terraform-" + uuid.NewString()

		resourcePath1 = ResourceType + "." + schedResourceLabel1
		resourcePath2 = ResourceType + "." + schedResourceLabel2
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateArchitectSchedulesResource(
					schedResourceLabel1,
					name,
					util.NullValue,
					description,
					start,
					end,
					rrule,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath1, "name", name),
					resource.TestCheckResourceAttr(resourcePath1, "description", description),
					resource.TestCheckResourceAttr(resourcePath1, "start", start),
					resource.TestCheckResourceAttr(resourcePath1, "end", end),
					resource.TestCheckResourceAttr(resourcePath1, "rrule", rrule),
					provider.TestDefaultHomeDivision(resourcePath1),
				),
			},
			{
				// Update start time
				Config: GenerateArchitectSchedulesResource(
					schedResourceLabel1,
					name,
					util.NullValue,
					description,
					start2,
					end,
					rrule,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath1, "name", name),
					resource.TestCheckResourceAttr(resourcePath1, "description", description),
					resource.TestCheckResourceAttr(resourcePath1, "start", start2),
					resource.TestCheckResourceAttr(resourcePath1, "end", end),
					resource.TestCheckResourceAttr(resourcePath1, "rrule", rrule),
					provider.TestDefaultHomeDivision(resourcePath1),
				),
			},
			{
				// Create with new division
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + GenerateArchitectSchedulesResource(
					schedResourceLabel2,
					name2,
					"genesyscloud_auth_division."+divResourceLabel+".id",
					description,
					start,
					end,
					rrule,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath2, "name", name2),
					resource.TestCheckResourceAttr(resourcePath2, "description", description),
					resource.TestCheckResourceAttr(resourcePath2, "start", start),
					resource.TestCheckResourceAttr(resourcePath2, "end", end),
					resource.TestCheckResourceAttr(resourcePath2, "rrule", rrule),
					resource.TestCheckResourceAttrPair(resourcePath2, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      resourcePath2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySchedulesDestroyed,
	})
}

func TestAccResourceArchitectSchedulesCreateFailsWhenStartDateNotInRRule(t *testing.T) {
	var (
		schedResourceLabel = "schedule"
		name               = "CX as Code Schedule " + uuid.NewString()
		description        = "Sample Schedule by CX as Code"
		start              = "2021-08-07T22:00:00.000000" // Saturday
		end                = "2021-08-08T23:00:00.000000"
		rrule              = "FREQ=DAILY;INTERVAL=1;BYDAY=MO,TU,WE,THU,FR,SU"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateArchitectSchedulesResource(
					schedResourceLabel,
					name,
					util.NullValue,
					description,
					start,
					end,
					rrule,
				),
				ExpectError: regexp.MustCompile(errorMessageToMatch),
			},
		},
		CheckDestroy: testVerifySchedulesDestroyed,
	})
}

func TestAccResourceArchitectSchedulesUpdateFailsWhenStartDateNotInRRule(t *testing.T) {
	var (
		schedResourceLabel = "schedule"
		name               = "CX as Code Schedule " + uuid.NewString()
		description        = "Sample Schedule by CX as Code"
		rrule              = "FREQ=DAILY;INTERVAL=1;BYDAY=MO,TU,WE,TH,FR"

		validStart = "2021-08-06T22:00:00.000000" // Friday
		validEnd   = "2021-08-06T23:00:00.000000"

		startUpdate = "2021-08-08T22:00:00.000000" // Sunday
		endUpdate   = "2021-08-09T22:00:00.000000"

		resourcePath = ResourceType + "." + schedResourceLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateArchitectSchedulesResource(
					schedResourceLabel,
					name,
					util.NullValue,
					description,
					validStart,
					validEnd,
					rrule,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", name),
					resource.TestCheckResourceAttr(resourcePath, "description", description),
					resource.TestCheckResourceAttr(resourcePath, "start", validStart),
					resource.TestCheckResourceAttr(resourcePath, "end", validEnd),
					resource.TestCheckResourceAttr(resourcePath, "rrule", rrule),
					provider.TestDefaultHomeDivision(resourcePath),
				),
			},
			{
				// Update
				Config: GenerateArchitectSchedulesResource(
					schedResourceLabel,
					name,
					util.NullValue,
					description,
					startUpdate,
					endUpdate,
					rrule,
				),
				ExpectError: regexp.MustCompile(errorMessageToMatch),
			},
		},
		CheckDestroy: testVerifySchedulesDestroyed,
	})
}

func testVerifySchedulesDestroyed(state *terraform.State) error {
	archAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		_, resp, err := archAPI.GetArchitectSchedule(rs.Primary.ID)
		if err != nil {
			if util.IsStatus404(resp) {
				// Schedule not found as expected
				continue
			}
			return fmt.Errorf("unexpected error: %s", err)
		}
		return fmt.Errorf("schedule (%s) still exists", rs.Primary.ID)
	}
	// Success. All schedules destroyed
	return nil
}
