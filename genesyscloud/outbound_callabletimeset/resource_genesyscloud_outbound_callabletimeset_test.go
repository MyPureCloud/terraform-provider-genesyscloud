package outbound_callabletimeset

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

func TestAccResourceOutboundCallabletimeset(t *testing.T) {

	var (
		resourceLabel = "callable-time-set"
		name1         = "Test Callable time set" + uuid.NewString()
		timeZone1     = "Africa/Abidjan"
		timeZone2     = "Europe/Dublin"

		name2 = "Test Callable time set" + uuid.NewString()

		timeBlock1 = GenerateCallableTimesBlock(
			timeZone1,
			GenerateTimeSlotsBlock("07:00:00", "18:00:00", "3"),
			GenerateTimeSlotsBlock("09:30:00", "22:30:00", "5"),
		)
		timeBlock2 = GenerateCallableTimesBlock(
			timeZone2,
			GenerateTimeSlotsBlock("05:30:30", "14:45:00", "1"),
			GenerateTimeSlotsBlock("10:15:45", "20:30:00", "6"),
		)
		timeBlock3 = GenerateCallableTimesBlock(
			timeZone1,
			GenerateTimeSlotsBlock("09:00:00", "21:30:30", "1"),
			GenerateTimeSlotsBlock("10:30:45", "23:00:15", "7"),
		)
		timeBlock4 = GenerateCallableTimesBlock(
			timeZone2,
			GenerateTimeSlotsBlock("08:15:15", "20:30:45", "2"),
			GenerateTimeSlotsBlock("01:00:00", "12:00:00", "4"),
		)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateOutboundCallabletimeset(
					resourceLabel,
					name1,
					timeBlock1,
					timeBlock2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceLabel, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceLabel, "callable_times.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("genesyscloud_outbound_callabletimeset."+resourceLabel, "callable_times.*", map[string]string{
						"time_zone_id": timeZone1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("genesyscloud_outbound_callabletimeset."+resourceLabel, "callable_times.*", map[string]string{
						"time_zone_id": timeZone2,
					}),
				),
			},
			{
				// Update with new name and callable times time slots
				Config: GenerateOutboundCallabletimeset(
					resourceLabel,
					name2,
					timeBlock3,
					timeBlock4,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceLabel, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceLabel, "callable_times.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("genesyscloud_outbound_callabletimeset."+resourceLabel, "callable_times.*", map[string]string{
						"time_zone_id": timeZone1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("genesyscloud_outbound_callabletimeset."+resourceLabel, "callable_times.*", map[string]string{
						"time_zone_id": timeZone2,
					}),
				),
			},
			{
				// Update with named callable times
				Config: GenerateOutboundCallabletimeset(
					resourceLabel,
					name1,
					GenerateCallableTimesBlock(
						timeZone1,
						`name = "America-NewYork-Morning"`,
						GenerateTimeSlotsBlock("09:00:00", "12:00:00", "1"),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"genesyscloud_outbound_callabletimeset."+resourceLabel, "name", name1,
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"genesyscloud_outbound_callabletimeset."+resourceLabel,
						"callable_times.*",
						map[string]string{
							"name":         "America-NewYork-Morning",
							"time_zone_id": timeZone1,
						},
					),
				),
			},
			{
				// Read
				ResourceName:      "genesyscloud_outbound_callabletimeset." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyCallabletimesetDestroyed,
	})
}

func testVerifyCallabletimesetDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_callabletimeset" {
			continue
		}
		timeSet, resp, err := outboundAPI.GetOutboundCallabletimeset(rs.Primary.ID)
		if timeSet != nil {
			return fmt.Errorf("Callable time set (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Callable time set not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All callable time sets deleted
	return nil
}
