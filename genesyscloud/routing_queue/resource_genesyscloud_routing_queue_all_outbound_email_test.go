package routing_queue

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	routingEmailRoute "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_email_route"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
)

// existingEmailDomain is a pre-existing verified email domain on the DCA test org. The test
// references it via a data source (rather than creating a new domain) because the org is at its
// email-domain cap; routes, however, can be freely created on an existing domain.
const existingEmailDomain = "bughuntdca.inindca.com"

// TestAccResourceRoutingQueueAllOutboundEmailAddresses verifies the all_outbound_email_addresses
// block on genesyscloud_routing_queue, which supports MULTIPLE outbound email domains/routes
// (the fix for AS-5605 — the deprecated outbound_email_address block only allowed one).
//
// It also directly addresses the PR #2568 review concern that a TypeList is order-sensitive and
// could produce a permanent plan diff if the API returns the addresses in a different order than
// declared. The test creates a queue with TWO addresses, then re-applies the SAME config and an
// order-swapped config; both steps rely on the built-in post-apply plan check that fails on any
// non-empty plan, so a spurious reorder diff would fail the test.
func TestAccResourceRoutingQueueAllOutboundEmailAddresses(t *testing.T) {
	var (
		queueResourceLabel = "test-queue-aoea"
		queueName          = "Terraform Test Queue AOEA-" + uuid.NewString()

		domainDataLabel = "aoea-domain"

		routeResourceLabel1 = "aoea-route1"
		routeResourceLabel2 = "aoea-route2"
		routePattern1       = "tfaoea1" + strings.Replace(uuid.NewString(), "-", "", -1)[:8]
		routePattern2       = "tfaoea2" + strings.Replace(uuid.NewString(), "-", "", -1)[:8]
		fromName1           = "AOEA One"
		fromName2           = "AOEA Two"

		domainRef = "data.genesyscloud_routing_email_domain." + domainDataLabel + ".id"
		route1Ref = "genesyscloud_routing_email_route." + routeResourceLabel1 + ".id"
		route2Ref = "genesyscloud_routing_email_route." + routeResourceLabel2 + ".id"

		queuePath = "genesyscloud_routing_queue." + queueResourceLabel
	)

	// Dependency HCL: reference the EXISTING domain via a data source (no domain creation), then
	// create two routes on it. depends_on is not needed for a data source over a static domain.
	emailDeps := generateExistingEmailDomainDataSource(domainDataLabel, existingEmailDomain) +
		routingEmailRoute.GenerateRoutingEmailRouteResource(
			routeResourceLabel1,
			domainRef,
			routePattern1,
			fromName1,
		) + routingEmailRoute.GenerateRoutingEmailRouteResource(
		routeResourceLabel2,
		domainRef,
		routePattern2,
		fromName2,
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create the queue with TWO outbound email addresses (order: route1, route2).
				// route1 is the default (outbound_email_address) and is a member of the list.
				Config: emailDeps + generateRoutingQueueWithAllOutboundEmail(
					queueResourceLabel, queueName,
					domainRef, route1Ref,
					generateAllOutboundEmailAddress(domainRef, route1Ref),
					generateAllOutboundEmailAddress(domainRef, route2Ref),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(queuePath, "all_outbound_email_addresses.#", "2"),
					resource.TestCheckResourceAttrPair(queuePath, "all_outbound_email_addresses.0.route_id", "genesyscloud_routing_email_route."+routeResourceLabel1, "id"),
					resource.TestCheckResourceAttrPair(queuePath, "all_outbound_email_addresses.1.route_id", "genesyscloud_routing_email_route."+routeResourceLabel2, "id"),
				),
			},
			{
				// Re-apply the SAME config. The framework runs a plan after this step and fails on
				// any non-empty plan — this is the core check that the API round-trip does not
				// produce a perpetual diff for two addresses (PR #2568 reviewer concern).
				Config: emailDeps + generateRoutingQueueWithAllOutboundEmail(
					queueResourceLabel, queueName,
					domainRef, route1Ref,
					generateAllOutboundEmailAddress(domainRef, route1Ref),
					generateAllOutboundEmailAddress(domainRef, route2Ref),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(queuePath, "all_outbound_email_addresses.#", "2"),
				),
			},
			{
				// SWAP the declared order (route2 first, route1 second). With a TypeList this is the
				// scenario that would surface an order-sensitivity problem. Kept as its own step so
				// the behavior is explicit and observable if the reviewer's concern materializes.
				PreConfig: func() { time.Sleep(5 * time.Second) },
				Config: emailDeps + generateRoutingQueueWithAllOutboundEmail(
					queueResourceLabel, queueName,
					domainRef, route1Ref,
					generateAllOutboundEmailAddress(domainRef, route2Ref),
					generateAllOutboundEmailAddress(domainRef, route1Ref),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(queuePath, "all_outbound_email_addresses.#", "2"),
				),
			},
			{
				// Import / read round-trip.
				ResourceName:      queuePath,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// generateExistingEmailDomainDataSource references a pre-existing email domain by name.
func generateExistingEmailDomainDataSource(dataLabel, domainName string) string {
	return fmt.Sprintf(`data "genesyscloud_routing_email_domain" "%s" {
	name = "%s"
}
`, dataLabel, domainName)
}

// generateAllOutboundEmailAddress builds a single all_outbound_email_addresses block.
// domainRef and routeRef are HCL references (e.g. genesyscloud_routing_email_route.x.id).
func generateAllOutboundEmailAddress(domainRef, routeRef string) string {
	return fmt.Sprintf(`
	all_outbound_email_addresses {
		domain_id = %s
		route_id  = %s
	}
`, domainRef, routeRef)
}

// generateRoutingQueueWithAllOutboundEmail builds a queue with the supplied
// all_outbound_email_addresses blocks plus a default outbound_email_address.
//
// The API contract (verified against AS-5605) requires that:
//   - email media settings exist with a valid alerting_timeout (>= 7) once outbound email is set, and
//   - the default outbound_email_address is a MEMBER of all_outbound_email_addresses
//     ("Default outbound email address missing from outbound email address list" otherwise).
//
// So the caller must pass a defaultDomainRef/defaultRouteRef that also appears in aoeaBlocks.
func generateRoutingQueueWithAllOutboundEmail(resourceLabel, name, defaultDomainRef, defaultRouteRef string, aoeaBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_queue" "%s" {
	name = "%s"
	media_settings_email {
		alerting_timeout_sec      = 30
		service_level_percentage  = 0.8
		service_level_duration_ms = 20000
	}
	outbound_email_address {
		domain_id = %s
		route_id  = %s
	}
	%s
}
`, resourceLabel, name, defaultDomainRef, defaultRouteRef, strings.Join(aoeaBlocks, "\n"))
}
