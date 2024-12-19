package integration_facebook

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

/*
The resource_genesyscloud_integration_facebook_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getIntegrationFacebookFromResourceData maps data from schema ResourceData object to a platformclientv2.Facebookintegrationrequest
func getIntegrationFacebookFromResourceData(d *schema.ResourceData) platformclientv2.Facebookintegrationrequest {

	supportedContentId := d.Get("supported_content_id").(string)
	messagingContentId := d.Get("messaging_setting_id").(string)
	pageAccessToken := d.Get("page_access_token").(string)
	userAccessToken := d.Get("user_access_token").(string)
	pageId := d.Get("page_id").(string)
	appId := d.Get("app_id").(string)
	appSecret := d.Get("app_secret").(string)

	return platformclientv2.Facebookintegrationrequest{
		Name:             platformclientv2.String(d.Get("name").(string)),
		SupportedContent: &platformclientv2.Supportedcontentreference{Id: &supportedContentId},
		MessagingSetting: &platformclientv2.Messagingsettingrequestreference{Id: &messagingContentId},
		PageAccessToken:  &pageAccessToken,
		UserAccessToken:  &userAccessToken,
		PageId:           &pageId,
		AppId:            &appId,
		AppSecret:        &appSecret,
	}

}

func generateFacebookIntegrationResource(
	resourceLabel string,
	name string,
	supportedContentId string,
	messagingSettingId string,
	pageAccessToken string,
	userAccessToken string,
	pageId string,
	appId string,
	appSecret string) string {
	return fmt.Sprintf(`
	resource "genesyscloud_integration_facebook" "%s" {
		name = "%s"
		supported_content_id = %s
		messaging_setting_id = %s
		page_access_token = "%s"
		user_access_token = "%s"
		page_id = "%s"
		app_id = "%s"
		app_secret = "%s"
	}
	`, resourceLabel, name, supportedContentId, messagingSettingId, pageAccessToken, userAccessToken, pageId, appId, appSecret)
}
