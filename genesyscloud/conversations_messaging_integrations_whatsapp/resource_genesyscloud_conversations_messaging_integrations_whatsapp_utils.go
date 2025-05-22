package conversations_messaging_integrations_whatsapp

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

/*
The resource_genesyscloud_conversations_messaging_integrations_whatsapp_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getConversationsMessagingIntegrationsWhatsappFromResourceData maps data from schema ResourceData object to a platformclientv2.Whatsappembeddedsignupintegrationrequest
func getConversationsMessagingIntegrationsWhatsappFromResourceData(d *schema.ResourceData) platformclientv2.Whatsappembeddedsignupintegrationrequest {

	supportedContentId := d.Get("supported_content_id").(string)
	messagingSettingId := d.Get("messaging_setting_id").(string)

	return platformclientv2.Whatsappembeddedsignupintegrationrequest{
		Name:                      platformclientv2.String(d.Get("name").(string)),
		SupportedContent:          &platformclientv2.Supportedcontentreference{Id: &supportedContentId},
		MessagingSetting:          &platformclientv2.Messagingsettingrequestreference{Id: &messagingSettingId},
		EmbeddedSignupAccessToken: platformclientv2.String(d.Get("embedded_signup_access_token").(string)),
	}
}

func flattenActivateWhatsapp(phoneNumber string, pin string) []interface{} {

	activateInterface := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(activateInterface, "phone_number", &phoneNumber)
	resourcedata.SetMapValueIfNotNil(activateInterface, "pin", &pin)

	return []interface{}{activateInterface}
}

func GenerateConversationsMessagingIntegrationsWhatsappResource(
	resourceLabel string,
	name string,
	supportedContent string,
	messagingSetting string,
	embeddedSignupAccessToken string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
		resource "%s" "%s" {
			name = "%s"
			supported_content_id = %s
			messaging_setting_id = %s
			embedded_signup_access_token = "%s"
			%s
		}
	`, ResourceType, resourceLabel, name, supportedContent, messagingSetting, embeddedSignupAccessToken, strings.Join(nestedBlocks, "\n"))
}

func GenerateActivateConversationsMessagingIntegrationsWhatsappResource(
	phoneNumber string,
	pin string,
) string {
	return fmt.Sprintf(`
		activate_whatsapp {
			phone_number = "%s"
			pin = "%s"
		}
	`, phoneNumber, pin)
}

func GenerateConversationsMessagingIntegrationWhatsappDataSource(
	resourceLabel string,
	name string,
	dependsOn string,
) string {
	return fmt.Sprintf(`
		data "%s" "%s" {
			name = "%s"
			depends_on = [%s]
		}
	`, ResourceType, resourceLabel, name, dependsOn)
}
