package telephony_providers_edges_did

import (
	"context"
	"fmt"
	archIvr "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	didPool "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDidBasic(t *testing.T) {
	var (
		didPoolStartPhoneNumber = "+14546555001"
		didPoolEndPhoneNumber   = "+14546555003"
		didPoolResourceLabel    = "didPool"
		ivrConfigResourceLabel  = "ivrConfig"
		ivrConfigName           = "test-config" + uuid.NewString()
		ivrConfigDnis           = []string{"+14546555002"}
		didPhoneNumber          = "+14546555002"
		didDataResourceLabel    = "didData"
	)

	// did pool cleanup
	resp, err := didPool.DeleteDidPoolWithStartAndEndNumber(context.Background(), didPoolStartPhoneNumber, didPoolEndPhoneNumber, sdkConfig)
	if err != nil {
		respStr := "<nil>"
		if resp != nil {
			respStr = strconv.Itoa(resp.StatusCode)
		}
		t.Logf("Failed to delete DID pool: %s. API Response: %s", err.Error(), respStr)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: didPool.GenerateDidPoolResource(&didPool.DidPoolStruct{
					ResourceLabel:    didPoolResourceLabel,
					StartPhoneNumber: didPoolStartPhoneNumber,
					EndPhoneNumber:   didPoolEndPhoneNumber,
					Description:      util.NullValue, // No description
					Comments:         util.NullValue, // No comments
					PoolProvider:     util.NullValue, // No provider
				}) + archIvr.GenerateIvrConfigResource(&archIvr.IvrConfigStruct{
					ResourceLabel: ivrConfigResourceLabel,
					Name:          ivrConfigName,
					Description:   "",
					Dnis:          ivrConfigDnis,
					DependsOn:     "genesyscloud_telephony_providers_edges_did_pool." + didPoolResourceLabel,
				}) + generateDidDataSource(didDataResourceLabel,
					didPhoneNumber,
					"genesyscloud_architect_ivr."+ivrConfigResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data."+ResourceType+"."+didDataResourceLabel, "phone_number", didPhoneNumber),
				),
			},
		},
	})
}

func generateDidDataSource(
	resourceLabel string,
	phoneNumber string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		phone_number = "%s"
		depends_on=[%s]
	}
	`, ResourceType, resourceLabel, phoneNumber, dependsOnResource)
}
