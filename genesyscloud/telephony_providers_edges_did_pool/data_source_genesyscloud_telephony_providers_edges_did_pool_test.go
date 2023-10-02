package telephony_providers_edges_did_pool

import (
	"context"
	"fmt"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
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

	// did pool cleanup
	defer func() {
		if _, err := gcloud.AuthorizeSdk(); err != nil {
			return
		}
		ctx := context.TODO()
		_ = DeleteDidPoolWithStartAndEndNumber(ctx, didPoolStartPhoneNumber, didPoolEndPhoneNumber)
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateDidPoolResource(&DidPoolStruct{
					didPoolRes,
					didPoolStartPhoneNumber,
					didPoolEndPhoneNumber,
					nullValue, // No description
					nullValue, // No comments
					nullValue, // No provider
				}) + generateDidPoolDataSource(didPoolDataRes,
					didPoolStartPhoneNumber,
					didPoolEndPhoneNumber,
					resourceName+"."+didPoolRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+resourceName+"."+didPoolDataRes, "id", "genesyscloud_telephony_providers_edges_did_pool."+didPoolRes, "id"),
				),
			},
		},
	})
}

func generateDidPoolDataSource(
	resourceID string,
	startPhoneNumber string,
	endPhoneNumber string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		start_phone_number = "%s"
		end_phone_number   = "%s"
		depends_on         = [%s]
	}
	`, resourceName, resourceID, startPhoneNumber, endPhoneNumber, dependsOnResource)
}
