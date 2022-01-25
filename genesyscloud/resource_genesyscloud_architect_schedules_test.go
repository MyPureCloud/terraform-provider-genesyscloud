package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func TestAccResourceArchitectSchedules(t *testing.T) {
	var (
		schedResource1 = "arch-sched1"
		name           = "CX as Code Schedule " + uuid.NewString()
		description    = "Sample Schedule by CX as Code"
		start          = "2021-08-04T08:00:00.000000"
		start2         = "2021-08-04T09:00:00.000000"
		end            = "2021-08-04T17:00:00.000000"
		rrule          = "FREQ=DAILY;INTERVAL=1"

		schedResource2 = "arch-sched2"
		name2          = "CX as Code Schedule 2 " + uuid.NewString()

		divResource = "test-division"
		divName     = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateArchitectSchedulesResource(
					schedResource1,
					name,
					nullValue,
					description,
					start,
					end,
					rrule,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource1, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource1, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource1, "start", start),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource1, "end", end),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource1, "rrule", rrule),
					testDefaultHomeDivision("genesyscloud_architect_schedules."+schedResource1),
				),
			},
			{
				// Update start time
				Config: generateArchitectSchedulesResource(
					schedResource1,
					name,
					nullValue,
					description,
					start2,
					end,
					rrule,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource1, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource1, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource1, "start", start2),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource1, "end", end),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource1, "rrule", rrule),
					testDefaultHomeDivision("genesyscloud_architect_schedules."+schedResource1),
				),
			},
			{
				// Create with new division
				Config: generateAuthDivisionBasic(divResource, divName) + generateArchitectSchedulesResource(
					schedResource2,
					name2,
					"genesyscloud_auth_division."+divResource+".id",
					description,
					start,
					end,
					rrule,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource2, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource2, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource2, "start", start),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource2, "end", end),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedules."+schedResource2, "rrule", rrule),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedules."+schedResource2, "division_id", "genesyscloud_auth_division."+divResource, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_architect_schedules." + schedResource2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySchedulesDestroyed,
	})
}

func generateArchitectSchedulesResource(
	schedResource1 string,
	name string,
	divisionId string,
	description string,
	start string,
	end string,
	rrule string) string {
	return fmt.Sprintf(`resource "genesyscloud_architect_schedules" "%s" {
		name = "%s"
		division_id = %s
		description = "%s"
		start = "%s"
		end = "%s"
		rrule = "%s"
	}
	`, schedResource1, name, divisionId, description, start, end, rrule)
}

func testVerifySchedulesDestroyed(state *terraform.State) error {
	archAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_schedules" {
			continue
		}

		sched, resp, err := archAPI.GetArchitectSchedule(rs.Primary.ID)
		if sched != nil {
			return fmt.Errorf("Schedule (%s) still exists", rs.Primary.ID)
		} else if isStatus404(resp) {
			// Schedule not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All schedules destroyed
	return nil
}
