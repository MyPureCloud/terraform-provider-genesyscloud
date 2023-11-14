package telephony_providers_edges_did

import (
	"context"
	"fmt"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	archIvr "terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	didPool "terraform-provider-genesyscloud/genesyscloud/telephony_providers_edges_did_pool"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDidBasic(t *testing.T) {
	var (
		didPoolStartPhoneNumber = "+45465550001"
		didPoolEndPhoneNumber   = "+45465550003"
		didPoolRes              = "didPool"
		ivrConfigRes            = "ivrConfig"
		ivrConfigName           = "test-config" + uuid.NewString()
		ivrConfigDnis           = []string{"+45465550002"}
		didPhoneNumber          = "+45465550002"
		didDataRes              = "didData"
	)

	// did pool cleanup
	defer func() {
		if _, err := gcloud.AuthorizeSdk(); err != nil {
			return
		}
		ctx := context.TODO()
		_ = didPool.DeleteDidPoolWithStartAndEndNumber(ctx, didPoolStartPhoneNumber, didPoolEndPhoneNumber)
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: didPool.GenerateDidPoolResource(&didPool.DidPoolStruct{
					ResourceID:       didPoolRes,
					StartPhoneNumber: didPoolStartPhoneNumber,
					EndPhoneNumber:   didPoolEndPhoneNumber,
					Description:      gcloud.NullValue, // No description
					Comments:         gcloud.NullValue, // No comments
					PoolProvider:     gcloud.NullValue, // No provider
				}) + archIvr.GenerateIvrConfigResource(&archIvr.IvrConfigStruct{
					ResourceID:  ivrConfigRes,
					Name:        ivrConfigName,
					Description: "",
					Dnis:        ivrConfigDnis,
					DependsOn:   "genesyscloud_telephony_providers_edges_did_pool." + didPoolRes,
				}) + generateDidDataSource(didDataRes,
					didPhoneNumber,
					"genesyscloud_architect_ivr."+ivrConfigRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data."+resourceName+"."+didDataRes, "phone_number", didPhoneNumber),
				),
			},
		},
	})
}

func generateDidDataSource(
	resourceID string,
	phoneNumber string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		phone_number = "%s"
		depends_on=[%s]
	}
	`, resourceName, resourceID, phoneNumber, dependsOnResource)
}
