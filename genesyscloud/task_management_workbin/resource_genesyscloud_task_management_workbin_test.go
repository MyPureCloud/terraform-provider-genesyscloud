package task_management_workbin

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_task_management_workbin_test.go contains all of the test cases for running the resource
tests for task_management_workbin.
*/

type testWorkbinResource struct {
	name        string
	divisionId  string
	description string
}

func TestAccResourceTaskManagementWorkbin(t *testing.T) {
	t.Parallel()
	var ()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps:             []resource.TestStep{},
		CheckDestroy:      testVerifyTaskManagementWorkbinDestroyed,
	})
}

func testVerifyTaskManagementWorkbinDestroyed(state *terraform.State) error {
	return nil
}

func generateWorkbinResource(resourceId string, name string, description string, divisionIdRef string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		description = "%s"
		division_id = %s
	}
	`, resourceName, resourceId, name, description, divisionIdRef)
}
