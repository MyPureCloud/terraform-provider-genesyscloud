package telephony_providers_edges_did_pool

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDidPoolBasic(t *testing.T) {
	var (
		didPoolStartPhoneNumber  = "+14546555007"
		didPoolEndPhoneNumber    = "+14546555008"
		didPoolResourceLabel     = "didPool"
		didPoolDataResourceLabel = "didPoolData"

		resourceFullPath   = ResourceType + "." + didPoolResourceLabel
		dataSourceFullPath = "data." + ResourceType + "." + didPoolDataResourceLabel
	)

	// did pool cleanup
	resp, err := DeleteDidPoolWithStartAndEndNumber(context.Background(), didPoolStartPhoneNumber, didPoolEndPhoneNumber, sdkConfig)
	if err != nil {
		respStr := "<nil>"
		if resp != nil {
			respStr = strconv.Itoa(resp.StatusCode)
		}
		t.Logf("Failed to delete did pool: %s. API Response: %s", err.Error(), respStr)
	}

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
					resource.TestCheckResourceAttrPair(dataSourceFullPath, "id", resourceFullPath, "id"),
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
