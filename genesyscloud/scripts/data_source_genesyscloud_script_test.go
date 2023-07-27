package scripts

import (
	"fmt"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test cases for Scripts Datasource
*/
func TestAccDataSourceScript(t *testing.T) {
	var (
		scriptDataSource = "script-data"
		resourceId       = "script"
		name             = "tfscript" + uuid.NewString()
		filePath         = getTestDataPath("resource", "genesyscloud_script", "test_script.json")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
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
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptDataSource(
					callbackDataSource,
					callbackName,
					"",
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

func generateScriptDataSource(dataSourceID, name, resourceId string) string {
	if resourceId != "" {
		return fmt.Sprintf(`data "genesyscloud_script" "%s" {
		name = "%s"
		depends_on = [genesyscloud_script.%s]
	}
	`, dataSourceID, name, resourceId)
	} else {
		return fmt.Sprintf(`data "genesyscloud_script" "%s" {
		name = "%s"
	}
	`, dataSourceID, name)
	}
}
