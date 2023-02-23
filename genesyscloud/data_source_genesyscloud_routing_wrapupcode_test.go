package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWrapupcode(t *testing.T) {
	var (
		codeRes  = "routing-wrapupcode"
		codeData = "codeData"
		codeName = "Terraform Code-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: generateRoutingWrapupcodeResource(
					codeRes,
					codeName,
				) + generateRoutingWrapupcodeDataSource(
					codeData,
					codeName,
					"genesyscloud_routing_wrapupcode."+codeRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_wrapupcode."+codeData, "id", "genesyscloud_routing_wrapupcode."+codeRes, "id"),
				),
			},
		},
	})
}

func generateRoutingWrapupcodeDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_wrapupcode" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
