package analytics_dataretention_settings

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the Analytics Data Retention Settings Resource.

Note: This is an organization-level singleton resource with no create or delete API operation.
Because the setting cannot be cleared once configured, there is no CheckDestroy assertion (delete is a no-op).

WARNING: Running this test permanently changes the analytics data retention setting of the test org; it cannot
be cleared afterwards. The values are deliberately high and the test ends on the higher one so that it never
lowers the setting. The API rejects a value (409) when existing recording media retention policies, recordings or
the snippet recording settings of the test org exceed it, so the test org must not hold longer retention than these values.
*/

func TestAccResourceAnalyticsDataretentionSettings(t *testing.T) {
	var (
		resourceLabel = "analytics_dataretention_settings"
		fullResource  = ResourceType + "." + resourceLabel

		retentionDays1 = "3649"
		retentionDays2 = "3650"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create / initial update
				Config: generateAnalyticsDataretentionSettingsResource(resourceLabel, retentionDays1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResource, "id", analyticsDataretentionSettingsId),
					resource.TestCheckResourceAttr(fullResource, "retention_days", retentionDays1),
				),
			},
			{
				// Update
				Config: generateAnalyticsDataretentionSettingsResource(resourceLabel, retentionDays2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fullResource, "retention_days", retentionDays2),
				),
			},
			{
				// Import/Read
				ResourceName:      fullResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceAnalyticsDataretentionSettingsValidation verifies that the ValidateFunc on retention_days
// rejects values below the allowed minimum at plan time, before any API call is made.
func TestAccResourceAnalyticsDataretentionSettingsValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config:      generateAnalyticsDataretentionSettingsResource("analytics_dataretention_settings", "0"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`expected retention_days to be at least \(1\)`),
			},
		},
	})
}

func generateAnalyticsDataretentionSettingsResource(resourceLabel string, retentionDays string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
	retention_days = %s
}
`, ResourceType, resourceLabel, retentionDays)
}
