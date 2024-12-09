package outbound_attempt_limit

import (
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundAttemptLimit(t *testing.T) {

	var (
		resourceLabel    = "attempt_limit"
		resourcePath     = ResourceType + "." + resourceLabel
		attemptLimitName = "Test Limit " + uuid.NewString()
		dataSourceLabel  = "attempt_limit_data"
		dataResourcePath = "data." + ResourceType + "." + dataSourceLabel
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateAttemptLimitResource(
					resourceLabel,
					attemptLimitName,
					"1",
					"",
					"",
					"",
					"",
				) + GenerateOutboundAttemptLimitDataSource(
					dataSourceLabel,
					attemptLimitName,
					resourcePath,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataResourcePath, "id",
						resourcePath, "id"),
				),
			},
		},
	})
}
