package routing_email_route

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_routing_email_route_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getRoutingEmailRouteFromResourceData maps data from schema ResourceData object to a platformclientv2.Inboundroute
func getRoutingEmailRouteFromResourceData(d *schema.ResourceData) platformclientv2.Inboundroute {
	id := d.Id()
	return platformclientv2.Inboundroute{
		Id:        &id,
		Pattern:   platformclientv2.String(d.Get("pattern").(string)),
		Queue:     util.BuildSdkDomainEntityRef(d, "queue_id"),
		Priority:  platformclientv2.Int(d.Get("priority").(int)),
		Skills:    util.BuildSdkDomainEntityRefArr(d, "skill_ids"),
		Language:  util.BuildSdkDomainEntityRef(d, "language_id"),
		FromName:  platformclientv2.String(d.Get("from_name").(string)),
		FromEmail: buildFromEmail(d),
		Flow:      util.BuildSdkDomainEntityRef(d, "flow_id"),
		AutoBcc:   buildAutoBccEmailAddresses(d),
		SpamFlow:  util.BuildSdkDomainEntityRef(d, "spam_flow_id"),
	}
}

// Build Functions

func buildFromEmail(d *schema.ResourceData) *string {
	if d.Get("from_email") != "" {
		return platformclientv2.String(d.Get("from_email").(string))
	}
	return nil
}

func buildAutoBccEmailAddresses(d *schema.ResourceData) *[]platformclientv2.Emailaddress {
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

func buildReplyEmailAddress(domainID string, routeID string) *platformclientv2.Queueemailaddress {
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

// Flatten Functions

func flattenAutoBccEmailAddress(emailAddress *[]platformclientv2.Emailaddress) []interface{} {
	if len(*emailAddress) == 0 {
		return nil
	}

	var emailAddressList []interface{}
	for _, emailAddress := range *emailAddress {
		emailAddressMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(emailAddressMap, "email", emailAddress.Email)
		resourcedata.SetMapValueIfNotNil(emailAddressMap, "name", emailAddress.Name)

		emailAddressList = append(emailAddressList, emailAddressMap)
	}

	return emailAddressList
}

func flattenReplyEmailAddress(settings platformclientv2.Queueemailaddress) map[string]interface{} {
	settingsMap := make(map[string]interface{})
	resourcedata.SetMapReferenceValueIfNotNil(settingsMap, "domain_id", settings.Domain)

	if settings.Route != nil {
		route := *settings.Route
		settingsMap["route_id"] = *route.Id
	}

	return settingsMap
}

// Helper Functions

func validateSdkReplyEmailAddress(d *schema.ResourceData) (bool, error) {
	replyEmailAddress := d.Get("reply_email_address").([]interface{})
	if replyEmailAddress != nil && len(replyEmailAddress) > 0 {
		settingsMap := replyEmailAddress[0].(map[string]interface{})

		routeID := settingsMap["route_id"].(string)
		selfReferenceRoute := settingsMap["self_reference_route"].(bool)

		if selfReferenceRoute && routeID != "" {
			return true, fmt.Errorf("can not set a reply email address route id directly, if the self_reference_route value is set to true")
		}

		if !selfReferenceRoute && routeID == "" {
			return true, fmt.Errorf("you must provide reply email address route id if the self_reference_route value is set to false")
		}
		return true, nil
	}
	return false, nil
}

func extractReplyEmailAddressValue(d *schema.ResourceData) (string, string, bool) {
	replyEmailAddress := d.Get("reply_email_address").([]interface{})
	if replyEmailAddress != nil && len(replyEmailAddress) > 0 {
		settingsMap := replyEmailAddress[0].(map[string]interface{})

		return settingsMap["domain_id"].(string), settingsMap["route_id"].(string), settingsMap["self_reference_route"].(bool)
	}

	return "", "", false
}

func isSelfReferenceRouteSet(d *schema.ResourceData) bool {
	replyEmailAddress := d.Get("reply_email_address").([]interface{})
	if replyEmailAddress != nil && len(replyEmailAddress) > 0 {
		settingsMap := replyEmailAddress[0].(map[string]interface{})
		return settingsMap["self_reference_route"].(bool)
	}

	return false
}

func importRoutingEmailRoute(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	// Import must specify domain ID and route ID
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) < 2 {
		return nil, fmt.Errorf("invalid email route import ID %s", d.Id())
	}
	_ = d.Set("domain_id", idParts[0])
	d.SetId(idParts[1])
	return []*schema.ResourceData{d}, nil
}

func GenerateRoutingEmailRouteResource(
	resourceID string,
	domainID string,
	pattern string,
	fromName string,
	otherAttrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_routing_email_route" "%s" {
            domain_id = %s
            pattern = "%s"
            from_name = "%s"
            %s
        }
        `, resourceID, domainID, pattern, fromName, strings.Join(otherAttrs, "\n"))
}
