package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

func getAllRoutingEmailRoutes(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		domains, _, getErr := routingAPI.GetRoutingEmailDomains(pageSize, pageNum, false, "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get routing email domains: %v", getErr)
		}

		if domains.Entities == nil || len(*domains.Entities) == 0 {
			return resources, nil
		}

		for _, domain := range *domains.Entities {
			for pageNum := 1; ; pageNum++ {
				const pageSize = 100
				routes, resp, getErr := routingAPI.GetRoutingEmailDomainRoutes(*domain.Id, pageSize, pageNum, "")
				if getErr != nil {
					if util.IsStatus404(resp) {
						// Domain not found
						break
					}
					return nil, diag.Errorf("Failed to get page of email routes: %v", getErr)
				}

				if routes.Entities == nil || len(*routes.Entities) == 0 {
					break
				}

				for _, route := range *routes.Entities {
					resources[*route.Id] = &resourceExporter.ResourceMeta{
						Name:     *route.Pattern + *domain.Id,
						IdPrefix: *domain.Id + "/",
					}
				}
			}
		}
	}
}

func importRoutingEmailRoute(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	// Import must specify domain ID and route ID
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 {
		return nil, fmt.Errorf("Invalid email route import ID %s", d.Id())
	}
	d.Set("domain_id", idParts[0])
	d.SetId(idParts[1])
	return []*schema.ResourceData{d}, nil
}

func createRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainID := d.Get("domain_id").(string)
	pattern := d.Get("pattern").(string)
	fromName := d.Get("from_name").(string)
	fromEmail := d.Get("from_email").(string)
	priority := d.Get("priority").(int)

	replyDomainID, replyRouteID, _ := extractReplyEmailAddressValue(d)
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	//Checking the self_reference_route flag and routeId rules
	if err := validateSdkReplyEmailAddress(d); err != nil {
		return diag.Errorf("Error occurred while validating the reply email address when creating the record: %s", err)
	}

	sdkRoute := platformclientv2.Inboundroute{
		Pattern:   &pattern,
		FromName:  &fromName,
		FromEmail: &fromEmail,
		Queue:     util.BuildSdkDomainEntityRef(d, "queue_id"),
		Priority:  &priority,
		Language:  util.BuildSdkDomainEntityRef(d, "language_id"),
		Flow:      util.BuildSdkDomainEntityRef(d, "flow_id"),
		SpamFlow:  util.BuildSdkDomainEntityRef(d, "spam_flow_id"),
		Skills:    util.BuildSdkDomainEntityRefArr(d, "skill_ids"),
		AutoBcc:   buildSdkAutoBccEmailAddresses(d),
	}

	//If the isSelfReferenceRoute() is set to false, we use the route id provided by the terraform script
	if !isSelfReferenceRouteSet(d) {
		sdkRoute.ReplyEmailAddress = buildSdkReplyEmailAddress(replyDomainID, replyRouteID)

	}

	log.Printf("Creating routing email route %s %s", pattern, domainID)
	route, _, err := routingAPI.PostRoutingEmailDomainRoutes(domainID, sdkRoute)

	if err != nil {
		return diag.Errorf("Failed to create routing email route %s: %s", pattern, err)
	}

	d.SetId(*route.Id)
	log.Printf("Created routing email route %s %s %s", pattern, domainID, *route.Id)

	//If the isSelfReferenceRoute() is set to true we need grab the route id for the route and reapply the reply address,
	if isSelfReferenceRouteSet(d) {
		sdkRoute.ReplyEmailAddress = buildSdkReplyEmailAddress(replyDomainID, *route.Id)
		_, _, err = routingAPI.PutRoutingEmailDomainRoute(domainID, *route.Id, sdkRoute)

		if err != nil {
			return diag.Errorf("Created routing email route %s %s %s, but failed to update the reply answer route to itself. Error %s", pattern, domainID, *route.Id, err)
		}
	}

	return readRoutingEmailRoute(ctx, d, meta)
}

func readRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainID := d.Get("domain_id").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading routing email route %s", d.Id())

	// The normal GET route API has a long cache TTL (5 minutes) which can result in stale data.
	// This can be bypassed by issuing a domain query instead.
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		var route *platformclientv2.Inboundroute
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			routes, resp, getErr := routingAPI.GetRoutingEmailDomainRoutes(domainID, pageSize, pageNum, "")
			if getErr != nil {
				if util.IsStatus404(resp) {
					// Domain not found, so route also does not exist
					d.SetId("")
					return retry.RetryableError(fmt.Errorf("Failed to read routing email route %s: %s", d.Id(), getErr))
				}
				return retry.NonRetryableError(fmt.Errorf("Failed to read routing email route %s: %s", d.Id(), getErr))
			}

			if routes.Entities == nil || len(*routes.Entities) == 0 {
				break
			}

			for _, queryRoute := range *routes.Entities {
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

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceRoutingEmailRoute())

		if route.Pattern != nil {
			d.Set("pattern", *route.Pattern)
		} else {
			d.Set("pattern", nil)
		}

		if route.FromEmail != nil {
			d.Set("from_email", *route.FromEmail)
		} else {
			d.Set("from_email", nil)
		}

		if route.FromName != nil {
			d.Set("from_name", *route.FromName)
		} else {
			d.Set("from_name", nil)
		}

		if route.Priority != nil {
			d.Set("priority", *route.Priority)
		} else {
			d.Set("priority", nil)
		}

		if route.Queue != nil && route.Queue.Id != nil {
			d.Set("queue_id", *route.Queue.Id)
		} else {
			d.Set("queue_id", nil)
		}

		if route.Language != nil && route.Language.Id != nil {
			d.Set("language_id", *route.Language.Id)
		} else {
			d.Set("language_id", nil)
		}

		if route.Flow != nil && route.Flow.Id != nil {
			d.Set("flow_id", *route.Flow.Id)
		} else {
			d.Set("flow_id", nil)
		}

		if route.SpamFlow != nil && route.SpamFlow.Id != nil {
			d.Set("spam_flow_id", *route.SpamFlow.Id)
		} else {
			d.Set("spam_flow_id", nil)
		}

		if route.Skills != nil {
			d.Set("skill_ids", util.SdkDomainEntityRefArrToSet(*route.Skills))
		} else {
			d.Set("skill_ids", nil)
		}

		if route.ReplyEmailAddress != nil {
			flattenedEmails := flattenQueueEmailAddress(*route.ReplyEmailAddress)
			_, _, selfReferenceRoute := extractReplyEmailAddressValue(d)

			//Set the self_reference_route
			flattenedEmails["self_reference_route"] = selfReferenceRoute

			//If the reply points back to the route then set the route_id to nil because we dont need to set the
			if selfReferenceRoute {
				flattenedEmails["route_id"] = nil
			}

			d.Set("reply_email_address", []interface{}{flattenedEmails})
		} else {
			d.Set("reply_email_address", nil)
		}

		if route.AutoBcc != nil {
			d.Set("auto_bcc", flattenAutoBccEmailAddresses(*route.AutoBcc))
		} else {
			d.Set("auto_bcc", nil)
		}

		log.Printf("Read routing email route %s", d.Id())
		return cc.CheckState()
	})
}

func updateRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	domainID := d.Get("domain_id").(string)
	pattern := d.Get("pattern").(string)
	fromName := d.Get("from_name").(string)
	fromEmail := d.Get("from_email").(string)
	priority := d.Get("priority").(int)

	//Checking the self_reference_route flag and routeId rules
	if err := validateSdkReplyEmailAddress(d); err != nil {
		return diag.Errorf("Error occurred while validating the reply email address while trying to update the record: %s", err)
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	replyDomainID, replyRouteID, _ := extractReplyEmailAddressValue(d)

	sdkRoute := platformclientv2.Inboundroute{
		Id:        &id,
		Pattern:   &pattern,
		FromName:  &fromName,
		FromEmail: &fromEmail,
		Queue:     util.BuildSdkDomainEntityRef(d, "queue_id"),
		Priority:  &priority,
		Language:  util.BuildSdkDomainEntityRef(d, "language_id"),
		Flow:      util.BuildSdkDomainEntityRef(d, "flow_id"),
		SpamFlow:  util.BuildSdkDomainEntityRef(d, "spam_flow_id"),
		Skills:    util.BuildSdkDomainEntityRefArr(d, "skill_ids"),
		AutoBcc:   buildSdkAutoBccEmailAddresses(d),
	}

	if isSelfReferenceRouteSet(d) {
		sdkRoute.ReplyEmailAddress = buildSdkReplyEmailAddress(replyDomainID, d.Id())
	}

	if !isSelfReferenceRouteSet(d) {
		sdkRoute.ReplyEmailAddress = buildSdkReplyEmailAddress(replyDomainID, replyRouteID)
	}

	log.Printf("Updating email route %s", d.Id())

	_, _, err := routingAPI.PutRoutingEmailDomainRoute(domainID, d.Id(), sdkRoute)

	if err != nil {
		return diag.Errorf("Failed to update email route %s: %s", d.Id(), err)
	}

	log.Printf("Updated routing email route %s", d.Id())
	return readRoutingEmailRoute(ctx, d, meta)
}

func deleteRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainID := d.Get("domain_id").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting email route %s", d.Id())
	_, err := routingAPI.DeleteRoutingEmailDomainRoute(domainID, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete email route %s: %s", d.Id(), err)
	}

	return util.WithRetries(ctx, 60*time.Second, func() *retry.RetryError {
		_, resp, err := routingAPI.GetRoutingEmailDomainRoute(domainID, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Routing email domain route deleted
				log.Printf("Deleted Routing email domain route %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting Routing email domain route %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Routing email domain route %s still exists", d.Id()))
	})
}

func isSelfReferenceRouteSet(d *schema.ResourceData) bool {
	replyEmailAddress := d.Get("reply_email_address").([]interface{})
	if replyEmailAddress != nil && len(replyEmailAddress) > 0 {
		settingsMap := replyEmailAddress[0].(map[string]interface{})
		return settingsMap["self_reference_route"].(bool)
	}

	return false
}

func validateSdkReplyEmailAddress(d *schema.ResourceData) error {
	replyEmailAddress := d.Get("reply_email_address").([]interface{})
	if replyEmailAddress != nil && len(replyEmailAddress) > 0 {
		settingsMap := replyEmailAddress[0].(map[string]interface{})

		routeID := settingsMap["route_id"].(string)
		selfReferenceRoute := settingsMap["self_reference_route"].(bool)

		if selfReferenceRoute && routeID != "" {
			return fmt.Errorf("can not set a reply email address route id directly, if the self_reference_route value is set to true")
		}

		if !selfReferenceRoute && routeID == "" {
			return fmt.Errorf("you must provide reply email address route id if the self_reference_route value is set to false")
		}
	}

	return nil
}

func extractReplyEmailAddressValue(d *schema.ResourceData) (string, string, bool) {
	replyEmailAddress := d.Get("reply_email_address").([]interface{})
	if replyEmailAddress != nil && len(replyEmailAddress) > 0 {
		settingsMap := replyEmailAddress[0].(map[string]interface{})

		return settingsMap["domain_id"].(string), settingsMap["route_id"].(string), settingsMap["self_reference_route"].(bool)
	}

	return "", "", false
}

func buildSdkReplyEmailAddress(domainID string, routeID string) *platformclientv2.Queueemailaddress {
	// For some reason the SDK expects a pointer to a pointer for this property
	inboundRoute := &platformclientv2.Inboundroute{
		Id: &routeID,
	}
	result := platformclientv2.Queueemailaddress{
		Domain: &platformclientv2.Domainentityref{Id: &domainID},
		Route:  &inboundRoute,
	}
	return &result
}

func buildSdkAutoBccEmailAddresses(d *schema.ResourceData) *[]platformclientv2.Emailaddress {
	if bccAddresses := d.Get("auto_bcc"); bccAddresses != nil {
		bccAddressList := bccAddresses.(*schema.Set).List()
		sdkEmails := make([]platformclientv2.Emailaddress, len(bccAddressList))
		for i, configBcc := range bccAddressList {
			bccMap := configBcc.(map[string]interface{})
			bccEmail := bccMap["email"].(string)
			bccName := bccMap["name"].(string)

			sdkEmails[i] = platformclientv2.Emailaddress{
				Email: &bccEmail,
				Name:  &bccName,
			}
		}
		return &sdkEmails
	}
	return nil
}

func flattenAutoBccEmailAddresses(addresses []platformclientv2.Emailaddress) *schema.Set {
	addressSet := schema.NewSet(schema.HashResource(bccEmailResource), []interface{}{})
	for _, sdkEmail := range addresses {
		address := make(map[string]interface{})
		if sdkEmail.Name != nil {
			address["name"] = *sdkEmail.Name
		}
		if sdkEmail.Email != nil {
			address["email"] = *sdkEmail.Email
		}
		addressSet.Add(address)
	}
	return addressSet
}
