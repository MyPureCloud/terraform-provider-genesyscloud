package outbound_contact_list

import (
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

const ResourceType = "genesyscloud_outbound_contact_list"
const S3Enabled = false

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource(ResourceType, DataSourceOutboundContactList())
	regInstance.RegisterResource(ResourceType, ResourceOutboundContactList())
	regInstance.RegisterExporter(ResourceType, OutboundContactListExporter())
}
