package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v53/platformclientv2"
)

func TestAccResourceArchitectSchedules(t *testing.T) {
	var (
		schedResource1 = "arch-sched1"
		name           = "CX as Code Schedule"
		description    = "Sample Schedule by CX as Code"
		start          = "2021-08-04T08:00:00.000000"
		start2         = "2021-08-04T09:00:00.000000"
		end            = "2021-08-04T17:00:00.000000"
		rrule          = "FREQ=DAILY;INTERVAL=1"
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
				),
			},
			{
				// Update start time
				Config: generateArchitectSchedulesResource(
					schedResource1,
					name,
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
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_architect_schedules." + schedResource1,
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
	description string,
	start string,
	end string,
	rrule string) string {
	return fmt.Sprintf(`resource "genesyscloud_architect_schedules" "%s" {
		name = "%s"
		description = "%s"
		start = "%s"
		end = "%s"
		rrule = "%s"
	}
	`, schedResource1, name, description, start, end, rrule)
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
		} else if resp != nil && resp.StatusCode == 404 {
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
