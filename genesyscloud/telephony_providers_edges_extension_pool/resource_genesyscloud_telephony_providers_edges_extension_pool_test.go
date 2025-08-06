package telephony_providers_edges_extension_pool

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceExtensionPoolBasic(t *testing.T) {
	t.Parallel()
	extensionPoolResourceLabel1 := "test-extensionpool1"
	extensionPoolResourcePath := ResourceType + "." + extensionPoolResourceLabel1
	extensionPoolStartNumber1 := "15000"
	extensionPoolEndNumber1 := "15001"
	extensionPoolDescription1 := "Test description"

	cleanupExtensionPool(t, extensionPoolStartNumber1, extensionPoolEndNumber1)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateExtensionPoolResource(&ExtensionPoolStruct{
					extensionPoolResourceLabel1,
					extensionPoolStartNumber1,
					extensionPoolEndNumber1,
					util.NullValue, // No description
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(extensionPoolResourcePath, "start_number", extensionPoolStartNumber1),
					resource.TestCheckResourceAttr(extensionPoolResourcePath, "end_number", extensionPoolEndNumber1),
					resource.TestCheckResourceAttr(extensionPoolResourcePath, "description", ""),
				),
			},
			{
				// Update
				Config: GenerateExtensionPoolResource(&ExtensionPoolStruct{
					extensionPoolResourceLabel1,
					extensionPoolStartNumber1,
					extensionPoolEndNumber1,
					strconv.Quote(extensionPoolDescription1),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(extensionPoolResourcePath, "start_number", extensionPoolStartNumber1),
					resource.TestCheckResourceAttr(extensionPoolResourcePath, "end_number", extensionPoolEndNumber1),
					resource.TestCheckResourceAttr(extensionPoolResourcePath, "description", extensionPoolDescription1),
				),
			},
			{
				// Import/Read
				ResourceName:      extensionPoolResourcePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyExtensionPoolsDestroyed,
	})
}

func testVerifyExtensionPoolsDestroyed(state *terraform.State) error {
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_telephony_providers_edges_extension_pool" {
			continue
		}

		extensionPool, resp, err := telephonyAPI.GetTelephonyProvidersEdgesExtensionpool(rs.Primary.ID)
		if extensionPool != nil && extensionPool.State != nil && *extensionPool.State == "deleted" {
			continue
		}

		if extensionPool != nil {
			return fmt.Errorf("Extension Pool (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// Extension pool not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All Extension pool destroyed
	return nil
}

func cleanupExtensionPool(t *testing.T, startNumber, endNumber string) {
	if err := DeleteExtensionPoolWithNumber(startNumber); err != nil {
		t.Logf("Failed to delete extension pool: %s", err.Error())
	}

	if err := DeleteExtensionPoolWithNumber(endNumber); err != nil {
		t.Logf("Failed to delete extension pool: %s", err.Error())
	}
}
