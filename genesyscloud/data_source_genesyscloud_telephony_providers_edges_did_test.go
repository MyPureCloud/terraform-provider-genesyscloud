package genesyscloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDidBasic(t *testing.T) {
	t.Parallel()
	var (
		didPoolStartPhoneNumber = "+45465550001"
		didPoolEndPhoneNumber   = "+45465550333"
		didPoolRes              = "didPool"
		ivrConfigRes            = "ivrConfig"
		ivrConfigName           = "test-config" + uuid.NewString()
		ivrConfigDnis           = []string{"+45465550002"}
		didPhoneNumber          = "+45465550002"
		didDataRes              = "didData"
	)

	err := authorizeSdk()
	if err != nil {
		t.Fatal(err)
	}
	deleteDidPoolWithNumber(didPoolStartPhoneNumber)
	deleteDidPoolWithNumber(didPoolEndPhoneNumber)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
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
				}) + generateIvrConfigResource(&ivrConfigStruct{
					ivrConfigRes,
					ivrConfigName,
					"",
					ivrConfigDnis,
					"genesyscloud_telephony_providers_edges_did_pool." + didPoolRes,
				}) + generateDidDataSource(didDataRes,
					didPhoneNumber,
					"genesyscloud_architect_ivr." + ivrConfigRes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_telephony_providers_edges_did."+didDataRes, "phone_number", didPhoneNumber),
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
	return fmt.Sprintf(`data "genesyscloud_telephony_providers_edges_did" "%s" {
		phone_number = "%s"
		depends_on=[%s]
	}
	`, resourceID, phoneNumber, dependsOnResource)
}
