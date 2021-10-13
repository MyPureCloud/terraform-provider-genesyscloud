package genesyscloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)


func TestAccDataSourceDidPoolBasic(t *testing.T) {
	didPoolResource := "test-didpool1"
	didPoolDataSource := "test-didpool1-data"
	didPoolStartPhoneNumber := "+13175550000"
	didPoolEndPhoneNumber := "+13175550005"
	//didPoolDescription := "Test description"
	//didPoolComments := "Test comments"
	//didPoolProvider := "PURE_CLOUD"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateDidPoolResource(&didPoolStruct{
					didPoolResource,
					didPoolStartPhoneNumber,
					didPoolEndPhoneNumber,
					nullValue, // No description
					nullValue, // No comments
					nullValue, // No provider
				}) +  generateDidPoolDataSource(&didPoolStruct{
					didPoolResource,
					didPoolStartPhoneNumber,
					didPoolEndPhoneNumber,
					nullValue, // No description
					nullValue, // No comments
					nullValue, // No provider
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_did_pool."+didPoolDataSource, "id", "genesyscloud_telephony_providers_edges_did_pool."+didPoolResource, "id" ),
				),
			},
		},
	})
}

func generateDidPoolDataSource(didPool *didPoolStruct) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_did_pool" "%s" {
		start_phone_number = "%s"
		end_phone_number = "%s"
		description = %s
		comments = %s
		pool_provider = %s
	}
	`, didPool.resourceID,
		didPool.startPhoneNumber,
		didPool.endPhoneNumber,
		didPool.description,
		didPool.comments,
		didPool.poolProvider)
}