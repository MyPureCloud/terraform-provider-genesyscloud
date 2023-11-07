package tfexporter

var dependentConsumerMap map[string]string

func SetDependentObjectMaps() map[string]string {
	if len(dependentConsumerMap) < 1 {
		dependentConsumerMap = make(map[string]string)
		dependentConsumerMap["genesyscloud_routing_queue"] = "QUEUE"
	}
	return dependentConsumerMap
}
