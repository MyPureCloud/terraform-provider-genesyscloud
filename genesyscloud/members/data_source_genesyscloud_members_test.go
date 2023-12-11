package members

import (
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the members Data Source
*/

func TestAccDataSourceMembers(t *testing.T) {
	t.Parallel()
	var ()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps:             []resource.TestStep{},
	})
}
