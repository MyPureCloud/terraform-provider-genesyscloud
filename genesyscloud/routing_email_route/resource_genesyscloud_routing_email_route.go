package routing_email_route

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

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

	inboundRoutesMap, respCode, err := proxy.getAllRoutingEmailRoute(ctx, "", "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, "Failed to get routing email route", respCode)
	}

	for domainId, inboundRoutes := range *inboundRoutesMap {
		for _, inboundRoute := range inboundRoutes {
			resources[*inboundRoute.Id] = &resourceExporter.ResourceMeta{
				BlockLabel: *inboundRoute.Pattern + domainId,
				IdPrefix:   domainId + "/",
			}
		}
	}
	return resources, nil
}

// createRoutingEmailRoute is used by the routing_email_route resource to create Genesys cloud routing email route
func createRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteProxy(sdkConfig)
	domainId := d.Get("domain_id").(string)

	routingEmailRoute := getRoutingEmailRouteFromResourceData(d)

	replyEmail, err := validateSdkReplyEmailAddress(d)
	// Checking the self_reference_route flag and routeId rules
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "Error occurred while validating the reply email address when creating the record", err)
	}

	replyDomainID, replyRouteID, _ := extractReplyEmailAddressValue(d)

	// If the isSelfReferenceRoute() is set to false, we use the route id provided by the terraform script
	if replyEmail && !isSelfReferenceRouteSet(d) {
		routingEmailRoute.ReplyEmailAddress = buildReplyEmailAddress(replyDomainID, replyRouteID)
	}

	log.Printf("Creating routing email route %s", d.Id())
	inboundRoute, resp, err := proxy.createRoutingEmailRoute(ctx, domainId, &routingEmailRoute)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create routing email route %s error: %s", *routingEmailRoute.FromName, err), resp)
	}

	d.SetId(*inboundRoute.Id)
	log.Printf("Created routing email route %s", *inboundRoute.Id)

	// If the isSelfReferenceRoute() is set to true we need grab the route id for the route and reapply the reply address,
	if replyEmail && isSelfReferenceRouteSet(d) {
		inboundRoute.ReplyEmailAddress = buildReplyEmailAddress(replyDomainID, *inboundRoute.Id)
		_, resp, err = proxy.updateRoutingEmailRoute(ctx, *inboundRoute.Id, domainId, inboundRoute)

		if err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Created routing email route %v %s %s, but failed to update the reply answer route to itself | error: %s", inboundRoute.Pattern, domainId, *inboundRoute.Id, err), resp)
		}
	}

	return readRoutingEmailRoute(ctx, d, meta)
}

// readRoutingEmailRoute is used by the routing_email_route resource to read an routing email route from genesys cloud
func readRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingEmailRoute(), constants.ConsistencyChecks(), ResourceType)
	domainId := d.Get("domain_id").(string)

	log.Printf("Reading routing email route %s", d.Id())
	var route *platformclientv2.Inboundroute

	// The normal GET route API has a long cache TTL (5 minutes) which can result in stale data.
	// This can be bypassed by issuing a domain query instead.
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		inboundRoutesMap, resp, getErr := proxy.getAllRoutingEmailRoute(ctx, domainId, "")
		if getErr != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read routing email route %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read routing email route %s | error: %s", d.Id(), getErr), resp))
		}

		for _, inboundRoutes := range *inboundRoutesMap {
			for _, queryRoute := range inboundRoutes {
				if queryRoute.Id != nil && *queryRoute.Id == d.Id() {
					route = &queryRoute
					break
				}
			}
		}
		if route == nil {
			d.SetId("")
			return nil
		}

		resourcedata.SetNillableValue(d, "pattern", route.Pattern)
		resourcedata.SetNillableReference(d, "queue_id", route.Queue)
		resourcedata.SetNillableValue(d, "priority", route.Priority)
		resourcedata.SetNillableReference(d, "language_id", route.Language)
		resourcedata.SetNillableValue(d, "from_name", route.FromName)
		resourcedata.SetNillableValue(d, "from_email", route.FromEmail)
		resourcedata.SetNillableReference(d, "flow_id", route.Flow)
		resourcedata.SetNillableValueWithInterfaceArrayWithFunc(d, "auto_bcc", route.AutoBcc, flattenAutoBccEmailAddress)
		resourcedata.SetNillableReference(d, "spam_flow_id", route.SpamFlow)

		if route.Skills != nil {
			_ = d.Set("skill_ids", util.SdkDomainEntityRefArrToSet(*route.Skills))
		} else {
			_ = d.Set("skill_ids", nil)
		}

		if route.ReplyEmailAddress != nil {
			flattenedEmails := flattenReplyEmailAddress(*route.ReplyEmailAddress)
			_, _, selfReferenceRoute := extractReplyEmailAddressValue(d)

			//Set the self_reference_route
			flattenedEmails["self_reference_route"] = selfReferenceRoute

			//If the reply points back to the route then set the route_id to nil because we dont need to set the
			if selfReferenceRoute {
				flattenedEmails["route_id"] = nil
			}

			_ = d.Set("reply_email_address", []interface{}{flattenedEmails})
		} else {
			_ = d.Set("reply_email_address", nil)
		}

		log.Printf("Read routing email route %s", d.Id())
		return cc.CheckState(d)
	})
}

// updateRoutingEmailRoute is used by the routing_email_route resource to update an routing email route in Genesys Cloud
func updateRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getRoutingEmailRouteProxy(sdkConfig)
	domainId := d.Get("domain_id").(string)

	routingEmailRoute := getRoutingEmailRouteFromResourceData(d)

	//Checking the self_reference_route flag and routeId rules
	replyEmail, err := validateSdkReplyEmailAddress(d)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, "Error occurred while validating the reply email address while trying to update the record", err)
	}

	replyDomainID, replyRouteID, _ := extractReplyEmailAddressValue(d)

	if replyEmail {
		if isSelfReferenceRouteSet(d) {
			routingEmailRoute.ReplyEmailAddress = buildReplyEmailAddress(replyDomainID, d.Id())
		} else if !isSelfReferenceRouteSet(d) {
			routingEmailRoute.ReplyEmailAddress = buildReplyEmailAddress(replyDomainID, replyRouteID)
		}
	}

	log.Printf("Updating routing email route %s", d.Id())
	inboundRoute, resp, err := proxy.updateRoutingEmailRoute(ctx, d.Id(), domainId, &routingEmailRoute)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update routing email route %s error: %s", *routingEmailRoute.FromName, err), resp)
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
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete routing email route %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 180*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getRoutingEmailRouteById(ctx, domainId, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted routing email route %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting routing email route %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("routing email route %s still exists", d.Id()), resp))
	})
}
