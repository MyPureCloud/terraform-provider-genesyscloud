package scripts

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"testing"

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
		filePath         = getTestDataPath("resource", resourceName, "test_script.json")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
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
					resource.TestCheckResourceAttrPair(fmt.Sprintf("data.%s.%s", resourceName, scriptDataSource), "id",
						resourceName+"."+resourceId, "id"),
				),
			},
		},
	})
}

// Test that published scripts can also return hard-coded default scripts
func TestAccDataSourceScriptPublishedDefaults(t *testing.T) {
	const (
		callbackDataSource      = "callback-script-data"
		defaultCallbackScriptId = "ffde0662-8395-9b04-7dcb-b90172109065"

		inboundDataSource      = "inbound-script-data"
		defaultInboundScriptId = "766f1221-047a-11e5-bba2-db8c0964d007"

		outboundDataSource      = "outbound-script-data"
		defaultOutboundScriptId = "476c2b71-7429-11e4-9a5b-3f91746bffa3"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptDataSource(
					callbackDataSource,
					constants.DefaultCallbackScriptName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.%s", resourceName, callbackDataSource), "id",
						defaultCallbackScriptId,
					),
				),
			},
			{
				Config: generateScriptDataSource(
					inboundDataSource,
					constants.DefaultInboundScriptName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.%s", resourceName, inboundDataSource), "id",
						defaultInboundScriptId,
					),
				),
			},
			{
				Config: generateScriptDataSource(
					outboundDataSource,
					constants.DefaultOutboundScriptName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.%s", resourceName, outboundDataSource), "id",
						defaultOutboundScriptId,
					),
				),
			},
		},
	})
}

func generateScriptDataSource(dataSourceID, name, resourceId string) string {
	if resourceId != "" {
		return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on = [%s.%s]
	}
	`, resourceName, dataSourceID, name, resourceName, resourceId)
	} else {
		return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
	}
	`, resourceName, dataSourceID, name)
	}
}
