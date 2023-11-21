package dependent_consumers

var dependentConsumerMap map[string]string

func SetDependentObjectMaps() map[string]string {
	if len(dependentConsumerMap) < 1 {
		dependentConsumerMap = make(map[string]string)
		dependentConsumerMap["QUEUE"] = "genesyscloud_routing_queue"
		dependentConsumerMap["INBOUNDCALLFLOW"] = "genesyscloud_flow"
		dependentConsumerMap["USER"] = "genesyscloud_user"
		dependentConsumerMap["GROUP"] = "genesyscloud_group"
		dependentConsumerMap["OUTBOUNDCALLFLOW"] = "genesyscloud_flow"
		dependentConsumerMap["KNOWLEDGEBASEDOCUMENT"] = "genesyscloud_knowledge_document"
		dependentConsumerMap["LANGUAGE"] = "genesyscloud_routing_language"
		dependentConsumerMap["ACDSKILL"] = "genesyscloud_routing_skill"
		dependentConsumerMap["ACDWRAPUPCODE"] = "resource_genesyscloud_routing_wrapupcode"
		dependentConsumerMap["CONTACTLIST"] = "resource_genesyscloud_outbound_contactlist"
		dependentConsumerMap["DATATABLE"] = "resource_genesyscloud_architect_datatable"
		dependentConsumerMap["FLOWOUTCOME"] = "resource_genesyscloud_flow_outcome"
		dependentConsumerMap["EMAILROUTE"] = "resource_genesyscloud_routing_email_route"
		dependentConsumerMap["EMERGENCYGROUP"] = "resource_genesyscloud_architect_emergencygroup"
		dependentConsumerMap["GRAMMAR"] = "resource_genesyscloud_architect_grammar"
		dependentConsumerMap["INBOUNDCHATFLOW"] = "genesyscloud_flow"
		dependentConsumerMap["INBOUNDEMAILFLOW"] = "genesyscloud_flow"
		dependentConsumerMap["INBOUNDSHORTMESSAGEFLOW"] = "genesyscloud_flow"
		dependentConsumerMap["INQUEUEEMAILFLOW"] = "genesyscloud_flow"
		dependentConsumerMap["INQUEUESHORTMESSAGEFLOW"] = "genesyscloud_flow"
		dependentConsumerMap["FLOWMILESTONE"] = "resource_genesyscloud_flow_milestone"
		dependentConsumerMap["IVRCONFIGURATION"] = "resource_genesyscloud_architect_ivr"
		dependentConsumerMap["KNOWLEDGEBASE"] = "resource_genesyscloud_knowledge_knowledgebase"
		dependentConsumerMap["OAUTHCLIENT"] = "resource_genesyscloud_oauth_client"
		dependentConsumerMap["RECORDINGPOLICY"] = "resource_genesyscloud_recording_media_retention_policy"
		dependentConsumerMap["RESPONSE"] = "resource_genesyscloud_responsemanagement_response"
		dependentConsumerMap["SCHEDULE"] = "resource_genesyscloud_architect_schedules"
		dependentConsumerMap["SCHEDULEGROUP"] = "resource_genesyscloud_architect_schedulegroups"
		dependentConsumerMap["USERPROMPT"] = "resource_genesyscloud_architect_user_prompt"
		dependentConsumerMap["WIDGET"] = "resource_genesyscloud_widget_deployment"
		dependentConsumerMap["COMPOSERSCRIPT"] = "genesyscloud_script"
		dependentConsumerMap["ACDLANGUAGE"] = "genesyscloud_routing_language"
		dependentConsumerMap["INQUEUECALLFLOW"] = "genesyscloud_flow"
	}
	return dependentConsumerMap
}
