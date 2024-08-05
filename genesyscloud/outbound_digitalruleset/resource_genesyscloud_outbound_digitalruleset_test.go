package outbound_digitalruleset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

/*
The resource_genesyscloud_outbound_digitalruleset_test.go contains all of the test cases for running the resource
tests for outbound_digitalruleset.
*/

func TestAccResourceOutboundDigitalruleset(t *testing.T) {
	t.Parallel()
	var ()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps:             []resource.TestStep{},
		CheckDestroy:      testVerifyOutboundDigitalrulesetDestroyed,
	})
}

func testVerifyOutboundDigitalrulesetDestroyed(state *terraform.State) error {
	return nil
}
