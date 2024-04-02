package flow_logLevel

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

func generateFlowLogLevelResource(
	flowId string,
	flowLoglevel string,
	resourceId string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_flow_loglevel" "%s" {
	  flow_id					= "%s"
	  flow_log_level 			= "%s"
	}
	`,
		resourceId,
		flowId,
		flowLoglevel)
}

// getFlowLogLevelSettingsRequestFromResourceData maps data from schema ResourceData object to a platformclientv2.Flowloglevelrequest
func getFlowLogLevelSettingsRequestFromResourceData(d *schema.ResourceData) platformclientv2.Flowloglevelrequest {
	return platformclientv2.Flowloglevelrequest{
		LogLevelCharacteristics: getFlowLogLevelFromResourceData(d),
	}
}

// getFlowLogLevelRequestFromFlowLogLevel maps data from schema ResourceData object to a platformclientv2.Flowloglevelrequest
func getFlowLogLevelRequestFromFlowLogLevel(d *platformclientv2.Flowloglevel) platformclientv2.Flowloglevelrequest {

	return platformclientv2.Flowloglevelrequest{
		LogLevelCharacteristics: d,
	}
}

// getFlowLogLevelFromResourceData maps data from schema ResourceData object to a platformclientv2.Flowloglevel
func getFlowLogLevelFromResourceData(d *schema.ResourceData) *platformclientv2.Flowloglevel {
	level := d.Get("flow_log_level").(string)
	if len(d.Get("flow_characteristics").([]interface{})) > 0 {
		return &platformclientv2.Flowloglevel{
			Level:           &level,
			Characteristics: getFlowLogLevelCharacteristicsFromResourceData(d),
		}
	} else {
		return &platformclientv2.Flowloglevel{
			Level: &level,
		}
	}

}

// getFlowLogLevelCharacteristicsFromResourceData maps data from schema ResourceData object to a platformclientv2.Flowcharacteristics
func getFlowLogLevelCharacteristicsFromResourceData(d *schema.ResourceData) *platformclientv2.Flowcharacteristics {
	characteristics := d.Get("flow_characteristics").([]interface{})[0].(map[string]interface{})
	communications := characteristics["communications"].(bool)
	eventError := characteristics["event_error"].(bool)
	eventOther := characteristics["event_other"].(bool)
	eventWarning := characteristics["event_warning"].(bool)
	executionInputOutputs := characteristics["execution_input_outputs"].(bool)
	executionItems := characteristics["execution_items"].(bool)
	names := characteristics["names"].(bool)
	variables := characteristics["variables"].(bool)

	return &platformclientv2.Flowcharacteristics{
		ExecutionItems:        &executionItems,
		ExecutionInputOutputs: &executionInputOutputs,
		Communications:        &communications,
		EventError:            &eventError,
		EventWarning:          &eventWarning,
		EventOther:            &eventOther,
		Variables:             &variables,
		Names:                 &names,
	}
}

// flattenPhoneNumber converts a platformclientv2.Phonenumber into a map and then into array for consumption by Terraform
func flattenFlowCharacteristics(characteristics *platformclientv2.Flowcharacteristics) []interface{} {
	characteristicsInterface := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(characteristicsInterface, "communications", characteristics.Communications)
	resourcedata.SetMapValueIfNotNil(characteristicsInterface, "event_error", characteristics.EventError)
	resourcedata.SetMapValueIfNotNil(characteristicsInterface, "event_other", characteristics.EventOther)
	resourcedata.SetMapValueIfNotNil(characteristicsInterface, "event_warning", characteristics.EventWarning)
	resourcedata.SetMapValueIfNotNil(characteristicsInterface, "execution_input_outputs", characteristics.ExecutionInputOutputs)
	resourcedata.SetMapValueIfNotNil(characteristicsInterface, "execution_items", characteristics.ExecutionItems)
	resourcedata.SetMapValueIfNotNil(characteristicsInterface, "names", characteristics.Names)
	resourcedata.SetMapValueIfNotNil(characteristicsInterface, "variables", characteristics.Variables)
	return []interface{}{characteristicsInterface}
}
