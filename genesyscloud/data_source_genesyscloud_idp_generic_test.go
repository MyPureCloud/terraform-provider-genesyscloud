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
		name1 = "generic"
		name2 = "generic2"
		//name1     = "generic"
		//name2 = "idpData"
		name                     = "test idp" + uuid.NewString()
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
					name1,
					generateStringArray(strconv.Quote(testCert1)),
					uri1,
					uri2,
					nullValue, // No relying party ID
					nullValue, // Not disabled
					nullValue, // no image
					nullValue, // No endpoint compression
					nullValue, // Default name ID format
				) + generateIdpGenericDataSource(
					name2,
					name,
					"genesyscloud_idp_generic." + name1,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_idp_generic." + name2, "id", "genesyscloud_idp_generic." + name1, "id"),
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
