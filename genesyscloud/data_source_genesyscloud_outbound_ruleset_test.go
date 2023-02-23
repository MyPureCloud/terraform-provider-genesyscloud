package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundRuleset(t *testing.T) {
	t.Parallel()
	var (
		ruleSetResourceId   = "rule-set-resource"
		ruleSetDataSourceId = "rule-set-data-source"
		ruleSetName         = "Test Rule Set " + uuid.NewString()
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
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
