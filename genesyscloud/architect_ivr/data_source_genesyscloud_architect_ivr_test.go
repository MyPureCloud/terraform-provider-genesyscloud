package architect_ivr

import (
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceArchitectIvr(t *testing.T) {
	t.Parallel()
	var (
		ivrResourceLabel = "arch-ivr"
		name             = "IVR " + uuid.NewString()
		description      = "Sample IVR by CX as Code"

		ivrDataSourceLabel = "arch-ivr-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceLabel: ivrResourceLabel,
					Name:          name,
					Description:   description,
					Dnis:          nil,
					DependsOn:     "",
				}) + GenerateIvrDataSource(ivrDataSourceLabel,
					ResourceType+"."+ivrResourceLabel+".name",
					ResourceType+"."+ivrResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+ivrDataSourceLabel, "id", ResourceType+"."+ivrResourceLabel, "id"),
				),
			},
		},
	})
}

/*
This is a unit test to test whether the Architect IVR data source is properly pulling the id back from the proxy
*/
func TestUnitDataSourceArchitectIvr(t *testing.T) {
	targetId := uuid.NewString()
	targetName := "MyTargetId"
	archProxy := &architectIvrProxy{}
	archProxy.getArchitectIvrIdByNameAttr = func(ctx context.Context, a *architectIvrProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
		assert.Equal(t, targetName, name)
		return targetId, false, nil, nil
	}
	internalProxy = archProxy
	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := DataSourceArchitectIvr().Schema

	//Setup a map of values
	resourceDataMap := map[string]interface{}{
		"Id":   targetId,
		"name": targetName,
	}

	//Found this gem TestResourceDataRaw here: https://github.com/hashicorp/terraform/issues/890
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	dataSourceIvrRead(ctx, d, gcloud)
	assert.Equal(t, targetId, d.Id())

	defer func() { internalProxy = nil }()
}
