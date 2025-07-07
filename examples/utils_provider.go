// Package examples provides utilities for testing Terraform examples.
package examples

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// utilsProvider creates a simple Terraform provider that implements utility resources
// needed for testing examples, such as certificate generation.
func utilsProvider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"utils_certificates": {
				Schema: map[string]*schema.Schema{
					"cert1": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"cert2": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
				Create: func(d *schema.ResourceData, m interface{}) error {
					uuid := uuid.New().String()
					d.SetId(uuid)
					d.Set("cert1", util.TestCert1)
					d.Set("cert2", util.TestCert2)
					return nil
				},
				Read: func(d *schema.ResourceData, m interface{}) error {
					return nil
				},
				Delete: func(d *schema.ResourceData, m interface{}) error {
					d.SetId("")
					return nil
				},
			},
		},
	}
}

// ExampleUtilsProviderFactory returns a provider factory map for the utils provider.
// This can be combined with other provider factories to make utility resources available
// during testing of examples.
func ExampleUtilsProviderFactory() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"utils": func() (*schema.Provider, error) { return utilsProvider(), nil },
	}
}
