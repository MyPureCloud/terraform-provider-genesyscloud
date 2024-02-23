package outbound_ruleset

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Outbound ruleset Data Source
*/
func TestAccDataSourceOutboundRuleset(t *testing.T) {
	t.Parallel()
	var (
		ruleSetResourceId   = "rule-set-resource"
		ruleSetDataSourceId = "rule-set-data-source"
		ruleSetName         = "Test Rule Set " + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
}
`, ruleSetResourceId, ruleSetName) + fmt.Sprintf(`data "genesyscloud_outbound_ruleset" "%s" {
  name            = "%s"
  depends_on      = [genesyscloud_outbound_ruleset.%s]
}
`, ruleSetDataSourceId, ruleSetName, ruleSetResourceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_ruleset."+ruleSetDataSourceId, "id", "genesyscloud_outbound_ruleset."+ruleSetResourceId, "id"),
				),
			},
		},
	})
}
