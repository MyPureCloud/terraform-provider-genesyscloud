package routing_utilization_label

import (
	"fmt"
	"regexp"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccResourceRoutingUtilizationLabelBasic(t *testing.T) {
	var (
		resourceLabel    = "test-label"
		labelName        = "Terraform Label " + uuid.NewString()
		updatedLabelName = "Updated " + labelName
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateRoutingUtilizationLabelResource(
					resourceLabel,
					labelName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization_label."+resourceLabel, "name", labelName),
				),
			},
			{
				// Update
				Config: generateRoutingUtilizationLabelResource(
					resourceLabel,
					updatedLabelName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization_label."+resourceLabel, "name", updatedLabelName),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_utilization_label." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: validateTestLabelDestroyed,
	})
}

func TestAccResourceRoutingUtilizationLabelInvalidNames(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config:      GenerateRoutingUtilizationLabelResource("resource", " abc", ""),
				ExpectError: regexp.MustCompile("to not start or end with spaces"),
			},
			{
				Config:      GenerateRoutingUtilizationLabelResource("resource", "abc ", ""),
				ExpectError: regexp.MustCompile("to not start or end with spaces"),
			},
			{
				Config:      GenerateRoutingUtilizationLabelResource("resource", " abc ", ""),
				ExpectError: regexp.MustCompile("to not start or end with spaces"),
			},
			{
				Config:      GenerateRoutingUtilizationLabelResource("resource", "abc*", ""),
				ExpectError: regexp.MustCompile("expected value of name to not contain any of"),
			},
			{
				Config:      GenerateRoutingUtilizationLabelResource("resource", "", ""),
				ExpectError: regexp.MustCompile("to not be an empty string"),
			},
		},
		CheckDestroy: validateTestLabelDestroyed,
	})
}

func validateTestLabelDestroyed(state *terraform.State) error {
	routingApi := platformclientv2.NewRoutingApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_utilization_label" {
			continue
		}

		_, resp, err := routingApi.GetRoutingUtilizationLabel(rs.Primary.ID)

		if util.IsStatus404(resp) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("Unexpected error: %s", err)
		}

		return fmt.Errorf("Label (%s) still exists", rs.Primary.ID)
	}

	return fmt.Errorf("No label resource found")
}
