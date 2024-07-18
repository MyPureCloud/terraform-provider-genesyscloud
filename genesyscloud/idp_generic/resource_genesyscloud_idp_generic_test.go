package idp_generic

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceIdpGeneric(t *testing.T) {
	var (
		name1               = "generic1"
		name2               = "generic2"
		uri1                = "https://test.com/1"
		uri2                = "https://test.com/2"
		relyingPartyID1     = "test-id1"
		relyingPartyID2     = "test-id2"
		nameIDFormatDefault = "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
		nameIDFormatEmail   = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
		base64Img           = "PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvM"
		uri3                = "https://example.com"
		slo_binding1        = "HTTP Redirect"
		slo_binding2        = "HTTP Post"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIdpGenericResource(
					name1,
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					uri1,
					uri2,
					util.NullValue, // No relying party ID
					util.NullValue, // Not disabled
					util.NullValue, // no image
					util.NullValue, // No endpoint compression
					util.NullValue, // Default name ID format
					uri3,
					slo_binding1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name", name1),
					util.ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "relying_party_identifier", ""),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "logo_image_data", ""),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "endpoint_compression", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name_identifier_format", nameIDFormatDefault),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "slo_binding", slo_binding1),
				),
			},
			{
				// Update with new values
				Config: generateIdpGenericResource(
					name2,
					util.GenerateStringArray(strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID1),
					util.TrueValue, // disabled
					strconv.Quote(base64Img),
					util.TrueValue, // Endpoint compression
					strconv.Quote(nameIDFormatEmail),
					uri3,
					slo_binding2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name", name2),
					util.ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "relying_party_identifier", relyingPartyID1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "disabled", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "logo_image_data", base64Img),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "endpoint_compression", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name_identifier_format", nameIDFormatEmail),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "slo_binding", slo_binding2),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpGenericResource(
					name2,
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					util.FalseValue, // disabled
					strconv.Quote(base64Img),
					util.TrueValue, // Endpoint compression
					strconv.Quote(nameIDFormatEmail),
					uri3,
					slo_binding1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name", name2),
					util.ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "slo_binding", slo_binding1),
				),
			},
			{
				// Update to one cert in array
				Config: generateIdpGenericResource(
					name2,
					util.GenerateStringArray(strconv.Quote(util.TestCert1)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					util.FalseValue, // disabled
					strconv.Quote(base64Img),
					util.TrueValue, // Endpoint compression
					strconv.Quote(nameIDFormatEmail),
					uri3,
					slo_binding2,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name", name2),
					util.ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", util.TestCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "slo_binding", slo_binding2),
				),
			},
			{
				// Update back to two certs in array
				Config: generateIdpGenericResource(
					name2,
					util.GenerateStringArray(strconv.Quote(util.TestCert1), strconv.Quote(util.TestCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					util.FalseValue, // disabled
					strconv.Quote(base64Img),
					util.TrueValue, // Endpoint compression
					strconv.Quote(nameIDFormatEmail),
					uri3,
					slo_binding1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name", name2),
					util.ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", util.TestCert1),
					util.ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", util.TestCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "disabled", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "slo_uri", uri3),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "slo_binding", slo_binding1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_idp_generic.generic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIdpGenericDestroyed,
	})
}

func generateIdpGenericResource(
	name string,
	certs string,
	issuerURI string,
	targetURI string,
	partyID string,
	disabled string,
	logoImageData string,
	endpointCompression string,
	nameIDFormat string,
	sloURI string,
	sloBinding string) string {
	return fmt.Sprintf(`resource "genesyscloud_idp_generic" "generic" {
        name = "%s"
		certificates = %s
		issuer_uri = "%s"
		target_uri = "%s"
        relying_party_identifier = %s
        disabled = %s
        logo_image_data = %s
        endpoint_compression = %s
        name_identifier_format = %s
		slo_uri = "%s"
		slo_binding = "%s"
	}
	`, name, certs, issuerURI, targetURI, partyID, disabled, logoImageData, endpointCompression, nameIDFormat, sloURI, sloBinding)
}

func testVerifyIdpGenericDestroyed(state *terraform.State) error {
	idpAPI := platformclientv2.NewIdentityProviderApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_idp_generic" {
			continue
		}

		generic, resp, err := idpAPI.GetIdentityprovidersGeneric()
		if generic != nil {
			return fmt.Errorf("Generic IDP still exists")
		} else if util.IsStatus404(resp) {
			// Generic IDP not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. Generic IDP config destroyed
	return nil
}
