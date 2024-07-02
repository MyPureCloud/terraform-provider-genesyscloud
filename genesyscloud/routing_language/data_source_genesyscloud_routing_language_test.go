package routing_language

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingLanguage(t *testing.T) {
	var (
		langResource   = "routing-language"
		langDataSource = "routing-language-data"
		langName       = "Terraform Language-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateRoutingLanguageResource(
					langResource,
					langName,
				),
			},
			{
				Config: GenerateRoutingLanguageResource(
					langResource,
					langName,
				) + generateRoutingLanguageDataSource(langDataSource, "genesyscloud_routing_language."+langResource+".name", "genesyscloud_routing_language."+langResource),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_language."+langDataSource, "id", "genesyscloud_routing_language."+langResource, "id"),
				),
			},
		},
	})
}

func generateRoutingLanguageDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_language" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
