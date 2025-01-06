package telephony_providers_edges_did_pool

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDidPoolBasic(t *testing.T) {
	var (
		didPoolStartPhoneNumber  = "+45465550007"
		didPoolEndPhoneNumber    = "+45465550008"
		didPoolResourceLabel     = "didPool"
		didPoolDataResourceLabel = "didPoolData"
	)

	// did pool cleanup
	defer func() {
		if _, err := provider.AuthorizeSdk(); err != nil {
			return
		}
		ctx := context.TODO()
		_, _ = DeleteDidPoolWithStartAndEndNumber(ctx, didPoolStartPhoneNumber, didPoolEndPhoneNumber)
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateDidPoolResource(&DidPoolStruct{
					didPoolResourceLabel,
					didPoolStartPhoneNumber,
					didPoolEndPhoneNumber,
					util.NullValue, // No description
					util.NullValue, // No comments
					util.NullValue, // No provider
				}) + generateDidPoolDataSource(didPoolDataResourceLabel,
					didPoolStartPhoneNumber,
					didPoolEndPhoneNumber,
					ResourceType+"."+didPoolResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+didPoolDataResourceLabel, "id", "genesyscloud_telephony_providers_edges_did_pool."+didPoolResourceLabel, "id"),
				),
			},
		},
	})
}

func generateDidPoolDataSource(
	resourceLabel string,
	startPhoneNumber string,
	endPhoneNumber string,
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		start_phone_number = "%s"
		end_phone_number   = "%s"
		depends_on         = [%s]
	}
	`, ResourceType, resourceLabel, startPhoneNumber, endPhoneNumber, dependsOnResource)
}
