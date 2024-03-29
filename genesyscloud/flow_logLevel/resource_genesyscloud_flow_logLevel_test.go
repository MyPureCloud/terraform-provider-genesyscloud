package flow_logLevel

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

func TestAccResourceFlowLogLevel(t *testing.T) {
	var (
		resourceId            = "flow_log_level" + uuid.NewString()
		communications        = false
		eventError            = false
		eventOther            = false
		eventWarning          = false
		executionInputOutputs = false
		executionItems        = true
		flowLoglevel          = "Base"
		names                 = false
		variables             = false
		flowId                = "e3aebe90-5a65-409e-9775-43d547b66e07"
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			util.TestAccPreCheck(t)
		},
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create using only required fields i.e. name
				Config: generateFlowLogLevelResource(
					communications,
					eventError,
					eventOther,
					eventWarning,
					executionInputOutputs,
					executionItems,
					flowId,
					flowLoglevel,
					names,
					resourceId,
					variables,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_loglevel."+resourceId, "logLevelCharacteristics", "false"),
				),
			},
		},
	})
}
