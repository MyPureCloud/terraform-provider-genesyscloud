package process_automation_trigger

import (
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource("genesyscloud_processautomation_trigger", ResourceProcessAutomationTrigger())
	regInstance.RegisterDataSource("genesyscloud_processautomation_trigger", dataSourceProcessAutomationTrigger())
	regInstance.RegisterExporter("genesyscloud_processautomation_trigger", ProcessAutomationTriggerExporter())
}
