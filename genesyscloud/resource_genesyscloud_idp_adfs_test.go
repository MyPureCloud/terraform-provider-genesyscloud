package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
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
					GenerateStringArray(strconv.Quote(testCert1)),
					uri1,
					uri2,
					NullValue, // No relying party ID
					NullValue, // Not disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "relying_party_identifier", ""),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "disabled", FalseValue),
				),
			},
			{
				// Update with new values
				Config: generateIdpAdfsResource(
					GenerateStringArray(strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID1),
					TrueValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "relying_party_identifier", relyingPartyID1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "disabled", TrueValue),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpAdfsResource(
					GenerateStringArray(strconv.Quote(testCert1), strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					FalseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert1),
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "disabled", FalseValue),
				),
			},
			{
				// Update to one cert in array
				Config: generateIdpAdfsResource(
					GenerateStringArray(strconv.Quote(testCert1)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					FalseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "disabled", FalseValue),
				),
			},
			{
				// Update back to two certs in array
				Config: generateIdpAdfsResource(
					GenerateStringArray(strconv.Quote(testCert1), strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					FalseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert1),
					ValidateStringInArray("genesyscloud_idp_adfs.adfs", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_adfs.adfs", "disabled", FalseValue),
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
