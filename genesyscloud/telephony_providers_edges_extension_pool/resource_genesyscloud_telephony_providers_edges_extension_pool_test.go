package telephony_providers_edges_extension_pool

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceExtensionPoolBasic(t *testing.T) {
	t.Parallel()
	extensionPoolResource1 := "test-extensionpool1"
	extensionPoolStartNumber1 := "15000"
	extensionPoolEndNumber1 := "15001"
	_, err := provider.AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	DeleteExtensionPoolWithNumber(extensionPoolStartNumber1)
	DeleteExtensionPoolWithNumber(extensionPoolEndNumber1)
	extensionPoolDescription1 := "Test description"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateExtensionPoolResource(&ExtensionPoolStruct{
					extensionPoolResource1,
					extensionPoolStartNumber1,
					extensionPoolEndNumber1,
					util.NullValue, // No description
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolResource1, "start_number", extensionPoolStartNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolResource1, "end_number", extensionPoolEndNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolResource1, "description", ""),
				),
			},
			{
				// Update
				Config: GenerateExtensionPoolResource(&ExtensionPoolStruct{
					extensionPoolResource1,
					extensionPoolStartNumber1,
					extensionPoolEndNumber1,
					strconv.Quote(extensionPoolDescription1),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolResource1, "start_number", extensionPoolStartNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolResource1, "end_number", extensionPoolEndNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolResource1, "description", extensionPoolDescription1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_extension_pool." + extensionPoolResource1,
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
