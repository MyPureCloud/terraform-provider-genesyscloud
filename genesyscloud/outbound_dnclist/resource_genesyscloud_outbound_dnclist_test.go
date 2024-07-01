package outbound_dnclist

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

const NullValue = "null"

func TestAccResourceOutboundDncListRdsListType(t *testing.T) {

	t.Parallel()
	var (
		resourceID    = "dnc_list"
		name          = "Test DNC List " + uuid.NewString()
		dncSourceType = "rds"
		contactMethod = "Phone"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateOutboundDncList(
					resourceID,
					name,
					dncSourceType,
					strconv.Quote(contactMethod),
					NullValue,
					NullValue,
					NullValue,
					[]string{},
					generateOutboundDncListEntriesBlock(
						[]string{strconv.Quote("+353747474747")},
						NullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "contact_method", contactMethod),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
					checkPhoneNumbersAddedToDncList("genesyscloud_outbound_dnclist."+resourceID, 1),
				),
			},
			{
				Config: generateOutboundDncList(
					resourceID,
					name,
					dncSourceType,
					strconv.Quote(contactMethod),
					NullValue,
					NullValue,
					NullValue,
					[]string{},
					generateOutboundDncListEntriesBlock(
						[]string{strconv.Quote("+353112222222"), strconv.Quote("+353221111111")},
						NullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "contact_method", contactMethod),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
					checkPhoneNumbersAddedToDncList("genesyscloud_outbound_dnclist."+resourceID, 3),
				),
			},
			{
				Config: generateOutboundDncList(
					resourceID,
					name,
					dncSourceType,
					strconv.Quote(contactMethod),
					NullValue,
					NullValue,
					NullValue,
					[]string{},
					generateOutboundDncListEntriesBlock(
						[]string{strconv.Quote("+353112222222"), strconv.Quote("+353221111111")},
						NullValue,
					),
					generateOutboundDncListEntriesBlock(
						[]string{strconv.Quote("+353112222222"), strconv.Quote("+353808080808")},
						NullValue,
					),
					generateOutboundDncListEntriesBlock(
						[]string{strconv.Quote("+353232323232")},
						NullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "contact_method", contactMethod),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
					// Expect two more to be added due to duplicate numbers
					checkPhoneNumbersAddedToDncList("genesyscloud_outbound_dnclist."+resourceID, 5),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_dnclist." + resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"entries"},
			},
		},
		CheckDestroy: testVerifyDncListDestroyed,
	})
}

func TestAccResourceOutboundDncListDncListType(t *testing.T) {
	t.Parallel()
	dncLoginId, present := os.LookupEnv("TEST_DNCCOM_LICENSE_KEY")
	if !present {
		t.Skip("Skipping because TEST_DNCCOM_LICENSE_KEY env variable is not set.")
	}
	var (
		resourceID    = "dnc_list"
		name          = "Test DNC List " + uuid.NewString()
		dncSourceType = "rds"
		contactMethod = "Phone"
		dncCodeB      = "B"
		dncCodeC      = "C"
		dncCodeP      = "P"
		dncCodeT      = "T"
		dncCodes      = []string{strconv.Quote(dncCodeB), strconv.Quote(dncCodeC)}

		nameUpdated         = "Test DNC List " + uuid.NewString()
		dncSourceTypeUpdate = "dnc.com"
		campaignId          = "12132"
		dncCodesUpdated     = append(dncCodes, strconv.Quote(dncCodeP), strconv.Quote(dncCodeT))
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateOutboundDncList(
					resourceID,
					name,
					dncSourceType,
					strconv.Quote(contactMethod),
					NullValue,
					NullValue,
					NullValue,
					[]string{},
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "contact_method", contactMethod),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Update to dnc type dnc.com
				Config: generateOutboundDncList(
					resourceID,
					nameUpdated,
					dncSourceTypeUpdate,
					NullValue,
					strconv.Quote(dncLoginId),
					strconv.Quote(campaignId),
					NullValue,
					dncCodes,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceTypeUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "login_id", dncLoginId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "campaign_id", campaignId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.0", dncCodeB),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.1", dncCodeC),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.#", fmt.Sprintf("%v", len(dncCodes))),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Update dnc codes array
				Config: generateOutboundDncList(
					resourceID,
					nameUpdated,
					dncSourceTypeUpdate,
					NullValue,
					strconv.Quote(dncLoginId),
					strconv.Quote(campaignId),
					NullValue,
					dncCodesUpdated,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceTypeUpdate),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "login_id", dncLoginId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "campaign_id", campaignId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.0", dncCodeB),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.1", dncCodeC),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.2", dncCodeP),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.3", dncCodeT),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.#", fmt.Sprintf("%v", len(dncCodesUpdated))),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_dnclist." + resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dnc_codes"},
			},
		},
		CheckDestroy: testVerifyDncListDestroyed,
	})
}

