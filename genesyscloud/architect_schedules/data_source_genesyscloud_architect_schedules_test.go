package architect_schedules

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceArchitectSchedule(t *testing.T) {
	var (
		schedRes    = "arch-sched1"
		schedData   = "schedData"
		name        = "CX as Code Schedule" + uuid.NewString()
		description = "Sample Schedule by CX as Code"
		start       = "2021-08-04T08:00:00.000000"
		end         = "2021-08-04T17:00:00.000000"
		rrule       = "FREQ=DAILY;INTERVAL=1"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateArchitectSchedulesResource(
					schedRes,
					name,
					util.NullValue,
					description,
					start,
					end,
					rrule,
				) + generateScheduleDataSource(
					schedData,
					name,
					"genesyscloud_architect_schedules."+schedRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_architect_schedules."+schedData, "id", "genesyscloud_architect_schedules."+schedRes, "id"),
				),
			},
		},
	})
}

func generateScheduleDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_architect_schedules" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
