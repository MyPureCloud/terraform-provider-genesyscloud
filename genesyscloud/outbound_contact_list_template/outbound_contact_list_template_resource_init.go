package outbound_contact_list_template

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const (
	resourceName = "genesyscloud_outbound_contact_list_template"
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource("genesyscloud_outbound_contact_list_template", DataSourceOutboundContactListTemplate())
	regInstance.RegisterResource("genesyscloud_outbound_contact_list_template", ResourceOutboundContactListTemplate())
	regInstance.RegisterExporter("genesyscloud_outbound_contact_list_template", OutboundContactListTemplateExporter())
}