func TestAccResourceOutboundDncListGryphonListType(t *testing.T) {
	t.Parallel()
	var gryphonLicense string
	var present bool

	present = false

	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "tca" {
		gryphonLicense, present = os.LookupEnv("TEST_DNC_GRYPHON_LICENSE_KEY")
	} else {
		gryphonLicense, present = os.LookupEnv("TEST_DNC_GRYPHON_PROD_LICENSE_KEY")
	}
	if !present {
		t.Skip("Skipping because TEST_DNC_GRYPHON_LICENSE_KEY env variable is not set.")
	}
	var (
		resourceID    = "dnc_list"
		name          = "Test DNC List " + uuid.NewString()
		dncSourceType = "gryphon"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateOutboundDncList(
					resourceID,
					name,
					dncSourceType,
					NullValue,
					NullValue,
					NullValue,
					strconv.Quote(gryphonLicense),
					[]string{},
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "license_id", gryphonLicense),
					provider.TestDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_dnclist." + resourceID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyDncListDestroyed,
	})
}

func generateOutboundDncListEntriesBlock(phoneNumbers []string, expirationDate string) string {
	return fmt.Sprintf(`
	entries {
		expiration_date = %s
		phone_numbers   = [%s]
	}
`, expirationDate, strings.Join(phoneNumbers, ", "))
}

func checkPhoneNumbersAddedToDncList(resource string, numberOfPhoneNumbersAdded int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		r := state.RootModule().Resources[resource]
		if r == nil {
			return fmt.Errorf("%s not found in state", resource)
		}
		outboundAPI := platformclientv2.NewOutboundApi()
		dncListDivisionViews, _, err := outboundAPI.GetOutboundDnclistsDivisionview(r.Primary.ID, true, true)
		if err != nil {
			return fmt.Errorf("error received when quering DNC list division view from API: %v", err)
		}
		if numberOfPhoneNumbersAdded != *dncListDivisionViews.Size {
			return fmt.Errorf("expected dnc list size to be: %v, got: %v", numberOfPhoneNumbersAdded, *dncListDivisionViews.Size)
		}
		return nil
	}
}

func testVerifyDncListDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_dnclist" {
			continue
		}

		dncList, resp, err := outboundAPI.GetOutboundDnclist(rs.Primary.ID, false, false)
		if dncList != nil {
			return fmt.Errorf("dnc list (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// dnc list not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All dnc lists destroyed
	return nil
}

func generateOutboundDncList(
	resourceId string,
	name string,
	dncSourceType string,
	contactMethod string,
	loginId string,
	campaignId string,
	licenseId string,
	dncCodes []string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`
resource "genesyscloud_outbound_dnclist" "%s" {
	name            = "%s"
	dnc_source_type = "%s"
	contact_method  = %s
	login_id        = %s
	license_id      = %s
	campaign_id     = %s
	dnc_codes = [%s]
    %s
}
`, resourceId, name, dncSourceType, contactMethod, loginId, licenseId, campaignId, strings.Join(dncCodes, ", "), strings.Join(nestedBlocks, "\n"))
}
