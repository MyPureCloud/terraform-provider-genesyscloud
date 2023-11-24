package task_management_worktype

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_task_management_worktype_test.go contains all of the test cases for running the resource
tests for task_management_worktype.
*/

func TestAccResourceTaskManagementWorktype(t *testing.T) {
	t.Parallel()
	var ()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps:             []resource.TestStep{},
		CheckDestroy:      testVerifyTaskManagementWorktypeDestroyed,
	})
}

func testVerifyTaskManagementWorktypeDestroyed(state *terraform.State) error {
	return nil
}
