package access_policy

import (
	"fmt"
	"testing"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAccessPolicy(t *testing.T) {
	var (
		resourceLabel  = "test-access-policy-ds"
		dataLabel      = "access-policy-data"
		name           = "TF DS Test Policy - " + uuid.NewString()
		description    = "Test access policy for data source test"
		targetResource = "authorization:role:view"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// First create the resource
				Config: generateAccessPolicyResource(
					resourceLabel,
					name,
					description,
					targetResource,
					"DENY",
					"ALL",
					"",
					"true",
					"",
				),
			},
			{
				// Then look it up via data source
				Config: generateAccessPolicyResource(
					resourceLabel,
					name,
					description,
					targetResource,
					"DENY",
					"ALL",
					"",
					"true",
					"",
				) + generateAccessPolicyDataSource(
					dataLabel,
					name,
					"genesyscloud_access_policy."+resourceLabel,
				),
				PreConfig: func() {
					t.Log("sleeping to allow for eventual consistency")
					time.Sleep(30 * time.Second)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_access_policy."+dataLabel, "id",
						"genesyscloud_access_policy."+resourceLabel, "id",
					),
				),
			},
		},
	})
}

// generateAccessPolicyDataSource generates a data source configuration for testing
func generateAccessPolicyDataSource(dataLabel, name, dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_access_policy" "%s" {
		name       = "%s"
		depends_on = [%s]
	}
	`, dataLabel, name, dependsOnResource)
}
