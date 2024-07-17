package outbound_contact_list_template

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const (
	resourceName = "genesyscloud_outbound_contact_list_template"
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource(resourceName, DataSourceOutboundContactListTemplate())
	regInstance.RegisterResource(resourceName, ResourceOutboundContactListTemplate())
	regInstance.RegisterExporter(resourceName, OutboundContactListTemplateExporter())
}
