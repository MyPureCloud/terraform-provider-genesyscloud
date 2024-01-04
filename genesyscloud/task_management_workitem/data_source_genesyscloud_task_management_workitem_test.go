package task_management_workitem

import (
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the task management workitem Data Source
*/

func TestAccDataSourceTaskManagementWorkitem(t *testing.T) {
	t.Parallel()
	var ()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps:             []resource.TestStep{},
	})
}
