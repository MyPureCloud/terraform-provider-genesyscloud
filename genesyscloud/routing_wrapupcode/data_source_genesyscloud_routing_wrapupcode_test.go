package routing_wrapupcode

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	authDivision "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/auth_division"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccFrameworkDataSourceRoutingWrapupcode(t *testing.T) {
	var (
		resourceLabel = "test_routing_wrapupcode"
		dataLabel     = "test_data_wrapupcode"
		name          = "Terraform Framework Data Wrapupcode " + uuid.NewString()
		description   = "Test wrapupcode for data source"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: getFrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: generateFrameworkRoutingWrapupcodeResource(resourceLabel, name, util.NullValue, description) +
					generateFrameworkRoutingWrapupcodeDataSource(dataLabel, name, "genesyscloud_routing_wrapupcode."+resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_wrapupcode."+dataLabel, "id", "genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
					resource.TestCheckResourceAttr("data.genesyscloud_routing_wrapupcode."+dataLabel, "name", name),
				),
			},
		},
	})
}

func TestAccFrameworkDataSourceRoutingWrapupcodeWithDivision(t *testing.T) {
	var (
		resourceLabel    = "test_routing_wrapupcode_div"
		dataLabel        = "test_data_wrapupcode_div"
		name             = "Terraform Framework Data Wrapupcode Div " + uuid.NewString()
		description      = "Test wrapupcode with division for data source"
		divResourceLabel = "test_division"
		divName          = "terraform-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: getFrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: authDivision.GenerateAuthDivisionBasic(divResourceLabel, divName) +
					generateFrameworkRoutingWrapupcodeResource(resourceLabel, name, "genesyscloud_auth_division."+divResourceLabel+".id", description) +
					generateFrameworkRoutingWrapupcodeDataSource(dataLabel, name, "genesyscloud_routing_wrapupcode."+resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_wrapupcode."+dataLabel, "id", "genesyscloud_routing_wrapupcode."+resourceLabel, "id"),
					resource.TestCheckResourceAttr("data.genesyscloud_routing_wrapupcode."+dataLabel, "name", name),
				),
			},
		},
	})
}

// generateFrameworkRoutingWrapupcodeDataSource generates a routing wrapupcode data source for Framework testing
func generateFrameworkRoutingWrapupcodeDataSource(dataLabel string, name string, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_wrapupcode" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, dataLabel, name, dependsOnResource)
}
