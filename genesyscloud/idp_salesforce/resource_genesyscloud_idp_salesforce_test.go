package idp_salesforce

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func TestAccResourceIdpSalesforce(t *testing.T) {
	var (
		name            = "Test Salesforce " + uuid.NewString()
		issuerUri       = "https://test.com/1"
		issuerUri2      = "https://test.com/1"
		targetUri       = "https://test.com/2"
		targetUri2      = "https://test.com/2"
		sloUri          = "http://example.com"
		sloBinding      = "HTTP Post"
		sloBinding2     = "HTTP Redirect"
		relyingPartyID1 = "test-id1"
		relyingPartyID2 = "test-id2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIdpSalesforceResource(
					strconv.Quote(name),
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					issuerUri,
					targetUri,
					strconv.Quote(sloUri),
					strconv.Quote(sloBinding),
					util.NullValue, // no relying_party_identifier
					util.NullValue, // Not disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "issuer_uri", issuerUri),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "target_uri", targetUri),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "slo_uri", sloUri),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "slo_binding", sloBinding),
				),
			},
			{
				// Update with new values
				Config: generateIdpSalesforceResource(
					strconv.Quote(name),
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					issuerUri2,
					targetUri2,
					strconv.Quote(sloUri),
					strconv.Quote(sloBinding2),
					strconv.Quote(relyingPartyID1),
					util.NullValue, // Not disabled
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "name", name),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "certificates.#", "1"),
					util.ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "issuer_uri", issuerUri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "target_uri", targetUri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "slo_uri", sloUri),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "slo_binding", sloBinding2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "relying_party_identifier", relyingPartyID1),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpSalesforceResource(
					strconv.Quote(name),
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					issuerUri2,
					targetUri2,
					strconv.Quote(sloUri),
					strconv.Quote(sloBinding2),
					strconv.Quote(relyingPartyID2),
					util.NullValue, // Not disabled
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "name", name),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "certificates.#", "2"),
					util.ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "issuer_uri", issuerUri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "target_uri", targetUri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "slo_uri", sloUri),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "slo_binding", sloBinding2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "relying_party_identifier", relyingPartyID2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_idp_salesforce.salesforce",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIdpSalesforceDestroyed,
	})
}

func generateIdpSalesforceResource(
	name,
	certs,
	issuerURI,
	targetURI,
	sloUri,
	sloBinding,
	relyingPartyIdentifier,
	disabled string) string {
	return fmt.Sprintf(`resource "genesyscloud_idp_salesforce" "salesforce" {
	    name         = %s
		certificates = %s
		issuer_uri   = "%s"
		target_uri   = "%s"
		slo_uri      = %s
		slo_binding  = %s
		relying_party_identifier = %s
        disabled = %s
	}
	`, name, certs, issuerURI, targetURI, sloUri, sloBinding, relyingPartyIdentifier, disabled)
}

func testVerifyIdpSalesforceDestroyed(state *terraform.State) error {
	idpAPI := platformclientv2.NewIdentityProviderApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_idp_salesforce" {
			continue
		}

		salesforce, resp, err := idpAPI.GetIdentityprovidersSalesforce()
		if salesforce != nil {
			return fmt.Errorf("Salesforce still exists")
		} else if util.IsStatus404(resp) {
			// Salesforce not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. Salesforce config destroyed
	return nil
}
