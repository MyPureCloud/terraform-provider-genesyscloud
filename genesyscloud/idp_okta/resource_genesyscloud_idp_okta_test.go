package idp_okta

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceIdpOkta(t *testing.T) {
	var (
		name1           = "Test okta " + uuid.NewString()
		uri1            = "https://test.com/1"
		uri2            = "https://test.com/2"
		relyingPartyID1 = "test-id1"
		relyingPartyID2 = "test-id2"
		uri3            = "https://example.com"
		slo_binding1    = "HTTP Redirect"
		slo_binding2    = "HTTP Post"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIdpOktaResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					uri1,
					uri2,
					util.NullValue, // Not disabled
					util.NullValue, // No relying party ID
					uri3,
					slo_binding1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "relying_party_identifier", ""),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "slo_binding", slo_binding1),
				),
			},
			{
				// Update with new values
				Config: generateIdpOktaResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					util.TrueValue, // disabled
					strconv.Quote(relyingPartyID1),
					uri3,
					slo_binding2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "disabled", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "relying_party_identifier", relyingPartyID1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "slo_binding", slo_binding2),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpOktaResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					util.FalseValue, // disabled
					strconv.Quote(relyingPartyID2),
					uri3,
					slo_binding1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "slo_binding", slo_binding1),
				),
			},
			{
				// Update to one cert in array
				Config: generateIdpOktaResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					uri2,
					uri1,
					util.FalseValue, // disabled
					strconv.Quote(relyingPartyID2),
					uri3,
					slo_binding2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "slo_binding", slo_binding2),
				),
			},
			{
				// Update back to two certs in array
				Config: generateIdpOktaResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					util.FalseValue, // disabled
					strconv.Quote(relyingPartyID2),
					uri3,
					slo_binding1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "slo_binding", slo_binding1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_idp_okta.okta",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIdpOktaDestroyed,
	})
}

func generateIdpOktaResource(
	name string,
	certs string,
	issuerURI string,
	targetURI string,
	disabled string,
	partyID string,
	sloURI string,
	sloBinding string) string {
	return fmt.Sprintf(`resource "genesyscloud_idp_okta" "okta" {
		name = "%s"
		certificates = %s
		issuer_uri = "%s"
		target_uri = "%s"
        disabled = %s
		relying_party_identifier = %s
		slo_uri = "%s"
		slo_binding = "%s"
	}
	`, name, certs, issuerURI, targetURI, disabled, partyID, sloURI, sloBinding)
}

func testVerifyIdpOktaDestroyed(state *terraform.State) error {
	idpAPI := platformclientv2.NewIdentityProviderApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_idp_okta" {
			continue
		}

		okta, resp, err := idpAPI.GetIdentityprovidersOkta()
		if okta != nil {
			return fmt.Errorf("Okta still exists")
		} else if util.IsStatus404(resp) {
			// Okta not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. Okta config destroyed
	return nil
}
