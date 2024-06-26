package architect_schedulegroups

import (
	"fmt"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	architectSchedules "terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceArchitectScheduleGroups(t *testing.T) {
	var (
		schedGroupResource1 = "arch-sched-group1"
		name                = "Schedule Group " + uuid.NewString()
		description         = "Sample Schedule Group by CX as Code"
		time_zone           = "Asia/Singapore"

		schedResource1 = "arch-sched1"
		schedResource2 = "arch-sched2"
		schedResource3 = "arch-sched3"
		schedResource4 = "arch-sched4"
		schedResource5 = "arch-sched5"
		openSched      = "Open Schedule" + uuid.NewString()
		closedSched    = "Closed Schedule" + uuid.NewString()
		holidaySched   = "Holiday Schedule" + uuid.NewString()
		schedDesc      = "Sample Schedule by CX as Code"
		start          = "2021-08-04T08:00:00.000000"
		end            = "2021-08-04T17:00:00.000000"
		rrule          = "FREQ=DAILY;INTERVAL=1"

		schedGroupResource2 = "arch-sched-group2"
		name2               = "Schedule Group " + uuid.NewString()
		openSched2          = "Open Schedule 2 " + uuid.NewString()
		closedSched2        = "Closed Schedule 2 " + uuid.NewString()

		divResource = "test-division"
		divName     = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: architectSchedules.GenerateArchitectSchedulesResource( // Create Open schedule
					schedResource1,
					openSched,
					util.NullValue,
					schedDesc,
					start,
					end,
					rrule,
				) + architectSchedules.GenerateArchitectSchedulesResource( // Create Closed schedule
					schedResource2,
					closedSched,
					util.NullValue,
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectScheduleGroupsResource(
					schedGroupResource1,
					name,
					util.NullValue,
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
					provider.TestDefaultHomeDivision("genesyscloud_architect_schedulegroups."+schedGroupResource1),
				),
			},
			{
				// Update to add Holiday Schedule
				Config: architectSchedules.GenerateArchitectSchedulesResource( // Create Open schedule
					schedResource1,
					openSched,
					util.NullValue,
					schedDesc,
					start,
					end,
					rrule,
				) + architectSchedules.GenerateArchitectSchedulesResource( // Create Closed schedule
					schedResource2,
					closedSched,
					util.NullValue,
					schedDesc,
					start,
					end,
					rrule,
				) + architectSchedules.GenerateArchitectSchedulesResource( // Create Holiday schedule
					schedResource3,
					holidaySched,
					util.NullValue,
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectScheduleGroupsResource(
					schedGroupResource1,
					name,
					util.NullValue,
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
					provider.TestDefaultHomeDivision("genesyscloud_architect_schedulegroups."+schedGroupResource1),
				),
			},
			{
				// Create with new division
				Config: gcloud.GenerateAuthDivisionBasic(divResource, divName) + architectSchedules.GenerateArchitectSchedulesResource( // Create Open schedule
					schedResource4,
					openSched2,
					"genesyscloud_auth_division."+divResource+".id",
					schedDesc,
					start,
					end,
					rrule,
				) + architectSchedules.GenerateArchitectSchedulesResource( // Create Closed schedule
					schedResource5,
					closedSched2,
					"genesyscloud_auth_division."+divResource+".id",
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectScheduleGroupsResource(
					schedGroupResource2,
					name2,
					"genesyscloud_auth_division."+divResource+".id",
					description,
					time_zone,
					generateSchedules("open_schedules_id", "genesyscloud_architect_schedules."+schedResource4+".id"),
					generateSchedules("closed_schedules_id", "genesyscloud_architect_schedules."+schedResource5+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResource2, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResource2, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResource2, "time_zone", time_zone),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResource2, "open_schedules_id.0", "genesyscloud_architect_schedules."+schedResource4, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResource2, "closed_schedules_id.0", "genesyscloud_architect_schedules."+schedResource5, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResource2, "division_id", "genesyscloud_auth_division."+divResource, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_architect_schedulegroups." + schedGroupResource2,
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
	divisionId string,
	description string,
	time_zone string,
	otherAttrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_architect_schedulegroups" "%s" {
		name = "%s"
		division_id = %s
		description = "%s"
		time_zone = "%s"
		%s
	}
	`, schedGroupResource1, name, divisionId, description, time_zone, strings.Join(otherAttrs, "\n"))
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
		} else if util.IsStatus404(resp) {
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
