package outbound_digitalruleset

import (
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
Test Class for the outbound digitalruleset Data Source
*/

func TestAccDataSourceOutboundDigitalruleset(t *testing.T) {
	t.Parallel()
	var ()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps:             []resource.TestStep{},
	})
}
