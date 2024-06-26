package outbound_callabletimeset

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceOutboundCallabletimeset(t *testing.T) {

	var (
		resourceId = "callable-time-set"
		name1      = "Test Callable time set" + uuid.NewString()
		timeZone1  = "Africa/Abidjan"
		timeZone2  = "Europe/Dublin"

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
					resourceId,
					name1,
					timeBlock1,
					timeBlock2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "name", name1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_zone_id", timeZone1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.0.start_time", "07:00:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.0.stop_time", "18:00:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.0.day", "3"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.1.start_time", "09:30:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.1.stop_time", "22:30:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.1.day", "5"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_zone_id", timeZone2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.0.start_time", "05:30:30"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.0.stop_time", "14:45:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.0.day", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.1.start_time", "10:15:45"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.1.stop_time", "20:30:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.1.day", "6"),
				),
			},
			{
				// Update with new name and callable times time slots
				Config: GenerateOutboundCallabletimeset(
					resourceId,
					name2,
					timeBlock3,
					timeBlock4,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "name", name2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_zone_id", timeZone1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.0.start_time", "09:00:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.0.stop_time", "21:30:30"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.0.day", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.1.start_time", "10:30:45"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.1.stop_time", "23:00:15"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.0.time_slots.1.day", "7"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_zone_id", timeZone2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.0.start_time", "08:15:15"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.0.stop_time", "20:30:45"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.0.day", "2"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.1.start_time", "01:00:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.1.stop_time", "12:00:00"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_callabletimeset."+resourceId, "callable_times.1.time_slots.1.day", "4"),
				),
			},
			{
				// Read
				ResourceName:      "genesyscloud_outbound_callabletimeset." + resourceId,
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
