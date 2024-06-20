package genesyscloud

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

func TestAccResourceIdpOnelogin(t *testing.T) {
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
				Config: generateIdpOneloginResource(
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					uri1,
					uri2,
					util.NullValue, // Not disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_onelogin.onelogin", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "disabled", util.FalseValue),
				),
			},
			{
				// Update with new values
				Config: generateIdpOneloginResource(
					util.GenerateStringArray(strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					util.TrueValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_onelogin.onelogin", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "disabled", util.TrueValue),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpOneloginResource(
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					util.FalseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_onelogin.onelogin", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_onelogin.onelogin", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "disabled", util.FalseValue),
				),
			},
			{
				// Update to one cert in array
				Config: generateIdpOneloginResource(
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					uri2,
					uri1,
					util.FalseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_onelogin.onelogin", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "disabled", util.FalseValue),
				),
			},
			{
				// Update back to two certs in array
				Config: generateIdpOneloginResource(
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					util.FalseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					util.ValidateStringInArray("genesyscloud_idp_onelogin.onelogin", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_onelogin.onelogin", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_onelogin.onelogin", "disabled", util.FalseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_idp_onelogin.onelogin",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIdpOneloginDestroyed,
	})
}

func generateIdpOneloginResource(
	certs string,
	issuerURI string,
	targetURI string,
	disabled string) string {
	return fmt.Sprintf(`resource "genesyscloud_idp_onelogin" "onelogin" {
		certificates = %s
		issuer_uri = "%s"
		target_uri = "%s"
        disabled = %s
	}
	`, certs, issuerURI, targetURI, disabled)
}

func testVerifyIdpOneloginDestroyed(state *terraform.State) error {
	idpAPI := platformclientv2.NewIdentityProviderApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_idp_onelogin" {
			continue
		}

		onelogin, resp, err := idpAPI.GetIdentityprovidersOnelogin()
		if onelogin != nil {
			return fmt.Errorf("Onelogin still exists")
		} else if util.IsStatus404(resp) {
			// Onelogin not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. Onelogin config destroyed
	return nil
}
