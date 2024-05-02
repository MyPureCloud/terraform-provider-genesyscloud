package scripts

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"

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
	var (
		callbackDataSource = "callback-script-data"
		callbackName       = "Default Callback Script"
		callbackId         = "1a5ab5f6-a967-4010-8c54-bd88f092e5a8"
		inboundDataSource  = "inbound-script-data"
		inboundName        = "Default Inbound Script"
		inboundId          = "28bfd948-427f-4956-947e-600ad17ced68"
		outboundDataSource = "outbound-script-data"
		outboundName       = "Default Outbound Script"
		outboundId         = "e86ac8b2-fdbc-4d85-8c6c-16da0a6493a0"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateScriptDataSource(
					callbackDataSource,
					callbackName,
					"",
				),
				PreConfig: func() {
					// Wait for a specified duration for it to get created properly
					time.Sleep(20 * time.Second)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.%s", resourceName, callbackDataSource), "id",
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
				PreConfig: func() {
					// Wait for a specified duration  for it to get created properly
					time.Sleep(20 * time.Second)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.%s", resourceName, inboundDataSource), "id",
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
				PreConfig: func() {
					// Wait for a specified duration for it to get created properly
					time.Sleep(20 * time.Second)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.%s", resourceName, outboundDataSource), "id",
						outboundId,
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
