package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDidPoolBasic(t *testing.T) {
	var (
		didPoolStartPhoneNumber = "+45465550001"
		didPoolEndPhoneNumber   = "+45465550002"
		didPoolRes              = "didPool"
		didPoolDataRes          = "didPoolData"
	)

	if _, err := AuthorizeSdk(); err != nil {
		t.Fatal(err)
	}
	if err := deleteDidPoolWithNumber(didPoolStartPhoneNumber); err != nil {
		t.Fatalf("error deleting did pool start number: %v", err)
	}
	if err := deleteDidPoolWithNumber(didPoolEndPhoneNumber); err != nil {
		t.Fatalf("error deleting did pool end number: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateDidPoolResource(&didPoolStruct{
					didPoolRes,
					didPoolStartPhoneNumber,
					didPoolEndPhoneNumber,
					nullValue, // No description
					nullValue, // No comments
					nullValue, // No provider
				}) + generateDidPoolDataSource(didPoolDataRes,
					didPoolStartPhoneNumber,
					didPoolEndPhoneNumber,
					"genesyscloud_telephony_providers_edges_did_pool."+didPoolRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_telephony_providers_edges_did_pool."+didPoolDataRes, "id", "genesyscloud_telephony_providers_edges_did_pool."+didPoolRes, "id"),
				),
			},
		},
	})
}

func generateDidPoolDataSource(
	resourceID string,
	startPhoneNumber string,
	endPhoneNumber string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_did_pool" "%s" {
		start_phone_number = "%s"
		end_phone_number = "%s"
		depends_on=[%s]
	}
	`, resourceID, startPhoneNumber, endPhoneNumber, dependsOnResource)
}
