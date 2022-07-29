package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v75/platformclientv2"
	"strconv"
	"strings"
	"testing"
)

const (
	dncLoginId     = "96CAAC02650543056DF1ADA796A0082ED152561EDEE1"
	gryphonLicense = "4ADA-9A3B-5DAD-4FAD-95A7-4C31-425B-8594"
)

func TestAccResourceOutboundDncList(t *testing.T) {
	t.Parallel()
	var (
		resourceID    = "dnc_list"
		name          = "Test DNC List " + uuid.NewString()
		dncSourceType = "rds"
		contactMethod = "Phone"
		dncCode1      = "B"
		dncCode2      = "D"
		dncCode3      = "P"
		dncCode4      = "S"
		dncCodes      = []string{strconv.Quote(dncCode1), strconv.Quote(dncCode2)}

		nameUpdated          = "Test DNC List " + uuid.NewString()
		dncSourceTypeUpdate1 = "dnc.com"
		dncCodesUpdated      = append(dncCodes, strconv.Quote(dncCode3), strconv.Quote(dncCode4))

		dncSourceTypeUpdate2 = "gryphon"
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
					[]string{},
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
					dncSourceTypeUpdate1,
					"",
					dncLoginId,
					"",
					dncCodes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceTypeUpdate1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "login_id", dncLoginId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.0", dncCode1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.1", dncCode2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.#", fmt.Sprintf("%v", len(dncCodes))),
					testDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Update dnc codes array
				Config: generateOutboundDncList(
					resourceID,
					nameUpdated,
					dncSourceTypeUpdate1,
					"",
					dncLoginId,
					"",
					dncCodesUpdated,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceTypeUpdate1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "login_id", dncLoginId),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.0", dncCode1),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.1", dncCode2),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.2", dncCode3),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.3", dncCode4),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_codes.#", fmt.Sprintf("%v", len(dncCodesUpdated))),
					testDefaultHomeDivision("genesyscloud_outbound_dnclist."+resourceID),
				),
			},
			{
				// Update to dnc type Gryphon
				Config: generateOutboundDncList(
					resourceID,
					nameUpdated,
					dncSourceTypeUpdate2,
					"",
					"",
					gryphonLicense,
					[]string{},
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_dnclist."+resourceID, "dnc_source_type", dncSourceTypeUpdate2),
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
	licenseId string,
	dncCodes []string) string {
	if contactMethod != "" {
		contactMethod = fmt.Sprintf(`contact_method = "%s"`, contactMethod)
	}
	if loginId != "" {
		loginId = fmt.Sprintf(`login_id = "%s"`, loginId)
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
	dnc_codes = [%s]
}
`, resourceId, name, dncSourceType, contactMethod, loginId, licenseId, strings.Join(dncCodes, ", "))
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
