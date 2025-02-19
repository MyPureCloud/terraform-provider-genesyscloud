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
		schedResourceLabel = "arch-sched1"
		schedDataLabel     = "schedData"
		name               = "CX as Code Schedule" + uuid.NewString()
		description        = "Sample Schedule by CX as Code"
		start              = "2021-08-04T08:00:00.000000"
		end                = "2021-08-04T17:00:00.000000"
		rrule              = "FREQ=DAILY;INTERVAL=1"

		resourcePath     = ResourceType + "." + schedResourceLabel
		dataResourcePath = "data." + ResourceType + "." + schedDataLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateArchitectSchedulesResource(
					schedResourceLabel,
					name,
					util.NullValue,
					description,
					start,
					end,
					rrule,
				) + generateScheduleDataSource(
					schedDataLabel,
					name,
					resourcePath),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataResourcePath, "id", resourcePath, "id"),
				),
			},
		},
	})
}

func generateScheduleDataSource(
	resourceLabel,
	name,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, resourceLabel, name, dependsOnResource)
}
