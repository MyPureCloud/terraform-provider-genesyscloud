package idp_okta

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v131/platformclientv2"
)

func TestAccResourceIdpOkta(t *testing.T) {
	var (
		uri1 = "https://test.com/1"
		uri2 = "https://test.com/2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIdpOktaResource(
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					uri1,
					uri2,
					util.NullValue, // Not disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "disabled", util.FalseValue),
				),
			},
			{
				// Update with new values
				Config: generateIdpOktaResource(
					util.GenerateStringArray(strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					util.TrueValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "disabled", util.TrueValue),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpOktaResource(
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					util.FalseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "disabled", util.FalseValue),
				),
			},
			{
				// Update to one cert in array
				Config: generateIdpOktaResource(
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					uri2,
					uri1,
					util.FalseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "disabled", util.FalseValue),
				),
			},
			{
				// Update back to two certs in array
				Config: generateIdpOktaResource(
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					util.FalseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_okta.okta", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_okta.okta", "disabled", util.FalseValue),
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
	certs string,
	issuerURI string,
	targetURI string,
	disabled string) string {
	return fmt.Sprintf(`resource "genesyscloud_idp_okta" "okta" {
		certificates = %s
		issuer_uri = "%s"
		target_uri = "%s"
        disabled = %s
	}
	`, certs, issuerURI, targetURI, disabled)
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
