package routing_language

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

func TestAccFrameworkDataSourceRoutingLanguage(t *testing.T) {
	var (
		langResourceLabel   = "routing-language"
		langDataSourceLabel = "routing-language-data"
		langName            = "Terraform Framework Language-" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: getFrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: generateFrameworkRoutingLanguageResource(
					langResourceLabel,
					langName,
				),
			},
			{
				Config: generateFrameworkRoutingLanguageResource(
					langResourceLabel,
					langName,
				) + generateFrameworkRoutingLanguageDataSource(langDataSourceLabel, "genesyscloud_routing_language."+langResourceLabel+".name", "genesyscloud_routing_language."+langResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_language."+langDataSourceLabel, "id", "genesyscloud_routing_language."+langResourceLabel, "id"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_routing_language."+langDataSourceLabel, "name", "genesyscloud_routing_language."+langResourceLabel, "name"),
				),
			},
		},
	})
}

func TestAccFrameworkDataSourceRoutingLanguageNotFound(t *testing.T) {
	var (
		langDataSourceLabel = "routing-language-data-not-found"
		langName            = "NonExistentLanguage" + uuid.NewString()
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { util.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: getFrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      generateFrameworkRoutingLanguageDataSource(langDataSourceLabel, fmt.Sprintf(`"%s"`, langName), ""),
				ExpectError: regexp.MustCompile(".*Could not find routing language.*"),
			},
		},
	})
}

// generateFrameworkRoutingLanguageDataSource generates a routing language data source for Framework testing
func generateFrameworkRoutingLanguageDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	dependsOn := ""
	if dependsOnResource != "" {
		dependsOn = fmt.Sprintf("depends_on=[%s]", dependsOnResource)
	}

	return fmt.Sprintf(`data "genesyscloud_routing_language" "%s" {
		name = %s
        %s
	}
	`, resourceLabel, name, dependsOn)
}
