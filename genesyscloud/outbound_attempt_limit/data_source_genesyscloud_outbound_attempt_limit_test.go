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
		attemptLimitName = "Test Limit " + uuid.NewString()
		dataSourceLabel  = "attempt_limit_data"
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
					"genesyscloud_outbound_attempt_limit."+resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_attempt_limit."+dataSourceLabel, "id",
						"genesyscloud_outbound_attempt_limit."+resourceLabel, "id"),
				),
			},
		},
	})
}
