package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v53/platformclientv2"
)

var (
	bccEmailResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"email": {
				Description: "Email address.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name associated with the email.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

func getAllRoutingEmailRoutes(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	routingAPI := platformclientv2.NewRoutingApiWithConfig(clientConfig)

	domains, _, getErr := routingAPI.GetRoutingEmailDomains()
	if getErr != nil {
		return nil, diag.Errorf("Failed to get routing email domains: %v", getErr)
	}

	if domains.Entities == nil || len(*domains.Entities) == 0 {
		return resources, nil
	}

	for _, domain := range *domains.Entities {
		for pageNum := 1; ; pageNum++ {
			routes, resp, getErr := routingAPI.GetRoutingEmailDomainRoutes(*domain.Id, 100, pageNum, "")
			if getErr != nil {
				if resp != nil && resp.StatusCode == 404 {
					// Domain not found
					break
				}
				return nil, diag.Errorf("Failed to get page of email routes: %v", getErr)
			}

			if routes.Entities == nil || len(*routes.Entities) == 0 {
				break
			}

			for _, route := range *routes.Entities {
				resources[*route.Id] = &ResourceMeta{
					Name:     *route.Pattern + *domain.Id,
					IdPrefix: *domain.Id + "/",
				}
			}
		}
	}

	return resources, nil
}

func routingEmailRouteExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllRoutingEmailRoutes),
		RefAttrs: map[string]*RefAttrSettings{
			"domain_id":                     {RefType: "genesyscloud_routing_email_domain"},
			"queue_id":                      {RefType: "genesyscloud_routing_queue"},
			"skill_ids":                     {RefType: "genesyscloud_routing_skill"},
			"language_id":                   {RefType: "genesyscloud_routing_language"},
			"flow_id":                       {}, // Ref type not yet defined
			"spam_flow_id":                  {}, // Ref type not yet defined
			"reply_email_address.domain_id": {RefType: "genesyscloud_routing_email_domain"},
			"reply_email_address.route_id":  {RefType: "genesyscloud_routing_email_route"},
		},
		RemoveIfMissing: map[string][]string{
			"reply_email_address": {"route_id"},
		},
	}
}

