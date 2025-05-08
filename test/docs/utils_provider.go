package testing

import (
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Create a simple utils provider that implements just what you need
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

func UtilsProviderFactory() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"utils": func() (*schema.Provider, error) { return utilsProvider(), nil },
	}
}
