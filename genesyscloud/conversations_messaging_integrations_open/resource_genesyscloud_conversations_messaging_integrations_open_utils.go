package conversations_messaging_integrations_open

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

/*
The resource_genesyscloud_conversations_messaging_integrations_open_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getConversationsMessagingIntegrationsOpenFromResourceData maps data from schema ResourceData object to a platformclientv2.Openintegrationrequest
func getConversationsMessagingIntegrationsOpenFromResourceData(d *schema.ResourceData) platformclientv2.Openintegrationrequest {

	var webhookHeaders map[string]string
	_ = json.Unmarshal([]byte(d.Get("webhook_headers").(string)), &webhookHeaders)

	supportedContentId := d.Get("supported_content_id").(string)
	messagingSettingId := d.Get("messaging_setting_id").(string)

	return platformclientv2.Openintegrationrequest{
		Name:                           platformclientv2.String(d.Get("name").(string)),
		SupportedContent:               &platformclientv2.Supportedcontentreference{Id: &supportedContentId},
		MessagingSetting:               &platformclientv2.Messagingsettingrequestreference{Id: &messagingSettingId},
		OutboundNotificationWebhookUrl: platformclientv2.String(d.Get("outbound_notification_webhook_url").(string)),
		OutboundNotificationWebhookSignatureSecretToken: platformclientv2.String(d.Get("outbound_notification_webhook_signature_secret_token").(string)),
		WebhookHeaders: &webhookHeaders,
	}
}

// getConversationsMessagingIntegrationsOpenFromResourceDataForUpdate maps data from schema ResourceData object to a platformclientv2.Openintegrationrequest
func getConversationsMessagingIntegrationsOpenFromResourceDataForUpdate(d *schema.ResourceData) platformclientv2.Openintegrationupdaterequest {

	var webhookHeaders map[string]string
	_ = json.Unmarshal([]byte(d.Get("webhook_headers").(string)), &webhookHeaders)

	supportedContentId := d.Get("supported_content_id").(string)
	messagingSettingId := d.Get("messaging_setting_id").(string)

	return platformclientv2.Openintegrationupdaterequest{
		Name:                           platformclientv2.String(d.Get("name").(string)),
		SupportedContent:               &platformclientv2.Supportedcontentreference{Id: &supportedContentId},
		MessagingSetting:               &platformclientv2.Messagingsettingrequestreference{Id: &messagingSettingId},
		OutboundNotificationWebhookUrl: platformclientv2.String(d.Get("outbound_notification_webhook_url").(string)),
		OutboundNotificationWebhookSignatureSecretToken: platformclientv2.String(d.Get("outbound_notification_webhook_signature_secret_token").(string)),
		WebhookHeaders: &webhookHeaders,
	}
}

func GenerateWebhookHeadersProperties(
	headerType string,
	headerValue string,
) string {
	return "webhook_headers = " + util.GenerateJsonEncodedProperties(
		util.GenerateJsonProperty(headerType, strconv.Quote(headerValue)))
}

func GenerateConversationMessagingOpenResource(
	resourceLabel string,
	name string,
	supportedContentId string,
	messagingSettingId string,
	outboundNotificationWebhookUrl string,
	outboundNotificationWebhookSignatureSecretToken string,
	webhookHeaders ...string,
) string {
	return fmt.Sprintf(`
			resource "genesyscloud_conversations_messaging_integrations_open" "%s" {
				name = "%s"
				supported_content_id = %s
				messaging_setting_id = %s
				outbound_notification_webhook_url = "%s"
				outbound_notification_webhook_signature_secret_token = "%s"
				%s
			}
	`, resourceLabel, name, supportedContentId, messagingSettingId, outboundNotificationWebhookUrl, outboundNotificationWebhookSignatureSecretToken, strings.Join(webhookHeaders, "\n"))
}
