package outbound_attempt_limit

import (
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAttemptLimit(t *testing.T) {

	var (
		resourceId       = "attempt_limit"
		attemptLimitName = "Test Limit " + uuid.NewString()
		dataSourceId     = "attempt_limit_data"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateAttemptLimitResource(
					resourceId,
					attemptLimitName,
					"1",
					"",
					"",
					"",
					"",
				) + GenerateOutboundAttemptLimitDataSource(
					dataSourceId,
					attemptLimitName,
					"genesyscloud_outbound_attempt_limit."+resourceId,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_attempt_limit."+dataSourceId, "id",
						"genesyscloud_outbound_attempt_limit."+resourceId, "id"),
				),
			},
		},
	})
}
