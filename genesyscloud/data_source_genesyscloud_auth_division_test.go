package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAuthDivision(t *testing.T) {
	var (
		divResource   = "auth-div"
		divDataSource = "auth-div-data"
		divName       = "Terraform Div-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAuthDivisionResource(
					divResource,
					divName,
					nullValue,
					nullValue,
				) + generateAuthDivisionDataSource(
					divDataSource,
					"genesyscloud_auth_division."+divResource+".name",
					"genesyscloud_auth_division."+divResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_auth_division."+divDataSource, "id", "genesyscloud_auth_division."+divResource, "id"),
				),
			},
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
