package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func TestAccResourceArchitectScheduleGroups(t *testing.T) {
	var (
		schedGroupResource1 = "arch-sched-group1"
		name                = "Schedule Group x" + uuid.NewString()
		description         = "Sample Schedule Group by CX as Code"
		time_zone           = "Asia/Singapore"

		schedResource1 = "arch-sched1"
		schedResource2 = "arch-sched2"
		schedResource3 = "arch-sched3"
		openSched      = "Open Schedule" + uuid.NewString()
		closedSched    = "Closed Schedule" + uuid.NewString()
		holidaySched   = "Holiday Schedule" + uuid.NewString()
		schedDesc      = "Sample Schedule by CX as Code"
		start          = "2021-08-04T08:00:00.000000"
		end            = "2021-08-04T17:00:00.000000"
		rrule          = "FREQ=DAILY;INTERVAL=1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateArchitectSchedulesResource( // Create Open schedule
					schedResource1,
					openSched,
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectSchedulesResource( // Create Closed schedule
					schedResource2,
					closedSched,
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectScheduleGroupsResource(
					schedGroupResource1,
					name,
					description,
					time_zone,
					generateSchedules("open_schedules_id", "genesyscloud_architect_schedules."+schedResource1+".id"),
					generateSchedules("closed_schedules_id", "genesyscloud_architect_schedules."+schedResource2+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResource1, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResource1, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResource1, "time_zone", time_zone),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResource1, "open_schedules_id.0", "genesyscloud_architect_schedules."+schedResource1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResource1, "closed_schedules_id.0", "genesyscloud_architect_schedules."+schedResource2, "id"),
					testDefaultHomeDivision("genesyscloud_architect_schedulegroups."+schedGroupResource1),
				),
			},
			{
				// Update to add Holiday Schedule
				Config: generateArchitectSchedulesResource( // Create Open schedule
					schedResource1,
					openSched,
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectSchedulesResource( // Create Closed schedule
					schedResource2,
					closedSched,
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectSchedulesResource( // Create Holiday schedule
					schedResource3,
					holidaySched,
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectScheduleGroupsResource(
					schedGroupResource1,
					name,
					description,
					time_zone,
					generateSchedules("open_schedules_id", "genesyscloud_architect_schedules."+schedResource1+".id"),
					generateSchedules("closed_schedules_id", "genesyscloud_architect_schedules."+schedResource2+".id"),
					generateSchedules("holiday_schedules_id", "genesyscloud_architect_schedules."+schedResource3+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResource1, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResource1, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResource1, "time_zone", time_zone),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResource1, "open_schedules_id.0", "genesyscloud_architect_schedules."+schedResource1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResource1, "closed_schedules_id.0", "genesyscloud_architect_schedules."+schedResource2, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResource1, "holiday_schedules_id.0", "genesyscloud_architect_schedules."+schedResource3, "id"),
					testDefaultHomeDivision("genesyscloud_architect_schedulegroups."+schedGroupResource1),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_architect_schedulegroups." + schedGroupResource1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"time_zone"},
			},
		},
		CheckDestroy: testVerifyScheduleGroupsDestroyed,
	})
}

func generateArchitectScheduleGroupsResource(
	schedGroupResource1 string,
	name string,
	description string,
	time_zone string,
	otherAttrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_architect_schedulegroups" "%s" {
		name = "%s"
		description = "%s"
		time_zone = "%s"
		%s
	}
	`, schedGroupResource1, name, description, time_zone, strings.Join(otherAttrs, "\n"))
}

func generateSchedules(
	propertyName string,
	scheduleIds ...string) string {
	return fmt.Sprintf(`
        %s = [%s]
	`, propertyName, strings.Join(scheduleIds, ","))
}

func testVerifyScheduleGroupsDestroyed(state *terraform.State) error {
	archAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_schedulegroups" {
			continue
		}

		schedGroup, resp, err := archAPI.GetArchitectSchedulegroup(rs.Primary.ID)
		if schedGroup != nil {
			return fmt.Errorf("Schedule group (%s) still exists", rs.Primary.ID)
		} else if isStatus404(resp) {
			// Schedule group not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All schedule groups destroyed
	return nil
}
