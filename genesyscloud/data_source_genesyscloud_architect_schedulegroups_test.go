package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/ronanwatkins/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceArchitectScheduleGroups(t *testing.T) {
	var (
		schedGroupResource = "arch-sched-group"
		name               = "Schedule Group x"
		description        = "Sample Schedule Group by CX as Code"
		time_zone          = "Asia/Singapore"

		schedGroupDataSource = "arch-sched-group-ds"

		schedResource = "arch-sched"
		openSched     = "Open Schedule"
		schedDesc     = "Sample Schedule by CX as Code"
		start         = "2021-08-04T08:00:00.000000"
		end           = "2021-08-04T17:00:00.000000"
		rrule         = "FREQ=DAILY;INTERVAL=1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateArchitectSchedulesResource( // Create Open schedule
					schedResource,
					openSched,
					schedDesc,
					start,
					end,
					rrule,
				) + generateArchitectScheduleGroupsResource(
					schedGroupResource,
					name,
					description,
					time_zone,
					generateSchedules("open_schedules_id", "genesyscloud_architect_schedules."+schedResource+".id"),
				) + generateScheduleGroupDataSource(
					schedGroupDataSource,
					"genesyscloud_architect_schedulegroups."+schedGroupResource+".name",
					"genesyscloud_architect_schedulegroups."+schedGroupResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_architect_schedulegroups."+schedGroupDataSource, "id", "genesyscloud_architect_schedulegroups."+schedGroupResource, "id"),
				),
			},
		},
	})
}

func generateScheduleGroupDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_architect_schedulegroups" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
