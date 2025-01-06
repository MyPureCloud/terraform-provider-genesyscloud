package external_contacts_organization

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceexternalOrganization(t *testing.T) {
	var (
		resourceLabelData = "data_external_organization"
		resourceLabel     = "resource_external_organization"

		resourcePath     = ResourceType + "." + resourceLabel
		dataResourcePath = "data." + ResourceType + "." + resourceLabelData

		name              = "john-" + uuid.NewString()
		phoneDisplay      = "+1 321-700-1243"
		countryCode       = "US"
		address           = "1011 New Hope St"
		city              = "Norristown"
		state             = "PA"
		postalCode        = "19401"
		twitterId         = "twitterId"
		twitterName       = "twitterName"
		twitterScreenName = "twitterScreenname"
		symbol            = "ABC"
		exchange          = "NYSE"
		tags              = []string{
			strconv.Quote("news"),
			strconv.Quote("channel"),
		}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create external contact with an lastname and others property
				Config: GenerateBasicExternalOrganizationResource(
					resourceLabel,
					name,
					phoneDisplay, countryCode,
					address, city, state, postalCode, countryCode,
					twitterId, twitterName, twitterScreenName,
					symbol, exchange,
					tags,
					"",
				) + generateExternalOrganizationDataSource(
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

func generateExternalOrganizationDataSource(resourceLabel, name, dependsOn string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, ResourceType, resourceLabel, name, dependsOn)
}
