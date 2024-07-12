package outbound_contact_list

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const resourceName = "genesyscloud_outbound_contact_list"

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource(resourceName, DataSourceOutboundContactList())
	regInstance.RegisterResource(resourceName, ResourceOutboundContactList())
	regInstance.RegisterExporter(resourceName, OutboundContactListExporter())
}
