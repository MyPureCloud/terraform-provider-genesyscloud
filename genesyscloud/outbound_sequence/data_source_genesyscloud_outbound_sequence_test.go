package outbound_sequence

import (
	"fmt"
	outboundCampaign "terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundSequence(t *testing.T) {

	t.Parallel()
	var (
		resourceId   = "sequence"
		dataSourceId = "sequence_data"
		sequenceName = "Test Campaign " + uuid.NewString()

		// Campaign
		campaignResourceId    = "campaign_resource"
		campaignName          = "Campaign " + uuid.NewString()
		contactListResourceId = "contact_list"
		carResourceId         = "car"
		siteId                = "site"
		outboundFlowFilePath  = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName              = "test flow " + uuid.NewString()
		emergencyNumber       = "+13128451429"
	)

	if err := edgeSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + outboundCampaign.GenerateOutboundCampaignBasic(
					campaignResourceId,
					campaignName,
					contactListResourceId,
					siteId,
					emergencyNumber,
					carResourceId,
					util.NullValue,
					outboundFlowFilePath,
					"data-sequence-test-flow",
					flowName,
					"${data.genesyscloud_auth_division_home.home.name}",
					"data-sequence-test-location",
					"data-sequence-test-wrapupcode",
				) + GenerateOutboundSequence(
					resourceId,
					sequenceName,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceId + ".id"},
					util.NullValue,
					util.NullValue,
				) + generateOutboundSequenceDataSource(
					dataSourceId,
					sequenceName,
					"genesyscloud_outbound_sequence."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_outbound_sequence."+dataSourceId, "id",
						"genesyscloud_outbound_sequence."+resourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundSequenceDataSource(
	id string,
	name string,
	dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_outbound_sequence" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, id, name, dependsOn)
}
