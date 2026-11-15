package analytics_dataretention_settings

import (
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Analytics Data Retention Settings Data Source.

WARNING: Like the resource test, this permanently sets the analytics data retention setting of the test org.
*/

func TestAccDataSourceAnalyticsDataretentionSettings(t *testing.T) {
	var (
		resourceLabel   = "analytics_dataretention_settings"
		dataSourceLabel = "analytics_dataretention_settings_data"
		fullResource    = ResourceType + "." + resourceLabel
		fullDataSource  = "data." + ResourceType + "." + dataSourceLabel
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateAnalyticsDataretentionSettingsResource(resourceLabel, "3650") +
					generateAnalyticsDataretentionSettingsDataSource(dataSourceLabel, fullResource),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(fullDataSource, "id", fullResource, "id"),
					// The data source should expose the same value as the managed resource
					resource.TestCheckResourceAttrPair(fullDataSource, "retention_days", fullResource, "retention_days"),
				),
			},
		},
	})
}

func generateAnalyticsDataretentionSettingsDataSource(dataSourceLabel string, dependsOnResource string) string {
	return `
data "` + ResourceType + `" "` + dataSourceLabel + `" {
	depends_on = [` + dependsOnResource + `]
}
`
}
