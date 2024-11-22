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
		langResourceLabel   = "routing-language"
		langDataSourceLabel = "routing-language-data"
		langName            = "Terraform Language-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateRoutingLanguageResource(
					langResourceLabel,
					langName,
				),
			},
			{
				Config: GenerateRoutingLanguageResource(
					langResourceLabel,
					langName,
				) + generateRoutingLanguageDataSource(langDataSourceLabel, "genesyscloud_routing_language."+langResourceLabel+".name", "genesyscloud_routing_language."+langResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_language."+langDataSourceLabel, "id", "genesyscloud_routing_language."+langResourceLabel, "id"),
				),
			},
		},
	})
}

func generateRoutingLanguageDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_language" "%s" {
		name = %s
        depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
