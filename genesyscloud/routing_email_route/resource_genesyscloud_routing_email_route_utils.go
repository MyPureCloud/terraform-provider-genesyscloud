package routing_email_route

import (
	"context"
	"fmt"
	"strings"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

/*
The resource_genesyscloud_routing_email_route_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getRoutingEmailRouteFromResourceData maps data from schema ResourceData object to a platformclientv2.Inboundroute
func getRoutingEmailRouteFromResourceData(d *schema.ResourceData) platformclientv2.Inboundroute {
	return platformclientv2.Inboundroute{
		Name:     platformclientv2.String(d.Get("name").(string)),
		Pattern:  platformclientv2.String(d.Get("pattern").(string)),
		QueueId:  gcloud.BuildSdkDomainEntityRef(d, "queue_id"),
		Priority: platformclientv2.Int(d.Get("priority").(int)),
		// TODO: Handle skills property
		LanguageId:           gcloud.BuildSdkDomainEntityRef(d, "language_id"),
		FromName:             platformclientv2.String(d.Get("from_name").(string)),
		FromEmail:            platformclientv2.String(d.Get("from_email").(string)),
		FlowId:               gcloud.BuildSdkDomainEntityRef(d, "flow_id"),
		ReplyEmailAddress:    buildQueueEmailAddress(d.Get("reply_email_address").([]interface{})),
		AutoBcc:              buildEmailAddresss(d.Get("auto_bcc").([]interface{})),
		SpamFlowId:           gcloud.BuildSdkDomainEntityRef(d, "spam_flow_id"),
		Signature:            buildSignature(d.Get("signature").([]interface{})),
		HistoryInclusion:     platformclientv2.String(d.Get("history_inclusion").(string)),
		AllowMultipleActions: platformclientv2.Bool(d.Get("allow_multiple_actions").(bool)),
	}
}

// buildQueueEmailAddresss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Queueemailaddress
func buildQueueEmailAddresss(queueEmailAddresss []interface{}) *[]platformclientv2.Queueemailaddress {
	queueEmailAddresssSlice := make([]platformclientv2.Queueemailaddress, 0)
	for _, queueEmailAddress := range queueEmailAddresss {
		var sdkQueueEmailAddress platformclientv2.Queueemailaddress
		queueEmailAddresssMap, ok := queueEmailAddress.(map[string]interface{})
		if !ok {
			continue
		}

		queueEmailAddresssSlice = append(queueEmailAddresssSlice, sdkQueueEmailAddress)
	}

	return &queueEmailAddresssSlice
}

// buildEmailAddresss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Emailaddress
func buildEmailAddresss(emailAddresss []interface{}) *[]platformclientv2.Emailaddress {
	emailAddresssSlice := make([]platformclientv2.Emailaddress, 0)
	for _, emailAddress := range emailAddresss {
		var sdkEmailAddress platformclientv2.Emailaddress
		emailAddresssMap, ok := emailAddress.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkEmailAddress.Email, emailAddresssMap, "email")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkEmailAddress.Name, emailAddresssMap, "name")

		emailAddresssSlice = append(emailAddresssSlice, sdkEmailAddress)
	}

	return &emailAddresssSlice
}

// buildSignatures maps an []interface{} into a Genesys Cloud *[]platformclientv2.Signature
func buildSignatures(signatures []interface{}) *[]platformclientv2.Signature {
	signaturesSlice := make([]platformclientv2.Signature, 0)
	for _, signature := range signatures {
		var sdkSignature platformclientv2.Signature
		signaturesMap, ok := signature.(map[string]interface{})
		if !ok {
			continue
		}

		sdkSignature.Enabled = platformclientv2.Bool(signaturesMap["enabled"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSignature.CannedResponseId, signaturesMap, "canned_response_id")
		sdkSignature.AlwaysIncluded = platformclientv2.Bool(signaturesMap["always_included"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSignature.InclusionType, signaturesMap, "inclusion_type")

		signaturesSlice = append(signaturesSlice, sdkSignature)
	}

	return &signaturesSlice
}

// buildInboundRoutes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Inboundroute
func buildInboundRoutes(inboundRoutes []interface{}) *[]platformclientv2.Inboundroute {
	inboundRoutesSlice := make([]platformclientv2.Inboundroute, 0)
	for _, inboundRoute := range inboundRoutes {
		var sdkInboundRoute platformclientv2.Inboundroute
		inboundRoutesMap, ok := inboundRoute.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundRoute.Name, inboundRoutesMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundRoute.Pattern, inboundRoutesMap, "pattern")
		sdkInboundRoute.QueueId = &platformclientv2.Domainentityref{Id: platformclientv2.String(inboundRoutesMap["queue_id"].(string))}
		sdkInboundRoute.Priority = platformclientv2.Int(inboundRoutesMap["priority"].(int))
		// TODO: Handle skills property
		sdkInboundRoute.LanguageId = &platformclientv2.Domainentityref{Id: platformclientv2.String(inboundRoutesMap["language_id"].(string))}
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundRoute.FromName, inboundRoutesMap, "from_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundRoute.FromEmail, inboundRoutesMap, "from_email")
		sdkInboundRoute.FlowId = &platformclientv2.Domainentityref{Id: platformclientv2.String(inboundRoutesMap["flow_id"].(string))}
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkInboundRoute.ReplyEmailAddress, inboundRoutesMap, "reply_email_address", buildQueueEmailAddress)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkInboundRoute.AutoBcc, inboundRoutesMap, "auto_bcc", buildEmailAddresss)
		sdkInboundRoute.SpamFlowId = &platformclientv2.Domainentityref{Id: platformclientv2.String(inboundRoutesMap["spam_flow_id"].(string))}
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkInboundRoute.Signature, inboundRoutesMap, "signature", buildSignature)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundRoute.HistoryInclusion, inboundRoutesMap, "history_inclusion")
		sdkInboundRoute.AllowMultipleActions = platformclientv2.Bool(inboundRoutesMap["allow_multiple_actions"].(bool))

		inboundRoutesSlice = append(inboundRoutesSlice, sdkInboundRoute)
	}

	return &inboundRoutesSlice
}

// buildQueueEmailAddresss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Queueemailaddress
func buildQueueEmailAddresss(queueEmailAddresss []interface{}) *[]platformclientv2.Queueemailaddress {
	queueEmailAddresssSlice := make([]platformclientv2.Queueemailaddress, 0)
	for _, queueEmailAddress := range queueEmailAddresss {
		var sdkQueueEmailAddress platformclientv2.Queueemailaddress
		queueEmailAddresssMap, ok := queueEmailAddress.(map[string]interface{})
		if !ok {
			continue
		}

		sdkQueueEmailAddress.DomainId = &platformclientv2.Domainentityref{Id: platformclientv2.String(queueEmailAddresssMap["domain_id"].(string))}
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueueEmailAddress.Route, queueEmailAddresssMap, "route", buildInboundRoute)

		queueEmailAddresssSlice = append(queueEmailAddresssSlice, sdkQueueEmailAddress)
	}

	return &queueEmailAddresssSlice
}

// flattenQueueEmailAddresss maps a Genesys Cloud *[]platformclientv2.Queueemailaddress into a []interface{}
func flattenQueueEmailAddresss(queueEmailAddresss *[]platformclientv2.Queueemailaddress) []interface{} {
	if len(*queueEmailAddresss) == 0 {
		return nil
	}

	var queueEmailAddressList []interface{}
	for _, queueEmailAddress := range *queueEmailAddresss {
		queueEmailAddressMap := make(map[string]interface{})

		queueEmailAddressList = append(queueEmailAddressList, queueEmailAddressMap)
	}

	return queueEmailAddressList
}

// flattenEmailAddresss maps a Genesys Cloud *[]platformclientv2.Emailaddress into a []interface{}
func flattenEmailAddresss(emailAddresss *[]platformclientv2.Emailaddress) []interface{} {
	if len(*emailAddresss) == 0 {
		return nil
	}

	var emailAddressList []interface{}
	for _, emailAddress := range *emailAddresss {
		emailAddressMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(emailAddressMap, "email", emailAddress.Email)
		resourcedata.SetMapValueIfNotNil(emailAddressMap, "name", emailAddress.Name)

		emailAddressList = append(emailAddressList, emailAddressMap)
	}

	return emailAddressList
}

// flattenSignatures maps a Genesys Cloud *[]platformclientv2.Signature into a []interface{}
func flattenSignatures(signatures *[]platformclientv2.Signature) []interface{} {
	if len(*signatures) == 0 {
		return nil
	}

	var signatureList []interface{}
	for _, signature := range *signatures {
		signatureMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(signatureMap, "enabled", signature.Enabled)
		resourcedata.SetMapValueIfNotNil(signatureMap, "canned_response_id", signature.CannedResponseId)
		resourcedata.SetMapValueIfNotNil(signatureMap, "always_included", signature.AlwaysIncluded)
		resourcedata.SetMapValueIfNotNil(signatureMap, "inclusion_type", signature.InclusionType)

		signatureList = append(signatureList, signatureMap)
	}

	return signatureList
}

// flattenInboundRoutes maps a Genesys Cloud *[]platformclientv2.Inboundroute into a []interface{}
func flattenInboundRoutes(inboundRoutes *[]platformclientv2.Inboundroute) []interface{} {
	if len(*inboundRoutes) == 0 {
		return nil
	}

	var inboundRouteList []interface{}
	for _, inboundRoute := range *inboundRoutes {
		inboundRouteMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "name", inboundRoute.Name)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "pattern", inboundRoute.Pattern)
		resourcedata.SetMapReferenceValueIfNotNil(inboundRouteMap, "queue_id", inboundRoute.QueueId)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "priority", inboundRoute.Priority)
		// TODO: Handle skills property
		resourcedata.SetMapReferenceValueIfNotNil(inboundRouteMap, "language_id", inboundRoute.LanguageId)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "from_name", inboundRoute.FromName)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "from_email", inboundRoute.FromEmail)
		resourcedata.SetMapReferenceValueIfNotNil(inboundRouteMap, "flow_id", inboundRoute.FlowId)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(inboundRouteMap, "reply_email_address", inboundRoute.ReplyEmailAddress, flattenQueueEmailAddress)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(inboundRouteMap, "auto_bcc", inboundRoute.AutoBcc, flattenEmailAddresss)
		resourcedata.SetMapReferenceValueIfNotNil(inboundRouteMap, "spam_flow_id", inboundRoute.SpamFlowId)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(inboundRouteMap, "signature", inboundRoute.Signature, flattenSignature)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "history_inclusion", inboundRoute.HistoryInclusion)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "allow_multiple_actions", inboundRoute.AllowMultipleActions)

		inboundRouteList = append(inboundRouteList, inboundRouteMap)
	}

	return inboundRouteList
}

// flattenQueueEmailAddresss maps a Genesys Cloud *[]platformclientv2.Queueemailaddress into a []interface{}
func flattenQueueEmailAddresss(queueEmailAddresss *[]platformclientv2.Queueemailaddress) []interface{} {
	if len(*queueEmailAddresss) == 0 {
		return nil
	}

	var queueEmailAddressList []interface{}
	for _, queueEmailAddress := range *queueEmailAddresss {
		queueEmailAddressMap := make(map[string]interface{})

		resourcedata.SetMapReferenceValueIfNotNil(queueEmailAddressMap, "domain_id", queueEmailAddress.DomainId)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueEmailAddressMap, "route", queueEmailAddress.Route, flattenInboundRoute)

		queueEmailAddressList = append(queueEmailAddressList, queueEmailAddressMap)
	}

	return queueEmailAddressList
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
