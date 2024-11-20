package routing_email_route

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingEmailRoute(t *testing.T) {
	var (
		domainResourceLabel = "routing-domain1"
		domainId            = fmt.Sprintf("terraformroutes.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))
		routeResourceLabel  = "email-route1"
		routePattern1       = "terraform1"
		fromEmail1          = "terraform1@test.com"
		fromName1           = "John Terraform"
		bccEmail1           = "test1@" + domainId
	)

	// Standard acceptance tests
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create email domain and basic route
				Config: routingEmailDomain.GenerateRoutingEmailDomainResource(
					domainResourceLabel,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + GenerateRoutingEmailRouteResource(
					routeResourceLabel,
					"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
					routePattern1,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingAutoBcc(fromName1, bccEmail1),
				) + generateRoutingEmailRouteDataSource(
					routeResourceLabel,
					routePattern1,
					"genesyscloud_routing_email_domain."+domainResourceLabel+".id",
					"genesyscloud_routing_email_route."+routeResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_routing_email_route."+routeResourceLabel, "id",
						"genesyscloud_routing_email_route."+routeResourceLabel, "id",
					),
				),
			},
		},
		CheckDestroy: testVerifyRoutingEmailRouteDestroyed,
	})
}

func generateRoutingEmailRouteDataSource(
	resourceLabel string,
	pattern string,
	domainId string,
	dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_routing_email_route" "%s" {
			pattern = "%s"
			domain_id = "%s"
			depends_on=[%s]
		}
	`, resourceLabel, pattern, domainId, dependsOn)
}
