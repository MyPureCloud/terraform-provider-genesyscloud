package outbound_contact_list

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource("genesyscloud_outbound_contact_list", DataSourceOutboundContactList())
	regInstance.RegisterResource("genesyscloud_outbound_contact_list", ResourceOutboundContactList())
	regInstance.RegisterExporter("genesyscloud_outbound_contact_list", OutboundContactListExporter())
}
