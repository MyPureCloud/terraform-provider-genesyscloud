package outbound

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	edgeSite "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_site"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceOutboundSequence(t *testing.T) {

	t.Parallel()
	var (
		// Sequence
		sequenceResource = "outbound_sequence"
		sequenceName1    = "Sequence " + uuid.NewString()
		sequenceName2    = "Sequence " + uuid.NewString()

		// Campaign resources
		campaignResourceId    = "campaign_resource"
		campaignName          = "Campaign " + uuid.NewString()
		contactListResourceId = "contact_list"
		carResourceId         = "car"
		siteId                = "site"
		outboundFlowFilePath  = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName              = "test flow " + uuid.NewString()
		emergencyNumber       = "+13172947329"
	)

	// necessary to avoid errors during site creation
	_, err := gcloud.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	err = edgeSite.DeleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateOutboundCampaignBasic(
					campaignResourceId,
					campaignName,
					contactListResourceId,
					siteId,
					emergencyNumber,
					carResourceId,
					nullValue,
					outboundFlowFilePath,
					"sequence-test-flow",
					flowName,
					"${data.genesyscloud_auth_division_home.home.name}",
					"sequence-test-location",
					"sequence-test-wrapupcode",
				) + generateOutboundSequence(
					sequenceResource,
					sequenceName1,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceId + ".id"},
					strconv.Quote("off"),
					trueValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "name", sequenceName1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "repeat", trueValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_sequence."+sequenceResource, "campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaignResourceId, "id"),
				),
			},
			{
				// Update with a new name, status and repeat value
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateOutboundCampaignBasic(
					campaignResourceId,
					campaignName,
					contactListResourceId,
					siteId,
					emergencyNumber,
					carResourceId,
					nullValue,
					outboundFlowFilePath,
					"sequence-test-flow",
					flowName,
					"${data.genesyscloud_auth_division_home.home.name}",
					"sequence-test-location",
					"sequence-test-wrapupcode",
				) + generateOutboundSequence(
					sequenceResource,
					sequenceName2,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceId + ".id"},
					strconv.Quote("on"),
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "name", sequenceName2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "status", "on"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "repeat", falseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_sequence."+sequenceResource, "campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaignResourceId, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_sequence." + sequenceResource,
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
		sequenceResource = "outbound_sequence"
		sequenceName1    = "Sequence " + uuid.NewString()
		sequenceName2    = "Sequence " + uuid.NewString()

		// Campaign resources
		campaignResourceId    = "campaign_resource"
		campaignName          = "Campaign " + uuid.NewString()
		contactListResourceId = "contact_list"
		carResourceId         = "car"
		siteId                = "site"
		outboundFlowFilePath  = "../../examples/resources/genesyscloud_flow/outboundcall_flow_example.yaml"
		flowName              = "test flow " + uuid.NewString()
		emergencyNumber       = "+13172947330"
	)

	// necessary to avoid errors during site creation
	_, err := gcloud.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}

	err = edgeSite.DeleteLocationWithNumber(emergencyNumber)
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateOutboundCampaignBasic(
					campaignResourceId,
					campaignName,
					contactListResourceId,
					siteId,
					emergencyNumber,
					carResourceId,
					nullValue,
					outboundFlowFilePath,
					"sequence-test-flow",
					flowName,
					"${data.genesyscloud_auth_division_home.home.name}",
					"sequence-test-location",
					"sequence-test-wrapupcode",
				) + generateOutboundSequence(
					sequenceResource,
					sequenceName1,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceId + ".id"},
					strconv.Quote("on"),
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "name", sequenceName1),
					gcloud.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_sequence."+sequenceResource, "status", []string{"on", "complete"}),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "repeat", falseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_sequence."+sequenceResource, "campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaignResourceId, "id"),
				),
			},
			{
				// Update with a new name, status and repeat value
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateOutboundCampaignBasic(
					campaignResourceId,
					campaignName,
					contactListResourceId,
					siteId,
					emergencyNumber,
					carResourceId,
					nullValue,
					outboundFlowFilePath,
					"sequence-test-flow",
					flowName,
					"${data.genesyscloud_auth_division_home.home.name}",
					"sequence-test-location",
					"sequence-test-wrapupcode",
				) + generateOutboundSequence(
					sequenceResource,
					sequenceName2,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceId + ".id"},
					strconv.Quote("off"),
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "name", sequenceName2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "status", "off"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "repeat", falseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_sequence."+sequenceResource, "campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaignResourceId, "id"),
				),
			},
			{
				// Turn back on to test that the sequence can be turned back on again, and ensure that the destroy
				// command can handle destroying a sequence that is "on"
				Config: fmt.Sprintf(`
data "genesyscloud_auth_division_home" "home" {}
`) + generateOutboundCampaignBasic(
					campaignResourceId,
					campaignName,
					contactListResourceId,
					siteId,
					emergencyNumber,
					carResourceId,
					nullValue,
					outboundFlowFilePath,
					"sequence-test-flow",
					flowName,
					"${data.genesyscloud_auth_division_home.home.name}",
					"sequence-test-location",
					"sequence-test-wrapupcode",
				) + generateOutboundSequence(
					sequenceResource,
					sequenceName2,
					[]string{"genesyscloud_outbound_campaign." + campaignResourceId + ".id"},
					strconv.Quote("on"),
					falseValue,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "name", sequenceName2),
					gcloud.VerifyAttributeInArrayOfPotentialValues("genesyscloud_outbound_sequence."+sequenceResource, "status", []string{"on", "complete"}),
					resource.TestCheckResourceAttr("genesyscloud_outbound_sequence."+sequenceResource, "repeat", falseValue),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_sequence."+sequenceResource, "campaign_ids.0",
						"genesyscloud_outbound_campaign."+campaignResourceId, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_sequence." + sequenceResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundSequenceDestroyed,
	})
}

func generateOutboundSequence(
	resourceId string,
	name string,
	campaignIds []string,
	status string,
	repeat string) string {
	return fmt.Sprintf(`
		resource "genesyscloud_outbound_sequence" "%s" {
			name = "%s"
			campaign_ids = [%s]
			status = %s
			repeat = %s
		}
	`, resourceId, name, strings.Join(campaignIds, ", "), status, repeat)
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
		} else if gcloud.IsStatus404(resp) {
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
