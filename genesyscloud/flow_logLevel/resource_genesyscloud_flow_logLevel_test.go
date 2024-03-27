package flow_logLevel

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

func TestAccResourceFlowOutcome(t *testing.T) {
	var (
		communications        = false
		eventError            = false
		eventOther            = false
		eventWarning          = false
		executionInputOutputs = false
		executionItems        = true
		flowLoglevel          = "Base"
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
					communications,
					eventError,
					eventOther,
					eventWarning,
					executionInputOutputs,
					executionItems,
					flowLoglevel,
					names,
					variables,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_flow_logLevel."+flowLoglevel, "names", "false"),
					resource.TestCheckResourceAttr("genesyscloud_flow_logLevel."+flowLoglevel, "executionItems", "true"),
					provider.TestDefaultHomeDivision("genesyscloud_flow_logLevel."+flowLoglevel),
				),
			},
		},
	})
}

func generateFlowLogLevelResource(
	communications bool,
	eventError bool,
	eventOther bool,
	eventWarning bool,
	executionInputOutputs bool,
	executionItems bool,
	flowLoglevel string,
	names bool,
	variables bool,
) string {
	return fmt.Sprintf(`resource "genesyscloud_flow_logLevel" "flowLogLevel" {
	  level { 
		level 					= "%s" 
      }
	  characteristics {
		communications          = "%s"
		eventError              = "%s"
		eventOther              = "%s"
		eventWarning            = "%s"
		executionItems          = "%s"
		executionInputOutputs   = "%s"
		names                   = "%s"
		variables               = "%s"
	  }
	}
	`, flowLoglevel,
		strconv.FormatBool(executionItems),
		strconv.FormatBool(executionInputOutputs),
		strconv.FormatBool(eventError),
		strconv.FormatBool(eventWarning),
		strconv.FormatBool(eventOther),
		strconv.FormatBool(communications),
		strconv.FormatBool(variables),
		strconv.FormatBool(names))
}
