package routing_queue_outbound_email_address

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
	"log"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailRoute "terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"
	"time"
)

func TestAccResourceRoutingQueueOutboundEmailAddress(t *testing.T) {
	var (
		outboundEmailAddressResource = "test-email-address"

		queueResource = "test-queue"
		queueName1    = "Terraform Test Queue-" + uuid.NewString()

		domainResource = "test-domain"
		domainId       = fmt.Sprintf("terraform.%s.com", strings.Replace(uuid.NewString(), "-", "", -1))

		routeResource = "test-route"
		routePattern  = "terraform1"
		fromName      = "John Terraform"
	)

	CleanupRoutingEmailDomains()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				Config: routingQueue.GenerateRoutingQueueResourceBasic(
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
						"genesyscloud_routing_queue_outbound_email_address."+outboundEmailAddressResource, "queue_id", "genesyscloud_routing_queue."+queueResource, "id",
					),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_outbound_email_address."+outboundEmailAddressResource, "domain_id", "genesyscloud_routing_email_domain."+domainResource, "id",
					),
					resource.TestCheckResourceAttrPair(
						"genesyscloud_routing_queue_outbound_email_address."+outboundEmailAddressResource, "route_id", "genesyscloud_routing_email_route."+routeResource, "id",
					),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_routing_queue_outbound_email_address." + outboundEmailAddressResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func generateRoutingQueueOutboundEmailAddressResource(resourceId, queueId, domainId, routeId string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue_outbound_email_address" "%s" {
		queue_id = %s
		domain_id = %s
		route_id = %s
	}`, resourceId, queueId, domainId, routeId)
}

func CleanupRoutingEmailDomains() {
	var sdkConfig *platformclientv2.Configuration
	var err error
	if sdkConfig, err = provider.AuthorizeSdk(); err != nil {
		log.Fatal(err)
	}

	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		routingEmailDomains, _, getErr := routingAPI.GetRoutingEmailDomains(pageSize, pageNum, false, "")
		if getErr != nil {
			log.Printf("failed to get page %v of routing email domains: %v", pageNum, getErr)
			return
		}

		if routingEmailDomains.Entities == nil || len(*routingEmailDomains.Entities) == 0 {
			return
		}

		for _, routingEmailDomain := range *routingEmailDomains.Entities {
			if routingEmailDomain.Id != nil && strings.HasPrefix(*routingEmailDomain.Id, "terraform") {
				_, err := routingAPI.DeleteRoutingEmailDomain(*routingEmailDomain.Id)
				if err != nil {
					log.Printf("Failed to delete routing email domain %s: %s", *routingEmailDomain.Id, err)
					continue
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}
