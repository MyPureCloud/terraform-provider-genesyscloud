package outbound_sequence

import (
	"fmt"
	"strconv"
	outboundCampaign "terraform-provider-genesyscloud/genesyscloud/outbound_campaign"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

func TestAccResourceOutboundSequence(t *testing.T) {
	t.Parallel()
	var (
		// Sequence
		sequenceResourceLabel = "outbound_sequence"
		sequenceName1         = "Sequence " + uuid.NewString()
		sequenceName2         = "Sequence " + uuid.NewString()

		// Campaign resources
		campaignResourceLabel    = "campaign_resource"
		campaignName             = "Campaign " + uuid.NewString()
		contactListResourceLabel = "contact_list"
		carResourceLabel         = "car"
		siteId                   = "site"
		outboundFlowFilePath     = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName                 = "test flow " + uuid.NewString()
		emergencyNumber          = "+13172947329"
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
				// Create
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
						"sequence-test-flow",
						flowName,
						"${data.genesyscloud_auth_division_home.home.name}",
						"sequence-test-location",
						"sequence-test-wrapupcode",
						divResourceLabel,
					) + GenerateOutboundSequence(
					sequenceResourceLabel,
					sequenceName1,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceLabel + ".id"},
					strconv.Quote("off"),
					util.TrueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "name", sequenceName1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "repeat", util.TrueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_sequence."+sequenceResourceLabel, "campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaignResourceLabel, "id"),
				),
			},
			{
				// Update with a new name, status and repeat value
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
						"sequence-test-flow",
						flowName,
						"${data.genesyscloud_auth_division_home.home.name}",
						"sequence-test-location",
						"sequence-test-wrapupcode",
						divResourceLabel,
					) + GenerateOutboundSequence(
					sequenceResourceLabel,
					sequenceName2,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceLabel + ".id"},
					strconv.Quote("on"),
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "name", sequenceName2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "status", "on"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "repeat", util.FalseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_sequence."+sequenceResourceLabel, "campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaignResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_sequence." + sequenceResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundSequenceDestroyed,
	})
}

func TestAccResourceOutboundSequenceStatus(t *testing.T) {
	t.Parallel()
	var (
		// Sequence
		sequenceResourceLabel = "outbound_sequence"
		sequenceName1         = "Sequence " + uuid.NewString()
		sequenceName2         = "Sequence " + uuid.NewString()

		// Campaign resources
		campaignResourceLabel    = "campaign_resource"
		campaignName             = "Campaign " + uuid.NewString()
		contactListResourceLabel = "contact_list"
		carResourceLabel         = "car"
		siteId                   = "site"
		outboundFlowFilePath     = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName                 = "test flow " + uuid.NewString()
		emergencyNumber          = "+13172947330"
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
				// Create
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
						"sequence-test-flow",
						flowName,
						"${data.genesyscloud_auth_division_home.home.name}",
						"sequence-test-location",
						"sequence-test-wrapupcode",
						divResourceLabel,
					) + GenerateOutboundSequence(
					sequenceResourceLabel,
					sequenceName1,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceLabel + ".id"},
					strconv.Quote("on"),
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "name", sequenceName1),
					util.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_sequence."+sequenceResourceLabel, "status", []string{"on", "complete"}),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "repeat", util.FalseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_sequence."+sequenceResourceLabel, "campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaignResourceLabel, "id"),
				),
			},
			{
				// Update with a new name, status and repeat value
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
						"sequence-test-flow",
						flowName,
						"${data.genesyscloud_auth_division_home.home.name}",
						"sequence-test-location",
						"sequence-test-wrapupcode",
						divResourceLabel,
					) + GenerateOutboundSequence(
					sequenceResourceLabel,
					sequenceName2,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceLabel + ".id"},
					strconv.Quote("off"),
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "name", sequenceName2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "repeat", util.FalseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_sequence."+sequenceResourceLabel, "campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaignResourceLabel, "id"),
				),
			},
			{
				// Turn back on to test that the sequence can be turned back on again, and ensure that the destroy
				// command can handle destroying a sequence that is "on"
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
						"sequence-test-flow",
						flowName,
						"${data.genesyscloud_auth_division_home.home.name}",
						"sequence-test-location",
						"sequence-test-wrapupcode",
						divResourceLabel,
					) + GenerateOutboundSequence(
					sequenceResourceLabel,
					sequenceName2,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceLabel + ".id"},
					strconv.Quote("on"),
					util.FalseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "name", sequenceName2),
					util.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_sequence."+sequenceResourceLabel, "status", []string{"on", "complete"}),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResourceLabel, "repeat", util.FalseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_sequence."+sequenceResourceLabel, "campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaignResourceLabel, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_sequence." + sequenceResourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundSequenceDestroyed,
	})
}

func testVerifyOutboundSequenceDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_sequence" {
			continue
		}
		sequence, resp, err := outboundAPI.GetOutboundSequence(rs.Primary.ID)
		if sequence != nil {
			return fmt.Errorf("sequence (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Sequence not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All sequences destroyed
	return nil
}
