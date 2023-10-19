package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceIdpGsuite(t *testing.T) {
	var (
		uri1            = "https://test.com/1"
		uri2            = "https://test.com/2"
		relyingPartyID1 = "test-id1"
		relyingPartyID2 = "test-id2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIdpGsuiteResource(
					generateStringArray(strconv.Quote(testCert1)),
					uri1,
					uri2,
					nullValue, // No relying party ID
					nullValue, // Not disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", testCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "relying_party_identifier", ""),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "disabled", falseValue),
				),
			},
			{
				// Update with new values
				Config: generateIdpGsuiteResource(
					generateStringArray(strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID1),
					trueValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "relying_party_identifier", relyingPartyID1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "disabled", trueValue),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpGsuiteResource(
					generateStringArray(strconv.Quote(testCert1), strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					falseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", testCert1),
					ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "disabled", falseValue),
				),
			},
			{
				// Update to one cert in array
				Config: generateIdpGsuiteResource(
					generateStringArray(strconv.Quote(testCert1)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					falseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", testCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "disabled", falseValue),
				),
			},
			{
				// Update back to two certs in array
				Config: generateIdpGsuiteResource(
					generateStringArray(strconv.Quote(testCert1), strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					falseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", testCert1),
					ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "disabled", falseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_idp_gsuite.gsuite",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIdpGsuiteDestroyed,
	})
}

func generateIdpGsuiteResource(
	certs string,
	issuerURI string,
	targetURI string,
	partyID string,
	disabled string) string {
	return fmt.Sprintf(`resource "genesyscloud_idp_gsuite" "gsuite" {
		certificates = %s
		issuer_uri = "%s"
		target_uri = "%s"
        relying_party_identifier = %s
        disabled = %s
	}
	`, certs, issuerURI, targetURI, partyID, disabled)
}

func testVerifyIdpGsuiteDestroyed(state *terraform.State) error {
	idpAPI := platformclientv2.NewIdentityProviderApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_idp_gsuite" {
			continue
		}

		gsuite, resp, err := idpAPI.GetIdentityprovidersGsuite()
		if gsuite != nil {
			return fmt.Errorf("GSuite still exists")
		} else if IsStatus404(resp) {
			// GSuite not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. GSuite config destroyed
	return nil
}
