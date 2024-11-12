package routing_wrapupcode

import (
	"fmt"
	authDivision "terraform-provider-genesyscloud/genesyscloud/auth_division"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWrapupcode(t *testing.T) {
	var (
		codeRes     = "routing-wrapupcode"
		codeData    = "codeData"
		codeName    = "Terraform Code-" + uuid.NewString()
		divResource = "test-division"
		divName     = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: authDivision.GenerateAuthDivisionBasic(divResource, divName) + GenerateRoutingWrapupcodeResource(
					codeRes,
					codeName,
					"genesyscloud_auth_division."+divResource+".id",
				) + generateRoutingWrapupcodeDataSource(
					codeData,
					codeName,
					resourceName+"."+codeRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+resourceName+"."+codeData, "id", resourceName+"."+codeRes, "id"),
				),
			},
		},
	})
}

func generateRoutingWrapupcodeDataSource(resourceLabel string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceName, resourceLabel, name, dependsOnResource)
}
