package routing_email_route

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

/*
The resource_genesyscloud_routing_email_route.go contains all of the methods that perform the core logic for a resource.
*/

// getAllAuthRoutingEmailRoute retrieves all of the routing email route via Terraform in the Genesys Cloud and is used for the exporter
func getAllRoutingEmailRoutes(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := newRoutingEmailRouteProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	inboundRoutes, err := proxy.getAllRoutingEmailRoute(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get routing email route: %v", err)
	}

	for _, inboundRoute := range *inboundRoutes {
		resources[*inboundRoute.Id] = &resourceExporter.ResourceMeta{Name: *inboundRoute.Name}
	}

	return resources, nil
}

// createRoutingEmailRoute is used by the routing_email_route resource to create Genesys cloud routing email route
func createRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteProxy(sdkConfig)

	routingEmailRoute := getRoutingEmailRouteFromResourceData(d)

	log.Printf("Creating routing email route %s", *routingEmailRoute.Name)
	inboundRoute, err := proxy.createRoutingEmailRoute(ctx, &routingEmailRoute)
	if err != nil {
		return diag.Errorf("Failed to create routing email route: %s", err)
	}

	d.SetId(*inboundRoute.Id)
	log.Printf("Created routing email route %s", *inboundRoute.Id)
	return readRoutingEmailRoute(ctx, d, meta)
}

// readRoutingEmailRoute is used by the routing_email_route resource to read an routing email route from genesys cloud
func readRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteProxy(sdkConfig)

	log.Printf("Reading routing email route %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		inboundRoute, respCode, getErr := proxy.getRoutingEmailRouteById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404ByInt(respCode) {
				return retry.RetryableError(fmt.Errorf("Failed to read routing email route %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read routing email route %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingEmailRoute())

		resourcedata.SetNillableValue(d, "name", inboundRoute.Name)
		resourcedata.SetNillableValue(d, "pattern", inboundRoute.Pattern)
		resourcedata.SetNillableReference(d, "queue_id", inboundRoute.QueueId)
		resourcedata.SetNillableValue(d, "priority", inboundRoute.Priority)
		// TODO: Handle skills property
		resourcedata.SetNillableReference(d, "language_id", inboundRoute.LanguageId)
		resourcedata.SetNillableValue(d, "from_name", inboundRoute.FromName)
		resourcedata.SetNillableValue(d, "from_email", inboundRoute.FromEmail)
		resourcedata.SetNillableReference(d, "flow_id", inboundRoute.FlowId)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "reply_email_address", inboundRoute.ReplyEmailAddress, flattenQueueEmailAddress)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "auto_bcc", inboundRoute.AutoBcc, flattenEmailAddresss)
		resourcedata.SetNillableReference(d, "spam_flow_id", inboundRoute.SpamFlowId)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "signature", inboundRoute.Signature, flattenSignature)
		resourcedata.SetNillableValue(d, "history_inclusion", inboundRoute.HistoryInclusion)
		resourcedata.SetNillableValue(d, "allow_multiple_actions", inboundRoute.AllowMultipleActions)

		log.Printf("Read routing email route %s %s", d.Id(), *inboundRoute.Name)
		return cc.CheckState()
	})
}

// updateRoutingEmailRoute is used by the routing_email_route resource to update an routing email route in Genesys Cloud
func updateRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteProxy(sdkConfig)

	routingEmailRoute := getRoutingEmailRouteFromResourceData(d)
	domainId := d.Get("domain_id").(string)

	//Checking the self_reference_route flag and routeId rules
	if err := validateSdkReplyEmailAddress(d); err != nil {
		return diag.Errorf("Error occurred while validating the reply email address while trying to update the record: %s", err)
	}
	// TODO: Check if Small functions are needed

	log.Printf("Updating routing email route %s", d.Id())
	inboundRoute, resp, err := proxy.updateRoutingEmailRoute(ctx, d.Id(), domainId, &routingEmailRoute)
	if err != nil {
		return diag.Errorf("Failed to update routing email route: %s %s", resp, err)
	}

	log.Printf("Updated routing email route %s", *inboundRoute.Id)
	return readRoutingEmailRoute(ctx, d, meta)
}

// deleteRoutingEmailRoute is used by the routing_email_route resource to delete an routing email route from Genesys cloud
func deleteRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteProxy(sdkConfig)
	domainId := d.Get("domain_id").(string)

	resp, err := proxy.deleteRoutingEmailRoute(ctx, domainId, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete routing email route %s: %s %s", d.Id(), resp, err)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, respCode, err := proxy.getRoutingEmailRouteById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404ByInt(respCode) {
				// Routing email domain route deleted
				log.Printf("Deleted routing email route %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting routing email route %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("routing email route %s still exists", d.Id()))
	})
}
