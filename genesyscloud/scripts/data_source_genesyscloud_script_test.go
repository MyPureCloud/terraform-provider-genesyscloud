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
		scriptDataSourceLabel = "script-data"
		resourceLabel         = "script"
		name                  = "tfscript" + uuid.NewString()
		filePath              = getTestDataPath("resource", ResourceType, "test_script.json")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptResource(
					resourceLabel,
					name,
					filePath,
					"",
				) + generateScriptDataSource(
					scriptDataSourceLabel,
					name,
					resourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(fmt.Sprintf("data.%s.%s", ResourceType, scriptDataSourceLabel), "id",
						ResourceType+"."+resourceLabel, "id"),
				),
			},
		},
	})
}

// Test that published scripts can also return hard-coded default scripts
func TestAccDataSourceScriptPublishedDefaults(t *testing.T) {
	const (
		callbackDataSourceLabel = "callback-script-data"
		defaultCallbackScriptId = "ffde0662-8395-9b04-7dcb-b90172109065"

		inboundDataSourceLabel = "inbound-script-data"
		defaultInboundScriptId = "766f1221-047a-11e5-bba2-db8c0964d007"

		outboundDataSourceLabel = "outbound-script-data"
		defaultOutboundScriptId = "476c2b71-7429-11e4-9a5b-3f91746bffa3"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptDataSource(
					callbackDataSourceLabel,
					constants.DefaultCallbackScriptName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.%s", ResourceType, callbackDataSourceLabel), "id",
						defaultCallbackScriptId,
					),
				),
			},
			{
				Config: generateScriptDataSource(
					inboundDataSourceLabel,
					constants.DefaultInboundScriptName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.%s", ResourceType, inboundDataSourceLabel), "id",
						defaultInboundScriptId,
					),
				),
			},
			{
				Config: generateScriptDataSource(
					outboundDataSourceLabel,
					constants.DefaultOutboundScriptName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.%s", ResourceType, outboundDataSourceLabel), "id",
						defaultOutboundScriptId,
					),
				),
			},
		},
	})
}

func generateScriptDataSource(dataSourceLabel, name, resourceLabel string) string {
	if resourceLabel != "" {
		return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on = [%s.%s]
	}
	`, ResourceType, dataSourceLabel, name, ResourceType, resourceLabel)
	} else {
		return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
	}
	`, ResourceType, dataSourceLabel, name)
	}
}
