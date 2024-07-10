package architect_emergencygroup

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func buildSdkEmergencyGroupCallFlows(d *schema.ResourceData) *[]platformclientv2.Emergencycallflow {
	var allCallFlows []platformclientv2.Emergencycallflow
	if callFlows, ok := d.GetOk("emergency_call_flows"); ok {
		for _, callFlow := range callFlows.([]interface{}) {
			callFlowSettings := callFlow.(map[string]interface{})
			var currentCallFlow platformclientv2.Emergencycallflow

			if flowID, ok := callFlowSettings["emergency_flow_id"].(string); ok {
				currentCallFlow.EmergencyFlow = &platformclientv2.Domainentityref{Id: &flowID}
			}

			if ivrIds, ok := callFlowSettings["ivr_ids"]; ok {
				ids := ivrIds.(*schema.Set).List()
				if len(ids) > 0 {
					sdkIvrIds := make([]platformclientv2.Domainentityref, len(ids))
					for i, id := range ids {
						ivrID := id.(string)
						sdkIvrIds[i] = platformclientv2.Domainentityref{Id: &ivrID}
					}
					currentCallFlow.Ivrs = &sdkIvrIds
				}
			}
			allCallFlows = append(allCallFlows, currentCallFlow)
		}
	}
	return &allCallFlows
}

func flattenEmergencyCallFlows(emergencyCallFlows []platformclientv2.Emergencycallflow) []interface{} {
	callFlows := make([]interface{}, len(emergencyCallFlows))
	for i, callFlow := range emergencyCallFlows {
		callFlowSettings := make(map[string]interface{})
		if callFlow.EmergencyFlow != nil {
			callFlowSettings["emergency_flow_id"] = *callFlow.EmergencyFlow.Id
		}
		if callFlow.Ivrs != nil && len(*callFlow.Ivrs) > 0 {
			ivrIds := make([]interface{}, len(*callFlow.Ivrs))
			for k, id := range *callFlow.Ivrs {
				ivrIds[k] = *id.Id
			}
			callFlowSettings["ivr_ids"] = schema.NewSet(schema.HashString, ivrIds)
		}
		callFlows[i] = callFlowSettings
	}
	return callFlows
}
