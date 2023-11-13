package genesyscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

// Returns the schema for the routing email domain
func DataSourceRoutingEmailDomain() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Email Domains. Select an email domain by name",
		ReadContext: ReadWithPooledClient(DataSourceRoutingEmailDomainRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Email domain name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

// Looks up the data for the Email Domain
func DataSourceRoutingEmailDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100

			domains, _, getErr := routingAPI.GetRoutingEmailDomains(pageSize, pageNum, false, "")

			if getErr != nil {
				return retry.NonRetryableError(fmt.Errorf("Error requesting email domain %s: %s", name, getErr))
			}

			//// No record found, keep trying for X seconds as this might an eventual consistency problem
			if domains.Entities == nil || len(*domains.Entities) == 0 {
				return retry.RetryableError(fmt.Errorf("No email domains found with name %s", name))
			}

			// Once I get a result, cycle through until we find a name that matches
			for _, domain := range *domains.Entities {
				if domain.Id != nil && *domain.Id == name {

					d.SetId(*domain.Id)
					return nil
				}
			}
		}
	})
}
