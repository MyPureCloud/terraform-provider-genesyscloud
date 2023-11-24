package authorization_product

import (
	registrar "terraform-provider-genesyscloud/genesyscloud/resource_register"
)

func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterDataSource("genesyscloud_authorization_product", DataSourceAuthorizationProduct())
}
