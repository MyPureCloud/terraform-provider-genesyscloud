package flow_logLevel

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

func TestAccResourceFlowOutcome(t *testing.T) {
	var (
		outcomeResource1      = "flow-outcome1"
		communications        = false
		eventError            = false
		eventOther            = false
		eventWarning          = false
		executionInputOutputs = false
		executionItems        = true
		names                 = false
		variables             = false
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create using only required fields i.e. name
				Config: generateFlowLogLevelResource(
					outcomeResource1,
					executionItems,
					executionInputOutputs,
					eventError,
					eventWarning,
					eventOther,
					communications,
					variables,
					names,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_outcome."+outcomeResource1, "names", "false"),
					resource.TestCheckResourceAttr("genesyscloud_flow_outcome."+outcomeResource1, "execution_items", "true"),
					provider.TestDefaultHomeDivision("genesyscloud_flow_outcome."+outcomeResource1),
				),
			},
		},
	})
}

func generateFlowLogLevelResource(
	flowLoglevel string,
	executionItems bool,
	executionInputOutputs bool,
	eventError bool,
	eventWarning bool,
	eventOther bool,
	communications bool,
	variables bool,
	names bool) string {
	return fmt.Sprintf(`resource "genesyscloud_flow_loglevel" "flowLogLevel" {
	  flow_log_level = "%s"
	  flow_characteristics {
		execution_items         = "%s"
		execution_input_outputs = "%s"
		communications          = "%s"
		event_error             = "%s"
		event_warning           = "%s"
		event_other             = "%s"
		variables               = "%s"
		names                   = "%s"
	  }
	}
	`, flowLoglevel,
		executionItems,
		executionInputOutputs,
		eventError,
		eventWarning,
		eventOther,
		communications,
		variables,
		names)
}
