package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundCampaign(t *testing.T) {
	var (
		resourceId           = "campaign"
		campaignName         = "Test Campaign " + uuid.NewString()
		dataSourceId         = "campaign_data"
		outboundFlowFilePath = "../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
	)
	// necessary to avoid errors during site creation
	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	emergencyNumber := "3173124740"
	err = deleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateOutboundCampaignBasic(
					resourceId,
					campaignName,
					"contact_list",
					"site",
					emergencyNumber,
					"car",
					nullValue,
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
