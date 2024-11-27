package telephony_providers_edges_did_pool

import (
	"context"
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccResourceDidPoolBasic(t *testing.T) {
	didPoolResourceLabel1 := "test-didpool1"
	didPoolStartPhoneNumber1 := "+14175540014"
	didPoolEndPhoneNumber1 := "+14175540015"

	// did pool cleanup
	defer func() {
		if _, err := provider.AuthorizeSdk(); err != nil {
			return
		}
		ctx := context.TODO()
		_, _ = DeleteDidPoolWithStartAndEndNumber(ctx, didPoolStartPhoneNumber1, didPoolEndPhoneNumber1)
	}()

	didPoolDescription1 := "Test description"
	didPoolComments1 := "Test comments"
	didPoolProvider1 := "PURE_CLOUD"

	resourcePath := ResourceType + "." + didPoolResourceLabel1

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateDidPoolResource(&DidPoolStruct{
					didPoolResourceLabel1,
					didPoolStartPhoneNumber1,
					didPoolEndPhoneNumber1,
					util.NullValue, // No description
					util.NullValue, // No comments
					util.NullValue, // No provider
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "start_phone_number", didPoolStartPhoneNumber1),
					resource.TestCheckResourceAttr(resourcePath, "end_phone_number", didPoolEndPhoneNumber1),
					resource.TestCheckResourceAttr(resourcePath, "description", ""),
					resource.TestCheckResourceAttr(resourcePath, "comments", ""),
				),
			},
			{
				// Update
				Config: GenerateDidPoolResource(&DidPoolStruct{
					didPoolResourceLabel1,
					didPoolStartPhoneNumber1,
					didPoolEndPhoneNumber1,
					strconv.Quote(didPoolDescription1),
					strconv.Quote(didPoolComments1),
					strconv.Quote(didPoolProvider1),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "start_phone_number", didPoolStartPhoneNumber1),
					resource.TestCheckResourceAttr(resourcePath, "end_phone_number", didPoolEndPhoneNumber1),
					resource.TestCheckResourceAttr(resourcePath, "description", didPoolDescription1),
					resource.TestCheckResourceAttr(resourcePath, "comments", didPoolComments1),
					resource.TestCheckResourceAttr(resourcePath, "pool_provider", didPoolProvider1),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + didPoolResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyDidPoolsDestroyed,
	})
}

func testVerifyDidPoolsDestroyed(state *terraform.State) error {
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		didPool, resp, err := telephonyAPI.GetTelephonyProvidersEdgesDidpool(rs.Primary.ID)
		if didPool != nil && didPool.State != nil && *didPool.State == "deleted" {
			continue
		}

		if didPool != nil {
			return fmt.Errorf("DID Pool (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// DID pool not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All DID pool destroyed
	return nil
}
