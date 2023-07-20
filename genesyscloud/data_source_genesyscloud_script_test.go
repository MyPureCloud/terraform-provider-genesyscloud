package genesyscloud

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util/testrunner"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceScript(t *testing.T) {
	var (
		scriptDataSource = "script-data"
		resourceId       = "script"
		name             = "tfscript" + uuid.NewString()
		filePath         = testrunner.GetTestDataPath("resource", "genesyscloud_script", "test_script.json")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptResource(
					resourceId,
					name,
					filePath,
					"",
				) + generateScriptDataSource(
					scriptDataSource,
					name,
					resourceId,
					false, // published
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_script."+scriptDataSource, "id",
						"genesyscloud_script."+resourceId, "id"),
				),
			},
		},
	})
}

// Test that published scripts can also return hard-coded default scripts
func TestAccDataSourceScriptPublishedDefaults(t *testing.T) {
	var (
		callbackDataSource = "callback-script-data"
		callbackName       = "Default Callback Script"
		callbackId         = "ffde0662-8395-9b04-7dcb-b90172109065"
		inboundDataSource  = "inbound-script-data"
		inboundName        = "Default Inbound Script"
		inboundId          = "766f1221-047a-11e5-bba2-db8c0964d007"
		outboundDataSource = "outbound-script-data"
		outboundName       = "Default Outbound Script"
		outboundId         = "476c2b71-7429-11e4-9a5b-3f91746bffa3"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptDataSource(
					callbackDataSource,
					callbackName,
					"",
					true, // published
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_script."+callbackDataSource, "id",
						callbackId,
					),
				),
			},
			{
				Config: generateScriptDataSource(
					inboundDataSource,
					inboundName,
					"",
					true, // published
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_script."+inboundDataSource, "id",
						inboundId,
					),
				),
			},
			{
				Config: generateScriptDataSource(
					outboundDataSource,
					outboundName,
					"",
					true, // published
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_script."+outboundDataSource, "id",
						outboundId,
					),
				),
			},
		},
	})
}

func generateScriptDataSource(dataSourceID, name, resourceId string, published bool) string {
	if resourceId != "" {
		return fmt.Sprintf(`data "genesyscloud_script" "%s" {
		name = "%s"
		depends_on = [genesyscloud_script.%s]
		published = %t
	}
	`, dataSourceID, name, resourceId, published)
	} else {
		return fmt.Sprintf(`data "genesyscloud_script" "%s" {
		name = "%s"
		published = %t
	}
	`, dataSourceID, name, published)
	}
}
