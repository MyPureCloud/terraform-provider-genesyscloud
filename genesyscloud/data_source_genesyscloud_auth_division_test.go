package genesyscloud

import (
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAuthDivision(t *testing.T) {
	var (
		divResource   = "auth-division"
		divDataSource = "auth-div-data"
		divName       = "Terraform Divisions-" + uuid.NewString()
		divisionID    string
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					time.Sleep(30 * time.Second)
				},
				Config: GenerateAuthDivisionResource(
					divResource,
					divName,
					util.NullValue,
					util.NullValue,
				),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["genesyscloud_auth_division."+divResource]
						if !ok {
							return fmt.Errorf("not found: %s", "genesyscloud_auth_division."+divResource)
						}
						divisionID = rs.Primary.ID
						log.Printf("Division ID: %s\n", divisionID) // Print ID
						return nil
					},
				),
				PreventPostDestroyRefresh: true,
			},
			{
				Config: GenerateAuthDivisionResource(
					divResource,
					divName,
					util.NullValue,
					util.NullValue,
				) + generateAuthDivisionDataSource(
					divDataSource,
					"genesyscloud_auth_division."+divResource+".name",
					"genesyscloud_auth_division."+divResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_auth_division."+divDataSource, "id", "genesyscloud_auth_division."+divResource, "id"),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_auth_division." + divResource,
				ImportState:       true,
				ImportStateVerify: true,
				Destroy:           true,
			},
		},
		CheckDestroy: func(state *terraform.State) error {
			time.Sleep(45 * time.Second)
			return testVerifyDivisionsDestroyed(state)
		},
	})
}

func generateAuthDivisionDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_auth_division" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
