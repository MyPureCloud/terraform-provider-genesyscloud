package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceIdpSalesforce(t *testing.T) {
	var (
		uri1 = "https://test.com/1"
		uri2 = "https://test.com/2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIdpSalesforceResource(
					generateStringArray(strconv.Quote(testCert1)),
					uri1,
					uri2,
					nullValue, // Not disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", testCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "issuer_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "target_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "disabled", falseValue),
				),
			},
			{
				// Update with new values
				Config: generateIdpSalesforceResource(
					generateStringArray(strconv.Quote(testCert2)),
					uri2,
					uri1,
					trueValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "disabled", trueValue),
				),
			},
			{
				// Update with multiple certs
				Config: generateIdpSalesforceResource(
					generateStringArray(strconv.Quote(testCert1), strconv.Quote(testCert2)),
					uri2,
					uri1,
					falseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", testCert1),
					ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "disabled", falseValue),
				),
			},
			{
				// Update back to one cert in array
				Config: generateIdpSalesforceResource(
					generateStringArray(strconv.Quote(testCert1)),
					uri2,
					uri1,
					falseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", testCert1),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "certificates.#", "1"),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "disabled", falseValue),
				),
			},
			{
				// Update back to two certs in array
				Config: generateIdpSalesforceResource(
					generateStringArray(strconv.Quote(testCert1), strconv.Quote(testCert2)),
					uri2,
					uri1,
					falseValue, // disabled
				),
				Check: resource.ComposeTestCheckFunc(
					ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", testCert1),
					ValidateStringInArray("genesyscloud_idp_salesforce.salesforce", "certificates", testCert2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "certificates.#", "2"),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "issuer_uri", uri2),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "target_uri", uri1),
					resource.TestCheckResourceAttr("genesyscloud_idp_salesforce.salesforce", "disabled", falseValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_idp_salesforce.salesforce",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyIdpSalesforceDestroyed,
	})
}

func generateIdpSalesforceResource(
	certs string,
	issuerURI string,
	targetURI string,
	disabled string) string {
	return fmt.Sprintf(`resource "genesyscloud_idp_salesforce" "salesforce" {
		certificates = %s
		issuer_uri = "%s"
		target_uri = "%s"
        disabled = %s
	}
	`, certs, issuerURI, targetURI, disabled)
}

func testVerifyIdpSalesforceDestroyed(state *terraform.State) error {
	idpAPI := platformclientv2.NewIdentityProviderApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_idp_salesforce" {
			continue
		}

		salesforce, resp, err := idpAPI.GetIdentityprovidersSalesforce()
		if salesforce != nil {
			return fmt.Errorf("Salesforce still exists")
		} else if IsStatus404(resp) {
			// Salesforce not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. Salesforce config destroyed
	return nil
}
