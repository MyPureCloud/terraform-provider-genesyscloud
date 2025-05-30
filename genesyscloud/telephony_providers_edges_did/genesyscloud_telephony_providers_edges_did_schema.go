package telephony_providers_edges_did

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ResourceType = "genesyscloud_telephony_providers_edges_did"

// SetRegistrar registers all resources, data sources and exporters in the package
func SetRegistrar(l registrar.Registrar) {
	l.RegisterDataSource(ResourceType, DataSourceDid())
}

// DataSourceDid registers the genesyscloud_telephony_providers_edges_did data source
func DataSourceDid() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud DID. The identifier is the E.164 phone number.",
		ReadContext: provider.ReadWithPooledClient(dataSourceDidRead),
		Schema: map[string]*schema.Schema{
			"phone_number": {
				Description:      "Phone number for the DID.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidatePhoneNumber,
			},
		},
	}
}
