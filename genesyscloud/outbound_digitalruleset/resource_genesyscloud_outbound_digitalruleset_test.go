package outbound_digitalruleset

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_outbound_digitalruleset_test.go contains all of the test cases for running the resource
tests for outbound_digitalruleset.
*/

func TestAccResourceOutboundDigitalruleset(t *testing.T) {
	t.Parallel()
	var ()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps:             []resource.TestStep{},
		CheckDestroy:      testVerifyOutboundDigitalrulesetDestroyed,
	})
}

func testVerifyOutboundDigitalrulesetDestroyed(state *terraform.State) error {
	return nil
}
