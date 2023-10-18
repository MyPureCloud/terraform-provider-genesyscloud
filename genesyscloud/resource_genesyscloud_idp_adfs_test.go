package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceIdpAdfs(t *testing.T) {
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
				Config: generateIdpAdfsResource(
					generateStringArray(strconv.Quote(testCert1)),
					uri1,
					uri2,
					nullValue, // No relying party ID
					nullValue, // Not disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "relying_party_identifier", ""),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "disabled", falseValue),
				),
			},
			{
				// Update with new values
				Config: generateIdpAdfsResource(
					generateStringArray(strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID1),
					trueValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "relying_party_identifier", relyingPartyID1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "disabled", trueValue),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpAdfsResource(
					generateStringArray(strconv.Quote(testCert1), strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					falseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert1),
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "disabled", falseValue),
				),
			},
			{
				// Update to one cert in array
				Config: generateIdpAdfsResource(
					generateStringArray(strconv.Quote(testCert1)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					falseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "disabled", falseValue),
				),
			},
			{
				// Update back to two certs in array
				Config: generateIdpAdfsResource(
					generateStringArray(strconv.Quote(testCert1), strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					falseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert1),
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "disabled", falseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_idp_adfs.adfs",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIdpAdfsDestroyed,
	})
}

func generateIdpAdfsResource(
	certs string,
	issuerURI string,
	targetURI string,
	partyID string,
	disabled string) string {
	return fmt.Sprintf(`resource "genesyscloud_idp_adfs" "adfs" {
		certificates = %s
		issuer_uri = "%s"
		target_uri = "%s"
        relying_party_identifier = %s
        disabled = %s
	}
	`, certs, issuerURI, targetURI, partyID, disabled)
}

func testVerifyIdpAdfsDestroyed(state *terraform.State) error {
	idpAPI := platformclientv2.NewIdentityProviderApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_idp_adfs" {
			continue
		}

		adfs, resp, err := idpAPI.GetIdentityprovidersAdfs()
		if adfs != nil {
			return fmt.Errorf("ADFS still exists")
		} else if IsStatus404(resp) {
			// ADFS not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. ADFS config destroyed
	return nil
}
