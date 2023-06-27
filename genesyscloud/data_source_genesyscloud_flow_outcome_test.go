package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFlowOutcome(t *testing.T) {
	t.Skip("Skipping until a DELETE method is publicly available for flow outcomes.")
	var (
		outcomeRes  = "flow-outcome"
		outcomeData = "outcomeData"
		name        = "Terraform Code-" + uuid.NewString()
		description = "Sample Outcome by CX as Code"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateFlowOutcomeResource(
					outcomeRes,
					name,
					nullValue,
					description,
				) + generateFlowOutcomeDataSource(
					outcomeData,
					name,
					"genesyscloud_flow_outcome."+outcomeRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_flow_outcome."+outcomeData, "id", "genesyscloud_flow_outcome."+outcomeRes, "id"),
				),
			},
		},
	})
}

func generateFlowOutcomeDataSource(resourceID string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_flow_outcome" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