func resourceRoutingEmailRoute() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Routing Email Domain Route",

		CreateContext: createWithPooledClient(createRoutingEmailRoute),
		ReadContext:   readWithPooledClient(readRoutingEmailRoute),
		UpdateContext: updateWithPooledClient(updateRoutingEmailRoute),
		DeleteContext: deleteWithPooledClient(deleteRoutingEmailRoute),
		Importer: &schema.ResourceImporter{
			StateContext: importRoutingEmailRoute,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Description: "ID of the routing domain such as: 'example.com'",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"pattern": {
				Description: "The search pattern that the mailbox name should match.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"from_name": {
				Description: "The sender name to use for outgoing replies.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"from_email": {
				Description: "The sender email to use for outgoing replies.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"queue_id": {
				Description: "The queue to route the emails to. This should not be set if a flow_id is specified.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"priority": {
				Description: "The priority to use for routing.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"skill_ids": {
				Description: "The skills to use for routing.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"language_id": {
				Description: "The language to use for routing.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"flow_id": {
				Description: "The flow to use for processing the email. This should not be set if a queue_id is specified.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"reply_email_address": {
				Description: "The route to use for email replies.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_id": {
							Description: "Domain of the route.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"route_id": {
							Description: "ID of the route.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"auto_bcc": {
				Description: "The recipients that should be automatically blind copied on outbound emails associated with this route.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        bccEmailResource,
			},
			"spam_flow_id": {
				Description: "The flow to use for processing inbound emails that have been marked as spam.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func importRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	sdkRoute := platformclientv2.Inboundroute{
		Pattern:           &pattern,
		FromName:          &fromName,
		FromEmail:         &fromEmail,
		Queue:             buildSdkDomainEntityRef(d, "queue_id"),
		Priority:          &priority,
		Language:          buildSdkDomainEntityRef(d, "language_id"),
		Flow:              buildSdkDomainEntityRef(d, "flow_id"),
		SpamFlow:          buildSdkDomainEntityRef(d, "spam_flow_id"),
		Skills:            buildSdkDomainEntityRefArr(d, "skill_ids"),
		ReplyEmailAddress: buildSdkReplyEmailAddress(d),
		AutoBcc:           buildSdkAutoBccEmailAddresses(d),
	}

	log.Printf("Creating routing email route %s %s", pattern, domainID)
	route, _, err := routingAPI.PostRoutingEmailDomainRoutes(domainID, sdkRoute)
	if err != nil {
		return diag.Errorf("Failed to create routing email route %s: %s", pattern, err)
	}

	d.SetId(*route.Id)
	log.Printf("Created routing email route %s %s %s", pattern, domainID, *route.Id)

	return readRoutingEmailRoute(ctx, d, meta)
}

func readRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainID := d.Get("domain_id").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Reading routing email route %s", d.Id())

	// The normal GET route API has a long cache TTL (5 minutes) which can result in stale data.
	// This can be bypassed by issuing a domain query instead.
	var route *platformclientv2.Inboundroute
	for pageNum := 1; ; pageNum++ {
		routes, resp, getErr := routingAPI.GetRoutingEmailDomainRoutes(domainID, 100, pageNum, "")
		if getErr != nil {
			if resp != nil && resp.StatusCode == 404 {
				// Domain not found, so route also does not exist
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to get page of email routes for domain %s: %v", domainID, getErr)
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
		d.Set("skill_ids", sdkDomainEntityRefArrToSet(*route.Skills))
	} else {
		d.Set("skill_ids", nil)
	}

	if route.ReplyEmailAddress != nil && *route.ReplyEmailAddress != nil {
		d.Set("reply_email_address", []interface{}{flattenQueueEmailAddress(**route.ReplyEmailAddress)})
	} else {
		d.Set("reply_email_address", nil)
	}

	if route.AutoBcc != nil {
		d.Set("auto_bcc", flattenAutoBccEmailAddresses(*route.AutoBcc))
	} else {
		d.Set("auto_bcc", nil)
	}

	log.Printf("Read routing email route %s", d.Id())
	return nil
}

func updateRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()
	domainID := d.Get("domain_id").(string)
	pattern := d.Get("pattern").(string)
	fromName := d.Get("from_name").(string)
	fromEmail := d.Get("from_email").(string)
	priority := d.Get("priority").(int)

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Updating email route %s", d.Id())

	_, _, err := routingAPI.PutRoutingEmailDomainRoute(domainID, d.Id(), platformclientv2.Inboundroute{
		Id:                &id,
		Pattern:           &pattern,
		FromName:          &fromName,
		FromEmail:         &fromEmail,
		Queue:             buildSdkDomainEntityRef(d, "queue_id"),
		Priority:          &priority,
		Language:          buildSdkDomainEntityRef(d, "language_id"),
		Flow:              buildSdkDomainEntityRef(d, "flow_id"),
		SpamFlow:          buildSdkDomainEntityRef(d, "spam_flow_id"),
		Skills:            buildSdkDomainEntityRefArr(d, "skill_ids"),
		ReplyEmailAddress: buildSdkReplyEmailAddress(d),
		AutoBcc:           buildSdkAutoBccEmailAddresses(d),
	})
	if err != nil {
		return diag.Errorf("Failed to update email route %s: %s", d.Id(), err)
	}

	log.Printf("Updated routing email route %s", d.Id())
	return readRoutingEmailRoute(ctx, d, meta)
}

func deleteRoutingEmailRoute(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainID := d.Get("domain_id").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	log.Printf("Deleting email route %s", d.Id())
	_, err := routingAPI.DeleteRoutingEmailDomainRoute(domainID, d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete email route %s: %s", d.Id(), err)
	}

	return nil
}

func buildSdkReplyEmailAddress(d *schema.ResourceData) **platformclientv2.Queueemailaddress {
	replyEmailAddress := d.Get("reply_email_address").([]interface{})
	if replyEmailAddress != nil && len(replyEmailAddress) > 0 {
		settingsMap := replyEmailAddress[0].(map[string]interface{})

		domainID := settingsMap["domain_id"].(string)
		routeID := settingsMap["route_id"].(string)

		// For some reason the SDK expects a pointer to a pointer for this property
		result := &platformclientv2.Queueemailaddress{
			Domain: &platformclientv2.Domainentityref{Id: &domainID},
			Route: &platformclientv2.Inboundroute{
				Id: &routeID,
			},
		}
		return &result
	}
	return nil
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
