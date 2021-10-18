package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIdpGeneric(t *testing.T) {
	var (
		idpRes = "identityProvider"
		idpDataRes = "identityProviderData"
		name = "test idp" + uuid.NewString()
		uri1  = "https://test.com/1"
		uri2  = "https://test.com/2"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIdpGenericResource(
					name,
					generateStringArray(strconv.Quote(testCert1)),
					uri1,
					uri2,
					nullValue, // No relying party ID
					nullValue, // Not disabled
					nullValue, // no image
					nullValue, // No endpoint compression
					nullValue, // Default name ID format
				) + generateIdpGenericDataSource(
					idpDataRes,
					name,
					"genesyscloud_idp_generic." + idpRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_idp_generic." + idpDataRes, "id", "genesyscloud_idp_generic." + idpRes, "id"),
				),
			},
		},
	})
}

func generateIdpGenericDataSource(
	resourceID string,
	name string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_idp_generic" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
