package routing_queue_outbound_email_address

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailRoute "terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
)

func TestAccResourceRoutingQueueOutboundEmailAddress(t *testing.T) {
	var (
		outboundEmailAddressResource = "test-email-address"

		queueResource = "test-queue"
		queueName1    = "Terraform Test Queue-" + uuid.NewString()

		domainResource = "test-domain"
		domainId       = fmt.Sprintf("terraform.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))

		routeResource = "email-route"
		routePattern  = "terraform1"
		fromName      = "John Terraform"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: gcloud.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName1,
				) + gcloud.GenerateRoutingEmailDomainResource(
					domainResource,
					domainId,
					util.FalseValue,
					util.NullValue,
				) + routingEmailRoute.GenerateRoutingEmailRouteResource(
					routeResource,
					"genesyscloud_routing_email_domain."+domainResource+".id",
					routePattern,
					fromName,
				) + generateRoutingQueueOutboundEmailAddressResource(
					outboundEmailAddressResource,
					"genesyscloud_routing_queue."+queueResource+".id",
					"genesyscloud_routing_email_domain."+domainResource+".id",
					"genesyscloud_routing_email_route."+routeResource+".id",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+outboundEmailAddressResource, "queue_id", "genesyscloud_routing_queue."+queueResource, "id",
					),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+outboundEmailAddressResource, "domain_id", "genesyscloud_routing_email_domain."+domainResource, "id",
					),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_conditional_group_routing."+outboundEmailAddressResource, "route_id", "genesyscloud_routing_email_route."+routeResource, "id",
					),
				),
			},
			{
				// Import/Read
				ResourceName:      resourceName + outboundEmailAddressResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateRoutingQueueOutboundEmailAddressResource(resourceId, queueId, domainId, routeId string) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		queue_id = %s
		domain_id = %s
		route_id = %s
	}`, resourceName, resourceId, queueId, domainId, routeId)
}
