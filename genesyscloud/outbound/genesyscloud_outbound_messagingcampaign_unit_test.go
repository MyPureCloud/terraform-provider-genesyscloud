package outbound

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitFlattenEmailConfig(t *testing.T) {
	var (
		emailColumns           = []string{"foo", "bar"}
		contentTemplateId      = "content-template-id"
		fromAddressDomainId    = "from-address-domain-id"
		friendlyName           = "friendly-name"
		localPart              = "local-part"
		replyToAddressDomainId = "reply-to-address-domain-id"
		replyToAddressRouteId  = "reply-to-address-route-id"

		emailConfig = platformclientv2.Emailconfig{
			EmailColumns:    &emailColumns,
			ContentTemplate: &platformclientv2.Domainentityref{Id: &contentTemplateId},
			FromAddress: &platformclientv2.Fromemailaddress{
				Domain:       &platformclientv2.Domainentityref{Id: &fromAddressDomainId},
				FriendlyName: &friendlyName,
				LocalPart:    &localPart,
			},
			ReplyToAddress: &platformclientv2.Replytoemailaddress{
				Domain: &platformclientv2.Domainentityref{Id: &replyToAddressDomainId},
				Route:  &platformclientv2.Domainentityref{Id: &replyToAddressRouteId},
			},
		}
	)

	emailConfigList := flattenEmailConfig(emailConfig)
	if len(emailConfigList) != 1 {
		t.Errorf("len(emailConfigList) = %d; want 1", len(emailConfigList))
	}
	emailConfigMap, ok := emailConfigList[0].(map[string]any)
	if !ok {
		t.Errorf("Expected item in emailConfigList to be map[string]any, got %T", emailConfigList[0])
	}

	emailColumnsList, ok := emailConfigMap["email_columns"].([]any)
	if !ok {
		t.Errorf("Expected email_columns to be set in map")
	}
	if !lists.AreEquivalent(lists.InterfaceListToStrings(emailColumnsList), emailColumns) {
		t.Errorf("Expected email columns in flattened map to be %v, got %v", emailColumns, emailColumnsList)
	}

	flattenedContentTemplateId, _ := emailConfigMap["content_template_id"].(string)
	if flattenedContentTemplateId != contentTemplateId {
		t.Errorf("Expected content_template_id to be %s, got %s", contentTemplateId, flattenedContentTemplateId)
	}

	if err := verifyFromAddressIsFlattenedCorrectly(emailConfigMap, emailConfig); err != nil {
		t.Error(err)
	}

	if err := verifyReplyToAddressIsFlattenedCorrectly(emailConfigMap, emailConfig); err != nil {
		t.Error(err)
	}
}

