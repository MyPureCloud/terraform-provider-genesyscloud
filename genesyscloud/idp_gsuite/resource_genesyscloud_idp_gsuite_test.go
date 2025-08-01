package idp_gsuite

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourceIdpGsuite(t *testing.T) {
	var (
		name1           = "Test gsuite " + uuid.NewString()
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
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIdpGsuiteResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					uri1,
					uri2,
					util.NullValue, // No relying party ID
					util.NullValue, // Not disabled
					uri3,
					slo_binding1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "slo_binding", slo_binding1),
				),
			},
			{
				// Update with new values
				Config: generateIdpGsuiteResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID1),
					util.TrueValue, // disabled
					uri3,
					slo_binding2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "relying_party_identifier", relyingPartyID1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "disabled", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "slo_binding", slo_binding2),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpGsuiteResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					util.FalseValue, // disabled
					uri3,
					slo_binding1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "slo_binding", slo_binding1),
				),
			},
			{
				// Update to one cert in array
				Config: generateIdpGsuiteResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					util.FalseValue, // disabled
					uri3,
					slo_binding2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "slo_binding", slo_binding2),
				),
			},
			{
				// Update back to two certs in array
				Config: generateIdpGsuiteResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					util.FalseValue, // disabled
					uri3,
					slo_binding1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_gsuite.gsuite", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_gsuite.gsuite", "slo_binding", slo_binding1),
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
	name string,
	certs string,
	issuerURI string,
	targetURI string,
	partyID string,
	disabled string,
	sloURI string,
	sloBinding string) string {
	return fmt.Sprintf(`resource "genesyscloud_idp_gsuite" "gsuite" {
		name = "%s"
		certificates = %s
		issuer_uri = "%s"
		target_uri = "%s"
        relying_party_identifier = %s
        disabled = %s
		slo_uri = "%s"
		slo_binding = "%s"
	}
	`, name, certs, issuerURI, targetURI, partyID, disabled, sloURI, sloBinding)
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
		} else if util.IsStatus404(resp) {
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
