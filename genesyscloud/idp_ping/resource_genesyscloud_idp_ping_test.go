package idp_ping

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

func TestAccResourceIdpPing(t *testing.T) {
	var (
		name1           = "Test ping " + uuid.NewString()
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
				Config: generateIdpPingResource(
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
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_ping.ping", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "relying_party_identifier", ""),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "slo_binding", slo_binding1),
				),
			},
			{
				// Update with new values
				Config: generateIdpPingResource(
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
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_ping.ping", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "relying_party_identifier", relyingPartyID1),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "disabled", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "slo_binding", slo_binding2),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpPingResource(
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
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_ping.ping", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_ping.ping", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "slo_binding", slo_binding1),
				),
			},
			{
				// Update to one cert in array
				Config: generateIdpPingResource(
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
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_ping.ping", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "slo_binding", slo_binding2),
				),
			},
			{
				// Update back to two certs
				Config: generateIdpPingResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					util.FalseValue, // disabled
					uri3,
					slo_binding2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_ping.ping", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_ping.ping", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_ping.ping", "slo_binding", slo_binding2),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_idp_ping.ping",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIdpPingDestroyed,
	})
}

func generateIdpPingResource(
	name string,
	certs string,
	issuerURI string,
	targetURI string,
	partyID string,
	disabled string,
	sloURI string,
	sloBinding string) string {
	return fmt.Sprintf(`resource "genesyscloud_idp_ping" "ping" {
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

func testVerifyIdpPingDestroyed(state *terraform.State) error {
	idpAPI := platformclientv2.NewIdentityProviderApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_idp_ping" {
			continue
		}

		ping, resp, err := idpAPI.GetIdentityprovidersPing()
		if ping != nil {
			return fmt.Errorf("Ping still exists")
		} else if util.IsStatus404(resp) {
			// Ping not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. Ping config destroyed
	return nil
}
