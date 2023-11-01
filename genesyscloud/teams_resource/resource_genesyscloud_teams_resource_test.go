package teams_resource

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

/*
The resource_genesyscloud_teams_resource_test.go contains all of the test cases for running the resource
tests for teams_resource.
*/

func TestAccResourceTeamsResource(t *testing.T) {
	t.Parallel()
	var ()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps:             []resource.TestStep{},
		CheckDestroy:      testVerifyTeamsResourceDestroyed,
	})
}

func testVerifyTeamsResourceDestroyed(state *terraform.State) error {
	return nil
}
