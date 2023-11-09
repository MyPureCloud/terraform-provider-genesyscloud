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
	}
	return dependentConsumerMap
}
