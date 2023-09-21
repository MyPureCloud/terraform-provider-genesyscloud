package outbound

import (
	"fmt"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

var (
	sdkConfig *platformclientv2.Configuration
)

func TestAccDataSourceOutboundCampaign(t *testing.T) {
	var (
		resourceId           = "campaign"
		campaignName         = "Test Campaign " + uuid.NewString()
		dataSourceId         = "campaign_data"
		outboundFlowFilePath = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
	)

	// necessary to avoid errors during site creation
	_, err := gcloud.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	emergencyNumber := "+13173124740"
	err = gcloud.DeleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
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
