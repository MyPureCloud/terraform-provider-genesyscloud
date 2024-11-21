package outbound_sequence

import (
	"fmt"
	outboundCampaign "terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundSequence(t *testing.T) {

	t.Parallel()
	var (
		resourceLabel   = "sequence"
		dataSourceLabel = "sequence_data"
		sequenceName    = "Test Campaign " + uuid.NewString()

		// Campaign
		campaignResourceLabel    = "campaign_resource"
		campaignName             = "Campaign " + uuid.NewString()
		contactListResourceLabel = "contact_list"
		carResourceLabel         = "car"
		siteId                   = "site"
		outboundFlowFilePath     = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName                 = "test flow " + uuid.NewString()
		emergencyNumber          = "+13128451429"
		divResourceLabel         = "test-outbound-sequence-division"
		divName                  = "terraform-" + uuid.NewString()
	)

	if err := edgeSite.DeleteLocationWithNumber(emergencyNumber, sdkConfig); err != nil {
		t.Skipf("failed to delete location with number %s: %v", emergencyNumber, err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: `data "genesyscloud_auth_division_home" "home" {}` + "\n" +
					authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					outboundCampaign.GenerateOutboundCampaignBasic(
						campaignResourceLabel,
						campaignName,
						contactListResourceLabel,
						siteId,
						emergencyNumber,
						carResourceLabel,
						util.NullValue,
						outboundFlowFilePath,
						"data-sequence-test-flow",
						flowName,
						"${data.genesyscloud_auth_division_home.home.name}",
						"data-sequence-test-location",
						"data-sequence-test-wrapupcode",
						divResourceLabel,
					) + GenerateOutboundSequence(
					resourceLabel,
					sequenceName,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceLabel + ".id"},
					util.NullValue,
					util.NullValue,
				) + generateOutboundSequenceDataSource(
					dataSourceLabel,
					sequenceName,
					"genesyscloud_outbound_sequence."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_outbound_sequence."+dataSourceLabel, "id",
						"genesyscloud_outbound_sequence."+resourceLabel, "id"),
				),
			},
		},
	})
}

func generateOutboundSequenceDataSource(
	dataSourceLabel string,
	name string,
	dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_outbound_sequence" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, dataSourceLabel, name, dependsOn)
}
