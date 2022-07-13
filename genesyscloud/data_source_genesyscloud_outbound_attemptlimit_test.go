package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceAttemptLimit(t *testing.T) {
	var (
		resourceId       = "attempt_limit"
		attemptLimitName = "Test Limit " + uuid.NewString()
		dataSourceId     = "attempt_limit_data"
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateAttemptLimitResource(
					resourceId,
					attemptLimitName,
					"1",
					"",
					"",
					"",
					"",
				) + generateOutboundAttemptLimitDataSource(
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

func generateOutboundAttemptLimitDataSource(id string, attemptLimitName string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_attempt_limit" "%s" {
	name = "%s"
	depends_on = [%s]
}
`, id, attemptLimitName, dependsOn)
}
