package outbound_campaign

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundCampaign(t *testing.T) {
	var (
		resourceId           = "campaign"
		campaignName         = "Test Campaign " + uuid.NewString()
		dataSourceId         = "campaign_data"
		outboundFlowFilePath = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
	)

	emergencyNumber := "+13173124740"
	if err := edgeSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s, %v", emergencyNumber, err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: `data "genesyscloud_auth_division_home" "home" {}` + GenerateOutboundCampaignBasic(
					resourceId,
					campaignName,
					"contact_list",
					"site",
					emergencyNumber,
					"car",
					util.NullValue,
					outboundFlowFilePath,
					"data-campaign-test-flow",
					"test flow "+uuid.NewString(),
					"${data.genesyscloud_auth_division_home.home.name}",
					"data-campaign-test-location",
					"data-campaign-test-wrapupcode",
				) + generateOutboundCampaignDataSource(
					dataSourceId,
					campaignName,
					"genesyscloud_outbound_campaign."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_campaign."+dataSourceId, "id",
						"genesyscloud_outbound_campaign."+resourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundCampaignDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_campaign" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, id, name, dependsOn)
}
