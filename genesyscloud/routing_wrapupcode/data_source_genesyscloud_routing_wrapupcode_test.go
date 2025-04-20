package routing_wrapupcode

import (
	"fmt"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWrapupcode(t *testing.T) {
	var (
		codeResourceLabel = "routing-wrapupcode"
		codeDataLabel     = "codeData"
		codeName          = "Terraform Code-" + uuid.NewString()
		divResourceLabel  = "test-division"
		divName           = "terraform-" + uuid.NewString()
		description       = "Terraform wrapup code description"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) + GenerateRoutingWrapupcodeResource(
					codeResourceLabel,
					codeName,
					"genesyscloud_auth_division."+divResourceLabel+".id",
					description,
				) + generateRoutingWrapupcodeDataSource(
					codeDataLabel,
					codeName,
					ResourceType+"."+codeResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+codeDataLabel, "id", ResourceType+"."+codeResourceLabel, "id"),
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
	`, ResourceType, resourceLabel, name, dependsOnResource)
}
