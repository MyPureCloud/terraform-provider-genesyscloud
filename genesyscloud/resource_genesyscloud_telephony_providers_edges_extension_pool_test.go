package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

type extensionPoolStruct struct {
	resourceID  string
	startNumber string
	endNumber   string
	description string
}

func TestAccResourceExtensionPoolBasic(t *testing.T) {
	t.Parallel()
	extensionPoolResource1 := "test-extensionpool1"
	extensionPoolStartNumber1 := "15000"
	extensionPoolEndNumber1 := "15001"
	_, err := AuthorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	deleteExtensionPoolWithNumber(extensionPoolStartNumber1)
	deleteExtensionPoolWithNumber(extensionPoolEndNumber1)

	extensionPoolDescription1 := "Test description"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateExtensionPoolResource(&extensionPoolStruct{
					extensionPoolResource1,
					extensionPoolStartNumber1,
					extensionPoolEndNumber1,
					nullValue, // No description
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolResource1, "start_number", extensionPoolStartNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolResource1, "end_number", extensionPoolEndNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_extension_pool."+extensionPoolResource1, "description", ""),
				),
			},
			{
				// Update
				Config: generateExtensionPoolResource(&extensionPoolStruct{
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

func deleteExtensionPoolWithNumber(startNumber string) error {
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		extensionPools, _, getErr := edgesAPI.GetTelephonyProvidersEdgesExtensionpools(100, pageNum, "", "")
		if getErr != nil {
			return getErr
		}

		if extensionPools.Entities == nil || len(*extensionPools.Entities) == 0 {
			break
		}

		for _, extensionPool := range *extensionPools.Entities {
			if extensionPool.StartNumber != nil && *extensionPool.StartNumber == startNumber {
				_, err := edgesAPI.DeleteTelephonyProvidersEdgesExtensionpool(*extensionPool.Id)
				time.Sleep(20 * time.Second)
				return err
			}
		}
	}

	return nil
}

func generateExtensionPoolResource(extensionPool *extensionPoolStruct) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_extension_pool" "%s" {
		start_number = "%s"
		end_number = "%s"
		description = %s
	}
	`, extensionPool.resourceID,
		extensionPool.startNumber,
		extensionPool.endNumber,
		extensionPool.description)
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

		if IsStatus404(resp) {
			// Extension pool not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All Extension pool destroyed
	return nil
}
