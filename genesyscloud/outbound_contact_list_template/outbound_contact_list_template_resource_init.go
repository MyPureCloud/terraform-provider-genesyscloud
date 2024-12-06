package outbound_contact_list_template

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const (
	ResourceType = "genesyscloud_outbound_contact_list_template"
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource(ResourceType, DataSourceOutboundContactListTemplate())
	regInstance.RegisterResource(ResourceType, ResourceOutboundContactListTemplate())
	regInstance.RegisterExporter(ResourceType, OutboundContactListTemplateExporter())
}
