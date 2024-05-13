package scripts

import (
	"fmt"
	"os"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
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
		callbackDataSource string
		callbackName       string
		callbackId         string
		inboundDataSource  string
		inboundName        string
		inboundId          string
		outboundDataSource string
		outboundName       string
		outboundId         string
	)
	if v := os.Getenv("GENESYSCLOUD_REGION"); v == "tca" {
		callbackDataSource = "callback-script-data"
		callbackName = "Default Callback Script"
		callbackId = "ec2275b0-126b-46c1-a4f4-5524c2e91c9d"
		inboundDataSource = "inbound-script-data"
		inboundName = "Default Inbound Script"
		inboundId = "77832ba6-e02f-4bdc-8200-b7e7043b756e"
		outboundDataSource = "outbound-script-data"
		outboundName = "Default Outbound Script"
		outboundId = "055be875-2f9e-4fa0-a331-2cdc690c20f9"
	} else if v == "us-east-1" {
		callbackDataSource = "callback-script-data"
		callbackName = "Default Callback Script"
		callbackId = "1a5ab5f6-a967-4010-8c54-bd88f092e5a8"
		inboundDataSource = "inbound-script-data"
		inboundName = "Default Inbound Script"
		inboundId = "28bfd948-427f-4956-947e-600ad17ced68"
		outboundDataSource = "outbound-script-data"
		outboundName = "Default Outbound Script"
		outboundId = "b82ddcec-6941-4831-9dca-9b6be19fa1ff"
	}

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
				PreConfig: func() {
					// Wait for a specified duration for it to get created properly
					time.Sleep(30 * time.Second)
				},
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
				PreConfig: func() {
					// Wait for a specified duration  for it to get created properly
					time.Sleep(30 * time.Second)
				},
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
				PreConfig: func() {
					// Wait for a specified duration for it to get created properly
					time.Sleep(30 * time.Second)
				},
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
