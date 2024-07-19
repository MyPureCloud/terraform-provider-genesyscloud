package routing_queue_outbound_email_address

import (
	"fmt"
	"log"
	"os"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailDomain "terraform-provider-genesyscloud/genesyscloud/routing_email_domain"
	routingEmailRoute "terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"
	"terraform-provider-genesyscloud/genesyscloud/util"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
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

	// Use this to save the id of the parent queue
	queueIdChan := make(chan string, 1)

	err := os.Setenv(featureToggles.OEAToggleName(), "enabled")
	if err != nil {
		t.Errorf("%s is not set", featureToggles.OEAToggleName())
	}

	cleanupRoutingEmailDomains()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, nil),
		Steps: []resource.TestStep{
			{
				// Create the queue first so we can save the id to a channel and use it in the later test steps
				// The reason we are doing this is that we need to verify the parent queue is never dropped and recreated because of OEA
				Config: routingQueue.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName1,
				),
				Check: resource.ComposeTestCheckFunc(
					func(state *terraform.State) error {
						resourceState, ok := state.RootModule().Resources["genesyscloud_routing_queue."+queueResource]
						if !ok {
							return fmt.Errorf("failed to find resource %s in state", "genesyscloud_routing_queue."+queueResource)
						}
						queueIdChan <- resourceState.Primary.ID

						return nil
					},
				),
			},
			{
				Config: routingQueue.GenerateRoutingQueueResourceBasic(
					queueResource,
					queueName1,
				) + routingEmailDomain.GenerateRoutingEmailDomainResource(
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
					resource.TestCheckResourceAttrWith("genesyscloud_routing_queue."+queueResource, "id", checkQueueId(queueIdChan)),
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

func checkQueueId(queueIdChan chan string) func(value string) error {
	return func(value string) error {
		queueId, ok := <-queueIdChan
		if !ok {
			return fmt.Errorf("queue id channel closed unexpectedly")
		}

		if value != queueId {
			return fmt.Errorf("queue id not equal to expected. Expected: %s, Actual: %s", queueId, value)
		}

		close(queueIdChan)
		return nil
	}
}

func cleanupRoutingEmailDomains() {
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
