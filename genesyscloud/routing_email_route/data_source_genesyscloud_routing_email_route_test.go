package routing_email_route

import (
	"fmt"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRoutingEmailRoute(t *testing.T) {
	var (
		name          = "Response-" + uuid.NewString()
		domainRes     = "routing-domain1"
		domainId      = fmt.Sprintf("terraformroutes.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))
		routeRes      = "email-route1"
		routePattern1 = "terraform1"
		fromEmail1    = "terraform1@test.com"
		fromName1     = "John Terraform"
		bccEmail1     = "test1@" + domainId
	)

	// Standard acceptance tests
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create email domain and basic route
				Config: gcloud.GenerateRoutingEmailDomainResource(
					domainRes,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + GenerateRoutingEmailRouteResource(
					routeRes,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					routePattern1,
					fromName1,
					fmt.Sprintf("from_email = \"%s\"", fromEmail1),
					generateRoutingAutoBcc(fromName1, bccEmail1),
				) + generateRoutingEmailRouteDataSource(
					routeRes,
					name,
					"genesyscloud_routing_email_domain."+domainRes+".id",
					"genesyscloud_routing_email_route."+routeRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_routing_email_route."+domainRes, "id",
						"genesyscloud_routing_email_route."+routeRes, "id",
					),
				),
			},
		},
		CheckDestroy: testVerifyRoutingEmailRouteDestroyed,
	})
}

func generateRoutingEmailRouteDataSource(
	resourceID string,
	name string,
	domainId string,
	dependsOn string) string {
	return fmt.Sprintf(`
		data "genesyscloud_routing_email_route" "%s" {
			name = "%s"
			domain_id = "%s"
			depends_on=[%s]
		}
	`, resourceID, name, domainId, dependsOn)
}
