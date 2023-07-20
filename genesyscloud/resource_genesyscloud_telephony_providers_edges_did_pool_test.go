package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

type didPoolStruct struct {
	resourceID       string
	startPhoneNumber string
	endPhoneNumber   string
	description      string
	comments         string
	poolProvider     string
}

func TestAccResourceDidPoolBasic(t *testing.T) {
	didPoolResource1 := "test-didpool1"
	didPoolStartPhoneNumber1 := "+14175540014"
	didPoolEndPhoneNumber1 := "+14175540015"
	if _, err := AuthorizeSdk(); err != nil {
		t.Fatal(err)
	}
	if err := deleteDidPoolWithNumber(didPoolStartPhoneNumber1); err != nil {
		t.Fatalf("error deleting did pool start number: %v", err)
	}
	if err := deleteDidPoolWithNumber(didPoolEndPhoneNumber1); err != nil {
		t.Fatalf("error deleting did pool end number: %v", err)
	}

	didPoolDescription1 := "Test description"
	didPoolComments1 := "Test comments"
	didPoolProvider1 := "PURE_CLOUD"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateDidPoolResource(&didPoolStruct{
					didPoolResource1,
					didPoolStartPhoneNumber1,
					didPoolEndPhoneNumber1,
					nullValue, // No description
					nullValue, // No comments
					nullValue, // No provider
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_did_pool."+didPoolResource1, "start_phone_number", didPoolStartPhoneNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_did_pool."+didPoolResource1, "end_phone_number", didPoolEndPhoneNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_did_pool."+didPoolResource1, "description", ""),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_did_pool."+didPoolResource1, "comments", ""),
				),
			},
			{
				// Update
				Config: generateDidPoolResource(&didPoolStruct{
					didPoolResource1,
					didPoolStartPhoneNumber1,
					didPoolEndPhoneNumber1,
					strconv.Quote(didPoolDescription1),
					strconv.Quote(didPoolComments1),
					strconv.Quote(didPoolProvider1),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_did_pool."+didPoolResource1, "start_phone_number", didPoolStartPhoneNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_did_pool."+didPoolResource1, "end_phone_number", didPoolEndPhoneNumber1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_did_pool."+didPoolResource1, "description", didPoolDescription1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_did_pool."+didPoolResource1, "comments", didPoolComments1),
					resource.TestCheckResourceAttr("genesyscloud_telephony_providers_edges_did_pool."+didPoolResource1, "pool_provider", didPoolProvider1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_telephony_providers_edges_did_pool." + didPoolResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyDidPoolsDestroyed,
	})
}

func deleteDidPoolWithStartNumber(startNumber string) error {
	edgesAPI := platformclientv2.NewTelephonyProvidersEdgeApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		didPools, _, getErr := edgesAPI.GetTelephonyProvidersEdgesDidpools(100, pageNum, "", nil)
		if getErr != nil {
			return getErr
		}

		if didPools.Entities == nil || len(*didPools.Entities) == 0 {
			break
		}

		for _, didPool := range *didPools.Entities {
			if didPool.StartPhoneNumber != nil && *didPool.StartPhoneNumber == startNumber {
				_, err := edgesAPI.DeleteTelephonyProvidersEdgesDidpool(*didPool.Id)
				time.Sleep(20 * time.Second)
				return err
			}
		}
	}

	return nil
}

func generateDidPoolResource(didPool *didPoolStruct) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_did_pool" "%s" {
		start_phone_number = "%s"
		end_phone_number = "%s"
		description = %s
		comments = %s
		pool_provider = %s
	}
	`, didPool.resourceID,
		didPool.startPhoneNumber,
		didPool.endPhoneNumber,
		didPool.description,
		didPool.comments,
		didPool.poolProvider)
}

func testVerifyDidPoolsDestroyed(state *terraform.State) error {
	telephonyAPI := platformclientv2.NewTelephonyProvidersEdgeApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_telephony_providers_edges_did_pool" {
			continue
		}

		didPool, resp, err := telephonyAPI.GetTelephonyProvidersEdgesDidpool(rs.Primary.ID)
		if didPool != nil && didPool.State != nil && *didPool.State == "deleted" {
			continue
		}

		if didPool != nil {
			return fmt.Errorf("DID Pool (%s) still exists", rs.Primary.ID)
		}

		if IsStatus404(resp) {
			// DID pool not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All DID pool destroyed
	return nil
}
