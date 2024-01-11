package architect_ivr

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceArchitectIvr(t *testing.T) {
	t.Parallel()
	var (
		ivrResource = "arch-ivr"
		name        = "IVR " + uuid.NewString()
		description = "Sample IVR by CX as Code"

		ivrDataSource = "arch-ivr-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateIvrConfigResource(&IvrConfigStruct{
					ResourceID:  ivrResource,
					Name:        name,
					Description: description,
					Dnis:        nil,
					DependsOn:   "",
				}) + GenerateIvrDataSource(ivrDataSource,
					resourceName+"."+ivrResource+".name",
					resourceName+"."+ivrResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+resourceName+"."+ivrDataSource, "id", resourceName+"."+ivrResource, "id"),
				),
			},
		},
	})
}

/*
This is a unit test to test whether the Architect IVR data source is properly pulling the id back from the proxy
*/
func TestDataSourceArchitectIvr(t *testing.T) {
	targetId := uuid.NewString()
	targetName := "MyTargetId"
	archProxy := &architectIvrProxy{}
	archProxy.getArchitectIvrIdByNameAttr = func(ctx context.Context, a *architectIvrProxy, name string) (string, bool, error) {
		assert.Equal(t, targetName, name)
		return targetId, false, nil
	}
	internalProxy = archProxy
	ctx := context.Background()
	gcloud := &gcloud.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

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
