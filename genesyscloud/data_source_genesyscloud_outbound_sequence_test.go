package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceOutboundSequence(t *testing.T) {
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
		outboundFlowFilePath  = "../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName              = "test flow " + uuid.NewString()
		emergencyNumber       = "+13128451429"

		campaignResource = generateOutboundCampaignBasic(
			campaignResourceId,
			campaignName,
			contactListResourceId,
			siteId,
			emergencyNumber,
			carResourceId,
			nullValue,
			outboundFlowFilePath,
			flowName,
		)
	)

	// necessary to avoid errors during site creation
	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	err = deleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: campaignResource +
					generateOutboundSequence(
						resourceId,
						sequenceName,
						[]string{"genesyscloud_outbound_campaign." + campaignResourceId + ".id"},
						nullValue,
						nullValue,
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
