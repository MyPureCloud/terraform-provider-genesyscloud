package external_contacts_external_source

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceexternalSource(t *testing.T) {
	var (
		resourceLabelData = "data_external_source"
		resourceLabel     = "resource_external_source"

		resourcePath     = ResourceType + "." + resourceLabel
		dataResourcePath = "data." + ResourceType + "." + resourceLabelData

		name         = "some-external-source-" + uuid.NewString()
		active       = true
		uri_template = "https://some.host/{{externalId.value}}"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create external source with name and active properties
				Config: GenerateBasicExternalSourceResource(
					resourceLabel,
					name,
					active,
					uri_template,
				) + generateExternalSourceDataSource(
					resourceLabelData,
					name,
					resourcePath,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataResourcePath, "id",
						resourcePath, "id",
					),
				),
			},
		},
	})
}

func generateExternalSourceDataSource(resourceLabel, name, dependsOn string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, ResourceType, resourceLabel, name, dependsOn)
}
