package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v80/platformclientv2"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestAccResourceOutboundDncListRdsListType(t *testing.T) {
	t.Parallel()
	var (
		resourceID    = "dnc_list"
		name          = "Test DNC List " + uuid.NewString()
		dncSourceType = "rds"
		contactMethod = "Phone"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundDncList(
					resourceID,
					name,
					dncSourceType,
					contactMethod,
					"",
					"",
					"",
					[]string{},
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "contact_method", contactMethod),
					testDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_dnclist." + resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dnc_source_type", "contact_method"},
			},
		},
		CheckDestroy: testVerifyDncListDestroyed,
	})
}

func TestAccResourceOutboundDncListDncListType(t *testing.T) {
	t.Parallel()
	dncLoginId, present := os.LookupEnv("TEST_DNCCOM_LICENSE_KEY")
	if !present {
		t.Skip()
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundDncList(
					resourceID,
					name,
					dncSourceType,
					contactMethod,
					"",
					"",
					"",
					[]string{},
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "contact_method", contactMethod),
					testDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Update to dnc type dnc.com
				Config: generateOutboundDncList(
					resourceID,
					nameUpdated,
					dncSourceTypeUpdate,
					"",
					dncLoginId,
					campaignId,
					"",
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
					testDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Update dnc codes array
				Config: generateOutboundDncList(
					resourceID,
					nameUpdated,
					dncSourceTypeUpdate,
					"",
					dncLoginId,
					campaignId,
					"",
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
					testDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_dnclist." + resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dnc_source_type", "login_id", "dnc_codes"},
			},
		},
		CheckDestroy: testVerifyDncListDestroyed,
	})
}

func TestAccResourceOutboundDncListGryphonListType(t *testing.T) {
	t.Parallel()
	gryphonLicense, present := os.LookupEnv("TEST_DNC_GRYPHON_LICENSE_KEY")
	if !present {
		t.Skip()
	}
	var (
		resourceID    = "dnc_list"
		name          = "Test DNC List " + uuid.NewString()
		dncSourceType = "gryphon"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundDncList(
					resourceID,
					name,
					dncSourceType,
					"",
					"",
					"",
					gryphonLicense,
					[]string{},
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "license_id", gryphonLicense),
					testDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_outbound_dnclist." + resourceID,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dnc_source_type"},
			},
		},
		CheckDestroy: testVerifyDncListDestroyed,
	})
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
	if contactMethod != "" {
		contactMethod = fmt.Sprintf(`contact_method = "%s"`, contactMethod)
	}
	if loginId != "" {
		loginId = fmt.Sprintf(`login_id = "%s"`, loginId)
	}
	if campaignId != "" {
		campaignId = fmt.Sprintf(`campaign_id = "%s"`, campaignId)
	}
	if licenseId != "" {
		licenseId = fmt.Sprintf(`license_id = "%s"`, licenseId)
	}
	return fmt.Sprintf(`
resource "genesyscloud_outbound_dnclist" "%s" {
	name            = "%s"
	dnc_source_type = "%s"
	%s
	%s
	%s
    %s
	dnc_codes = [%s]
    %s
}
`, resourceId, name, dncSourceType, contactMethod, loginId, licenseId, campaignId, strings.Join(dncCodes, ", "), strings.Join(nestedBlocks, "\n"))
}

func generateOutboundDncListBasic(resourceId string, name string) string {
	return fmt.Sprintf(`
resource "genesyscloud_outbound_dnclist" "%s" {
	name            = "%s"
	dnc_source_type = "rds"	
	contact_method  = "Phone"
}
`, resourceId, name)
}

func generateDncListEntry(expiryDate string, phoneNumbers []string) string {
	if expiryDate != "" {
		expiryDate = fmt.Sprintf(`expiration_date = "%s"`, expiryDate)
	}
	return fmt.Sprintf(`
	entries {
		%s
		phone_numbers = [%s]
	}
`, expiryDate, strings.Join(phoneNumbers, ", "))
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
		} else if isStatus404(resp) {
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
