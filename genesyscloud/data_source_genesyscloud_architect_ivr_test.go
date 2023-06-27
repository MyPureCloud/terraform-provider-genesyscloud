package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceArchitectIvr(t *testing.T) {
	t.Parallel()
	var (
		ivrResource = "arch-ivr"
		name        = "IVR " + uuid.NewString()
		description = "Sample IVR by CX as Code"

		ivrDataSource = "arch-ivr-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateIvrConfigResource(&ivrConfigStruct{
					resourceID:  ivrResource,
					name:        name,
					description: description,
					dnis:        nil,
					depends_on:  "",
				}) + generateIvrDataSource(ivrDataSource,
					"genesyscloud_architect_ivr."+ivrResource+".name",
					"genesyscloud_architect_ivr."+ivrResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_architect_ivr."+ivrDataSource, "id", "genesyscloud_architect_ivr."+ivrResource, "id"),
				),
			},
		},
	})
}

func generateIvrDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_architect_ivr" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