func TestUnitBuildEmailConfig(t *testing.T) {
	var (
		emailColumns            = []any{"foo", "bar"}
		contentTemplateId       = "content-template-id"
		fromAddressDomainId     = "from-address-domain-id"
		fromAddressFriendlyName = "friendly-name"
		fromAddressLocalPart    = "local-part"
		replyToAddressDomainId  = "reply-to-address-domain-id"
		replyToAddressRouteId   = "reply-to-address-route-id"
	)
	resourceSchema := ResourceOutboundMessagingCampaign().Schema
	resourceDataMap := buildEmailConfigResourceDataMap(emailColumns, contentTemplateId, fromAddressDomainId, fromAddressFriendlyName,
		fromAddressLocalPart, replyToAddressDomainId, replyToAddressRouteId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

	emailConfig := buildEmailConfig(d)

	if !lists.AreEquivalent(lists.InterfaceListToStrings(emailColumns), *emailConfig.EmailColumns) {
		t.Errorf("Expected emailColumns to be %v, got %s", emailColumns, *emailConfig.EmailColumns)
	}
	assert.Equal(t, contentTemplateId, *emailConfig.ContentTemplate.Id, "Expected ContentTemplate.Id to be %s, got %s", contentTemplateId, *emailConfig.ContentTemplate.Id)
	assert.Equal(t, fromAddressDomainId, *emailConfig.FromAddress.Domain.Id, "Expected FromAddress.Domain.Id to be %s, got %s", fromAddressDomainId, *emailConfig.FromAddress.Domain.Id)
	assert.Equal(t, fromAddressFriendlyName, *emailConfig.FromAddress.FriendlyName, "Expected FromAddress.FriendlyName to be %s, got %s", fromAddressFriendlyName, *emailConfig.FromAddress.FriendlyName)
	assert.Equal(t, fromAddressLocalPart, *emailConfig.FromAddress.LocalPart, "Expected localPart to be %s, got %s", fromAddressLocalPart, *emailConfig.FromAddress.LocalPart)
	assert.Equal(t, replyToAddressDomainId, *emailConfig.ReplyToAddress.Domain.Id, "Expected ReplyToAddress.Domain.Id to be %s, got %s", replyToAddressDomainId, *emailConfig.ReplyToAddress.Domain.Id)
	assert.Equal(t, replyToAddressRouteId, *emailConfig.ReplyToAddress.Route.Id, "Expected ReplyToAddress.Route.Id to be %s, got %s", replyToAddressRouteId, *emailConfig.ReplyToAddress.Route.Id)
}

func buildEmailConfigResourceDataMap(
	emailColumns []any,
	contentTemplateId,
	fromAddressDomainId,
	fromAddressFriendlyName,
	fromAddressLocalPart,
	replyToAddressDomainId,
	replyToAddressRouteId string) map[string]any {
	return map[string]any{
		"email_config": []any{
			map[string]any{
				"email_columns":       emailColumns,
				"content_template_id": contentTemplateId,
				"from_address": []any{
					map[string]any{
						"domain_id":     fromAddressDomainId,
						"friendly_name": fromAddressFriendlyName,
						"local_part":    fromAddressLocalPart,
					},
				},
				"reply_to_address": []any{
					map[string]any{
						"domain_id": replyToAddressDomainId,
						"route_id":  replyToAddressRouteId,
					},
				},
			},
		},
	}
}

func verifyFromAddressIsFlattenedCorrectly(emailConfigMap map[string]any, emailConfig platformclientv2.Emailconfig) error {
	flattenedFromAddressList, ok := emailConfigMap["from_address"].([]any)
	if !ok || len(flattenedFromAddressList) != 1 {
		return fmt.Errorf("expected from_address to be an interface array in the flattened map")
	}
	flattenedFromAddressMap, ok := flattenedFromAddressList[0].(map[string]any)
	if !ok {
		return fmt.Errorf("expected item in list to be type map[string]any, got %T", flattenedFromAddressList[0])
	}

	flattenedFromAddressDomainId, _ := flattenedFromAddressMap["domain_id"].(string)
	if flattenedFromAddressDomainId != *emailConfig.FromAddress.Domain.Id {
		return fmt.Errorf("expected domain_id to be %s, got %s", *emailConfig.FromAddress.Domain.Id, flattenedFromAddressDomainId)
	}
	flattenedFromAddressFriendlyName, _ := flattenedFromAddressMap["friendly_name"].(string)
	if flattenedFromAddressFriendlyName != *emailConfig.FromAddress.FriendlyName {
		return fmt.Errorf("expected friendly_name to be %s, got %s", *emailConfig.FromAddress.FriendlyName, flattenedFromAddressFriendlyName)
	}
	flattenedLocalPart, _ := flattenedFromAddressMap["local_part"].(string)
	if flattenedLocalPart != *emailConfig.FromAddress.LocalPart {
		return fmt.Errorf("expected local_part to be %s, got %s", *emailConfig.FromAddress.LocalPart, flattenedLocalPart)
	}
	return nil
}

func verifyReplyToAddressIsFlattenedCorrectly(emailConfigMap map[string]any, emailConfig platformclientv2.Emailconfig) error {
	flattenedReplyToAddressList, ok := emailConfigMap["reply_to_address"].([]any)
	if !ok || len(flattenedReplyToAddressList) != 1 {
		return fmt.Errorf("expected reply_to_address to be an interface array in the flattened map with a length of 1")
	}
	flattenedReplyToAddressMap, ok := flattenedReplyToAddressList[0].(map[string]any)
	if !ok {
		return fmt.Errorf("expected item in flattenedReplyToAddressList to be type map[string]any, got %T", flattenedReplyToAddressList[0])
	}
	flattenedReplyToAddressDomainId, _ := flattenedReplyToAddressMap["domain_id"].(string)
	if flattenedReplyToAddressDomainId != *emailConfig.ReplyToAddress.Domain.Id {
		return fmt.Errorf("expected domain_id to be %s, got %s", *emailConfig.ReplyToAddress.Domain.Id, flattenedReplyToAddressDomainId)
	}
	flattenedReplyToAddressRouteId, _ := flattenedReplyToAddressMap["route_id"].(string)
	if flattenedReplyToAddressRouteId != *emailConfig.ReplyToAddress.Route.Id {
		return fmt.Errorf("expected route_id to be %s, got %s", *emailConfig.ReplyToAddress.Route.Id, flattenedReplyToAddressRouteId)
	}
	return nil
}
