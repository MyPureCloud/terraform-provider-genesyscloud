package architect_schedulegroups

import (
	"fmt"
	"strings"
	architectSchedules "terraform-provider-genesyscloud/genesyscloud/architect_schedules"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccResourceArchitectScheduleGroups(t *testing.T) {
	var (
		schedGroupResourceLabel1 = "arch-sched-group1"
		name                     = "Schedule Group " + uuid.NewString()
		description              = "Sample Schedule Group by CX as Code"
		time_zone                = "Asia/Singapore"

		schedResourceLabel1 = "arch-sched1"
		schedResourceLabel2 = "arch-sched2"
		schedResourceLabel3 = "arch-sched3"
		schedResourceLabel4 = "arch-sched4"
		schedResourceLabel5 = "arch-sched5"
		openSched           = "Open Schedule" + uuid.NewString()
		closedSched         = "Closed Schedule" + uuid.NewString()
		holidaySched        = "Holiday Schedule" + uuid.NewString()
		schedDesc           = "Sample Schedule by CX as Code"
		start               = "2021-08-04T08:00:00.000000"
		end                 = "2021-08-04T17:00:00.000000"
		rrule               = "FREQ=DAILY;INTERVAL=1"

		schedGroupResourceLabel2 = "arch-sched-group2"
		name2                    = "Schedule Group " + uuid.NewString()
		openSched2               = "Open Schedule 2 " + uuid.NewString()
		closedSched2             = "Closed Schedule 2 " + uuid.NewString()

		divResourceLabel = "test-division"
		divName          = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: architectSchedules.GenerateArchitectSchedulesResource( // Create Open schedule
					schedResourceLabel1,
					openSched,
					util.NullValue,
					schedDesc,
					start,
					end,
					rrule,
				) + architectSchedules.GenerateArchitectSchedulesResource( // Create Closed schedule
					schedResourceLabel2,
					closedSched,
					util.NullValue,
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectScheduleGroupsResource(
					schedGroupResourceLabel1,
					name,
					util.NullValue,
					description,
					time_zone,
					generateSchedules("open_schedules_id", "genesyscloud_architect_schedules."+schedResourceLabel1+".id"),
					generateSchedules("closed_schedules_id", "genesyscloud_architect_schedules."+schedResourceLabel2+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "time_zone", time_zone),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "open_schedules_id.0", "genesyscloud_architect_schedules."+schedResourceLabel1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "closed_schedules_id.0", "genesyscloud_architect_schedules."+schedResourceLabel2, "id"),
					provider.TestDefaultHomeDivision("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1),
				),
			},
			{
				// Update to add Holiday Schedule
				Config: architectSchedules.GenerateArchitectSchedulesResource( // Create Open schedule
					schedResourceLabel1,
					openSched,
					util.NullValue,
					schedDesc,
					start,
					end,
					rrule,
				) + architectSchedules.GenerateArchitectSchedulesResource( // Create Closed schedule
					schedResourceLabel2,
					closedSched,
					util.NullValue,
					schedDesc,
					start,
					end,
					rrule,
				) + architectSchedules.GenerateArchitectSchedulesResource( // Create Holiday schedule
					schedResourceLabel3,
					holidaySched,
					util.NullValue,
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectScheduleGroupsResource(
					schedGroupResourceLabel1,
					name,
					util.NullValue,
					description,
					time_zone,
					generateSchedules("open_schedules_id", "genesyscloud_architect_schedules."+schedResourceLabel1+".id"),
					generateSchedules("closed_schedules_id", "genesyscloud_architect_schedules."+schedResourceLabel2+".id"),
					generateSchedules("holiday_schedules_id", "genesyscloud_architect_schedules."+schedResourceLabel3+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "time_zone", time_zone),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "open_schedules_id.0", "genesyscloud_architect_schedules."+schedResourceLabel1, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "closed_schedules_id.0", "genesyscloud_architect_schedules."+schedResourceLabel2, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1, "holiday_schedules_id.0", "genesyscloud_architect_schedules."+schedResourceLabel3, "id"),
					provider.TestDefaultHomeDivision("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel1),
				),
			},
			{
				// Create with new division
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + architectSchedules.GenerateArchitectSchedulesResource( // Create Open schedule
					schedResourceLabel4,
					openSched2,
					"genesyscloud_auth_division."+divResourceLabel+".id",
					schedDesc,
					start,
					end,
					rrule,
				) + architectSchedules.GenerateArchitectSchedulesResource( // Create Closed schedule
					schedResourceLabel5,
					closedSched2,
					"genesyscloud_auth_division."+divResourceLabel+".id",
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectScheduleGroupsResource(
					schedGroupResourceLabel2,
					name2,
					"genesyscloud_auth_division."+divResourceLabel+".id",
					description,
					time_zone,
					generateSchedules("open_schedules_id", "genesyscloud_architect_schedules."+schedResourceLabel4+".id"),
					generateSchedules("closed_schedules_id", "genesyscloud_architect_schedules."+schedResourceLabel5+".id"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel2, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel2, "description", description),
					resource.TestCheckResourceAttr("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel2, "time_zone", time_zone),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel2, "open_schedules_id.0", "genesyscloud_architect_schedules."+schedResourceLabel4, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel2, "closed_schedules_id.0", "genesyscloud_architect_schedules."+schedResourceLabel5, "id"),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_schedulegroups."+schedGroupResourceLabel2, "division_id", "genesyscloud_auth_division."+divResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_architect_schedulegroups." + schedGroupResourceLabel2,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"time_zone"},
			},
		},
		CheckDestroy: testVerifyScheduleGroupsDestroyed,
	})
}

func generateArchitectScheduleGroupsResource(
	resourceLabel string,
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
	`, resourceLabel, name, divisionId, description, time_zone, strings.Join(otherAttrs, "\n"))
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
