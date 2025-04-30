package testing

import (
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Create a simple random provider that implements just what you need
func randomProvider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"random_uuid": {
				Schema: map[string]*schema.Schema{
					"result": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
				Create: func(d *schema.ResourceData, m interface{}) error {
					uuid := uuid.New().String()
					d.SetId(uuid)
					d.Set("result", uuid)
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

func RandomProviderFactory() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"random": func() (*schema.Provider, error) { return randomProvider(), nil },
	}
}
