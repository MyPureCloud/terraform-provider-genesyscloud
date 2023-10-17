package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
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
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIdpGenericResource(
					name1,
					generateStringArray(strconv.Quote(testCert1)),
					uri1,
					uri2,
					nullValue, // No relying party ID
					nullValue, // Not disabled
					nullValue, // no image
					nullValue, // No endpoint compression
					nullValue, // Default name ID format
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name", name1),
					ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", testCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "relying_party_identifier", ""),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "disabled", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "logo_image_data", ""),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "endpoint_compression", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name_identifier_format", nameIDFormatDefault),
				),
			},
			{
				// Update with new values
				Config: generateIdpGenericResource(
					name2,
					generateStringArray(strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID1),
					trueValue, // disabled
					strconv.Quote(base64Img),
					trueValue, // Endpoint compression
					strconv.Quote(nameIDFormatEmail),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name", name2),
					ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "relying_party_identifier", relyingPartyID1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "disabled", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "logo_image_data", base64Img),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "endpoint_compression", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name_identifier_format", nameIDFormatEmail),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpGenericResource(
					name2,
					generateStringArray(strconv.Quote(testCert1), strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					falseValue, // disabled
					strconv.Quote(base64Img),
					trueValue, // Endpoint compression
					strconv.Quote(nameIDFormatEmail),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name", name2),
					ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", testCert1),
					ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "disabled", falseValue),
				),
			},
			{
				// Update to one cert in array
				Config: generateIdpGenericResource(
					name2,
					generateStringArray(strconv.Quote(testCert1)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					falseValue, // disabled
					strconv.Quote(base64Img),
					trueValue, // Endpoint compression
					strconv.Quote(nameIDFormatEmail),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name", name2),
					ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", testCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "disabled", falseValue),
				),
			},
			{
				// Update back to two certs in array
				Config: generateIdpGenericResource(
					name2,
					generateStringArray(strconv.Quote(testCert1), strconv.Quote(testCert2)),
					uri2,
					uri1,
					strconv.Quote(relyingPartyID2),
					falseValue, // disabled
					strconv.Quote(base64Img),
					trueValue, // Endpoint compression
					strconv.Quote(nameIDFormatEmail),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "name", name2),
					ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", testCert1),
					ValidateStringInArray("genesyscloud_idp_generic.generic", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "relying_party_identifier", relyingPartyID2),
					resource.TestCheckResourceAttr("genesyscloud_idp_generic.generic", "disabled", falseValue),
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
	nameIDFormat string) string {
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
	}
	`, name, certs, issuerURI, targetURI, partyID, disabled, logoImageData, endpointCompression, nameIDFormat)
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
		} else if IsStatus404(resp) {
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
