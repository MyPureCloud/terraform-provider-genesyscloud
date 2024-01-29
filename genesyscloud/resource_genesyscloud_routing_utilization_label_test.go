package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

func TestAccResourceRoutingUtilizationLabelBasic(t *testing.T) {
	var (
		resourceName     = "test-label"
		labelName        = "Terraform Label " + uuid.NewString()
		updatedLabelName = "Updated " + labelName
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			TestAccPreCheck(t)
			if err := checkIfLabelsAreEnabled(); err != nil {
				t.Skipf("%v", err) // be sure to skip the test and not fail it
			}
		},
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateRoutingUtilizationLabelResource(
					resourceName,
					labelName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization_label."+resourceName, "name", labelName),
				),
			},
			{
				// Update
				Config: GenerateRoutingUtilizationLabelResource(
					resourceName,
					updatedLabelName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_routing_utilization_label."+resourceName, "name", updatedLabelName),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_utilization_label." + resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: validateTestLabelDestroyed,
	})
}

func GenerateRoutingUtilizationLabelResource(resourceID string, name string, dependsOnResource string) string {
	dependsOn := ""

	if dependsOnResource != "" {
		dependsOn = fmt.Sprintf("depends_on=[genesyscloud_routing_utilization_label.%s]", dependsOnResource)
	}

	return fmt.Sprintf(`resource "genesyscloud_routing_utilization_label" "%s" {
		name = "%s"
		%s
	}
	`, resourceID, name, dependsOn)
}

func validateTestLabelDestroyed(state *terraform.State) error {
	routingApi := platformclientv2.NewRoutingApi()

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_routing_utilization_label" {
			continue
		}

		_, resp, err := routingApi.GetRoutingUtilizationLabel(rs.Primary.ID)

		if IsStatus404(resp) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("Unexpected error: %s", err)
		}

		return fmt.Errorf("Label (%s) still exists", rs.Primary.ID)
	}

	return fmt.Errorf("No label resource found")
}

func checkIfLabelsAreEnabled() error { // remove once the feature is globally enabled
	api := platformclientv2.NewRoutingApiWithConfig(sdkConfig) // the variable sdkConfig exists at a package level in ./genesyscloud and is already authorized
	_, resp, _ := api.GetRoutingUtilizationLabels(100, 1, "", "")
	if resp.StatusCode == 501 {
		return fmt.Errorf("feature is not yet implemented in this org.")
	}
	return nil
}
