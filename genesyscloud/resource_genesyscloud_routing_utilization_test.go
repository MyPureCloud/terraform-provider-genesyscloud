package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRoutingUtilization(t *testing.T) {
	t.Parallel()
	var (
		maxCapacity1  = "3"
		maxCapacity2  = "4"
		utilTypeCall  = "call"
		utilTypeEmail = "email"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingUtilizationResource(
					GenerateRoutingUtilMediaType("call", maxCapacity1, falseValue),
					GenerateRoutingUtilMediaType("callback", maxCapacity1, falseValue),
					GenerateRoutingUtilMediaType("chat", maxCapacity1, falseValue),
					GenerateRoutingUtilMediaType("email", maxCapacity1, falseValue),
					GenerateRoutingUtilMediaType("message", maxCapacity1, falseValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.include_non_acd", falseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.include_non_acd", falseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.include_non_acd", falseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.include_non_acd", falseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.interruptible_media_types"),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.maximum_capacity", maxCapacity1),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.include_non_acd", falseValue),
					resource.TestCheckNoResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.interruptible_media_types"),
				),
			},
			{
				// Update with a new max capacities and interruptible media types
				Config: generateRoutingUtilizationResource(
					GenerateRoutingUtilMediaType("call", maxCapacity2, trueValue, strconv.Quote(utilTypeEmail)),
					GenerateRoutingUtilMediaType("callback", maxCapacity2, trueValue, strconv.Quote(utilTypeCall)),
					GenerateRoutingUtilMediaType("chat", maxCapacity2, trueValue, strconv.Quote(utilTypeCall)),
					GenerateRoutingUtilMediaType("email", maxCapacity2, trueValue, strconv.Quote(utilTypeCall)),
					GenerateRoutingUtilMediaType("message", maxCapacity2, trueValue, strconv.Quote(utilTypeCall)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "call.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "call.0.interruptible_media_types", utilTypeEmail),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "callback.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "callback.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "chat.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "chat.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "email.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "email.0.interruptible_media_types", utilTypeCall),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.maximum_capacity", maxCapacity2),
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization.routing-util", "message.0.include_non_acd", trueValue),
					ValidateStringInArray("genesyscloud_routing_utilization.routing-util", "message.0.interruptible_media_types", utilTypeCall),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_utilization.routing-util",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateRoutingUtilizationResource(mediaTypes ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_utilization" "routing-util" {
		%s
	}
	`, strings.Join(mediaTypes, "\n"))
}
