package flow_loglevel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

// getFlowLogLevelFromResourceData maps data from schema ResourceData object to a platformclientv2.Flowloglevel
func getFlowLogLevelFromResourceData(d *schema.ResourceData) *platformclientv2.Flowloglevel {
	logLevel := platformclientv2.Flowloglevel{
		Level: platformclientv2.String(d.Get("flow_log_level").(string)),
	}
	if len(d.Get("flow_characteristics").([]interface{})) > 0 {
		logLevel.Characteristics = getFlowLogLevelCharacteristicsFromResourceData(d)
	}
	return &logLevel

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
