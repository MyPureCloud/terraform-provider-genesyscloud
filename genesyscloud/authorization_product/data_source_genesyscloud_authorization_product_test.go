package authorization_product

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAuthorizationProduct(t *testing.T) {

	var (
		productName  = "botFlows"
		dataSourceId = productName
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateAuthorizationProductDataSource(
					dataSourceId,
					productName,
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.genesyscloud_authorization_product."+dataSourceId, "name", productName),
				),
			},
		},
	})
}

/** Unit Test **/
func TestUnitDataSourceAuthorizationProduct(t *testing.T) {
	tId := uuid.NewString()
	authProxy := &authProductProxy{}

	authProxy.getAuthorizationProductAttr = func(ctx context.Context, a *authProductProxy, name string) (id string, retry bool, err error) {
		return name, false, nil
	}
	internalProxy = authProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	dataSourceSchema := DataSourceAuthorizationProduct().Schema

	//Setup a map of values
	dataSourceDataMap := buildDataSourceAuthProductMap(tId)

	d := schema.TestResourceDataRaw(t, dataSourceSchema, dataSourceDataMap)
	d.SetId(tId)

	diag := dataSourceAuthorizationProductRead(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())

}

func buildDataSourceAuthProductMap(tId string) map[string]interface{} {
	dataSourceDataMap := map[string]interface{}{
		"name": tId,
	}
	return dataSourceDataMap
}
