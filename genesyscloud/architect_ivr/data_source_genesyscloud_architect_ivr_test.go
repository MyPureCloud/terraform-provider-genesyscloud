package architect_ivr

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"testing"

	"github.com/google/uuid"

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
