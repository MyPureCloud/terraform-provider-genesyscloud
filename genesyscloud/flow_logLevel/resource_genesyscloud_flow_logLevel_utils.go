package flow_logLevel

import (
	"fmt"
	"strconv"
)

func generateFlowLogLevelResource(
	communications bool,
	eventError bool,
	eventOther bool,
	eventWarning bool,
	executionInputOutputs bool,
	executionItems bool,
	flowId string,
	flowLoglevel string,
	names bool,
	variables bool,
) string {
	return fmt.Sprintf(`resource "genesyscloud_flow_logLevel" "flowLogLevel" {
	  flow_id					= "%s"
	  flow_log_level 					= "%s"
	  flow_characteristics {
		communications          = "%s"
		event_error              = "%s"
		event_other              = "%s"
		event_warning            = "%s"
		execution_input_outputs   = "%s"
		execution_items          = "%s"
		names                   = "%s"
		variables               = "%s"
	  }
	}
	`,
		flowId,
		flowLoglevel,
		strconv.FormatBool(communications),
		strconv.FormatBool(eventError),
		strconv.FormatBool(eventOther),
		strconv.FormatBool(eventWarning),
		strconv.FormatBool(executionInputOutputs),
		strconv.FormatBool(executionItems),
		strconv.FormatBool(names),
		strconv.FormatBool(variables))
}
