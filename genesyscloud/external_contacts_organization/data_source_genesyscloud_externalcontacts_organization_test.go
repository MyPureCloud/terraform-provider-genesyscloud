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
		uniqueStr         = uuid.NewString()
		resourceLabelData = "data-externalOrganization"
		resourceLabel     = "resource-externalOrganization"

		name              = "john-" + uniqueStr
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
					"genesyscloud_externalcontacts_organization."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_externalcontacts_organization."+resourceLabelData, "id",
						"genesyscloud_externalcontacts_organization."+resourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateExternalOrganizationDataSource(resourceID string, name string, dependsOn string) string {
	return fmt.Sprintf(`data "genesyscloud_externalcontacts_organization" "%s" {
		name = "%s"
		depends_on = [%s]
	}
	`, resourceID, name, dependsOn)
}
