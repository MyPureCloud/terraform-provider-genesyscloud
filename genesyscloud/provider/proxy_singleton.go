package provider

import (
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/mrmo"
)

// SingletonOrFresh returns a new proxy when MRMO standalone mode is active
// (MRMO_CXASCODE_INTEGRATION_ENABLED set via mrmo.Activate). Otherwise it
// preserves the legacy process singleton used by Terraform and unit tests.
func SingletonOrFresh[T any](
	slot **T,
	clientConfig *platformclientv2.Configuration,
	factory func(*platformclientv2.Configuration) *T,
) *T {
	if mrmo.IsActive() {
		return factory(clientConfig)
	}
	if *slot == nil {
		*slot = factory(clientConfig)
	}
	return *slot
}
