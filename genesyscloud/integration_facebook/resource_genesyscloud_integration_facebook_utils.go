package integration_facebook

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
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

	if pageAccessToken != "" {
		pageId = ""
		userAccessToken = ""
	} else {
		pageAccessToken = ""
	}
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
	resourceId string,
	name string,
	supportedContentId string,
	messagingSettingId string,
	pageAccessToken string,
	userAccessToken string,
	pageId string,
	appId string,
	appSecret string) string {
	return fmt.Sprintf(` resource "genesyscloud_integration_facebook" "%s" {
		name = "%s"
		supported_content_id = "%s"
		messaging_setting_id = "%s"
		page_access_token = "%s"
		user_access_token = "%s"
		page_id = "%s"
		app_id = "%s"
		app_secret = "%s"
	}
	`, resourceId, name, supportedContentId, messagingSettingId, pageAccessToken, userAccessToken, pageId, appId, appSecret)
}

/*

// buildMediaTypes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatype
func buildMediaTypes(mediaTypes []interface{}) *[]platformclientv2.Mediatype {
	mediaTypesSlice := make([]platformclientv2.Mediatype, 0)
	for _, mediaType := range mediaTypes {
		var sdkMediaType platformclientv2.Mediatype
		mediaTypesMap, ok := mediaType.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkMediaType.Type, mediaTypesMap, "type")

		mediaTypesSlice = append(mediaTypesSlice, sdkMediaType)
	}

	return &mediaTypesSlice
}

// buildMediaTypeAccesss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatypeaccess
func buildMediaTypeAccesss(mediaTypeAccesss []interface{}) *[]platformclientv2.Mediatypeaccess {
	mediaTypeAccesssSlice := make([]platformclientv2.Mediatypeaccess, 0)
	for _, mediaTypeAccess := range mediaTypeAccesss {
		var sdkMediaTypeAccess platformclientv2.Mediatypeaccess
		mediaTypeAccesssMap, ok := mediaTypeAccess.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaTypeAccess.Inbound, mediaTypeAccesssMap, "inbound", buildMediaTypes)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaTypeAccess.Outbound, mediaTypeAccesssMap, "outbound", buildMediaTypes)

		mediaTypeAccesssSlice = append(mediaTypeAccesssSlice, sdkMediaTypeAccess)
	}

	return &mediaTypeAccesssSlice
}

// buildMediaTypess maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatypes
func buildMediaTypess(mediaTypess []interface{}) *[]platformclientv2.Mediatypes {
	mediaTypessSlice := make([]platformclientv2.Mediatypes, 0)
	for _, mediaTypes := range mediaTypess {
		var sdkMediaTypes platformclientv2.Mediatypes
		mediaTypessMap, ok := mediaTypes.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaTypes.Allow, mediaTypessMap, "allow", buildMediaTypeAccess)

		mediaTypessSlice = append(mediaTypessSlice, sdkMediaTypes)
	}

	return &mediaTypessSlice
}

// buildSupportedContentReferences maps an []interface{} into a Genesys Cloud *[]platformclientv2.Supportedcontentreference
func buildSupportedContentReferences(supportedContentReferences []interface{}) *[]platformclientv2.Supportedcontentreference {
	supportedContentReferencesSlice := make([]platformclientv2.Supportedcontentreference, 0)
	for _, supportedContentReference := range supportedContentReferences {
		var sdkSupportedContentReference platformclientv2.Supportedcontentreference
		supportedContentReferencesMap, ok := supportedContentReference.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSupportedContentReference.Name, supportedContentReferencesMap, "name")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkSupportedContentReference.MediaTypes, supportedContentReferencesMap, "media_types", buildMediaTypes)

		supportedContentReferencesSlice = append(supportedContentReferencesSlice, sdkSupportedContentReference)
	}

	return &supportedContentReferencesSlice
}

// buildMessagingSettingRequestReferences maps an []interface{} into a Genesys Cloud *[]platformclientv2.Messagingsettingrequestreference
func buildMessagingSettingRequestReferences(messagingSettingRequestReferences []interface{}) *[]platformclientv2.Messagingsettingrequestreference {
	messagingSettingRequestReferencesSlice := make([]platformclientv2.Messagingsettingrequestreference, 0)
	for _, messagingSettingRequestReference := range messagingSettingRequestReferences {
		var sdkMessagingSettingRequestReference platformclientv2.Messagingsettingrequestreference
		messagingSettingRequestReferencesMap, ok := messagingSettingRequestReference.(map[string]interface{})
		if !ok {
			continue
		}

		messagingSettingRequestReferencesSlice = append(messagingSettingRequestReferencesSlice, sdkMessagingSettingRequestReference)
	}

	return &messagingSettingRequestReferencesSlice
}

// flattenMediaTypes maps a Genesys Cloud *[]platformclientv2.Mediatype into a []interface{}
func flattenMediaTypes(mediaTypes *[]platformclientv2.Mediatype) []interface{} {
	if len(*mediaTypes) == 0 {
		return nil
	}

	var mediaTypeList []interface{}
	for _, mediaType := range *mediaTypes {
		mediaTypeMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(mediaTypeMap, "type", mediaType.Type)

		mediaTypeList = append(mediaTypeList, mediaTypeMap)
	}

	return mediaTypeList
}

// flattenMediaTypeAccesss maps a Genesys Cloud *[]platformclientv2.Mediatypeaccess into a []interface{}
func flattenMediaTypeAccesss(mediaTypeAccesss *[]platformclientv2.Mediatypeaccess) []interface{} {
	if len(*mediaTypeAccesss) == 0 {
		return nil
	}

	var mediaTypeAccessList []interface{}
	for _, mediaTypeAccess := range *mediaTypeAccesss {
		mediaTypeAccessMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaTypeAccessMap, "inbound", mediaTypeAccess.Inbound, flattenMediaTypes)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaTypeAccessMap, "outbound", mediaTypeAccess.Outbound, flattenMediaTypes)

		mediaTypeAccessList = append(mediaTypeAccessList, mediaTypeAccessMap)
	}

	return mediaTypeAccessList
}

// flattenMediaTypess maps a Genesys Cloud *[]platformclientv2.Mediatypes into a []interface{}
func flattenMediaTypess(mediaTypess *[]platformclientv2.Mediatypes) []interface{} {
	if len(*mediaTypess) == 0 {
		return nil
	}

	var mediaTypesList []interface{}
	for _, mediaTypes := range *mediaTypess {
		mediaTypesMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaTypesMap, "allow", mediaTypes.Allow, flattenMediaTypeAccess)

		mediaTypesList = append(mediaTypesList, mediaTypesMap)
	}

	return mediaTypesList
}

// flattenSupportedContentReferences maps a Genesys Cloud *[]platformclientv2.Supportedcontentreference into a []interface{}
func flattenSupportedContentReferences(supportedContentReferences *[]platformclientv2.Supportedcontentreference) []interface{} {
	if len(*supportedContentReferences) == 0 {
		return nil
	}

	var supportedContentReferenceList []interface{}
	for _, supportedContentReference := range *supportedContentReferences {
		supportedContentReferenceMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(supportedContentReferenceMap, "name", supportedContentReference.Name)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(supportedContentReferenceMap, "media_types", supportedContentReference.MediaTypes, flattenMediaTypes)

		supportedContentReferenceList = append(supportedContentReferenceList, supportedContentReferenceMap)
	}

	return supportedContentReferenceList
}

// flattenMessagingSettingRequestReferences maps a Genesys Cloud *[]platformclientv2.Messagingsettingrequestreference into a []interface{}
func flattenMessagingSettingRequestReferences(messagingSettingRequestReferences *[]platformclientv2.Messagingsettingrequestreference) []interface{} {
	if len(*messagingSettingRequestReferences) == 0 {
		return nil
	}

	var messagingSettingRequestReferenceList []interface{}
	for _, messagingSettingRequestReference := range *messagingSettingRequestReferences {
		messagingSettingRequestReferenceMap := make(map[string]interface{})

		messagingSettingRequestReferenceList = append(messagingSettingRequestReferenceList, messagingSettingRequestReferenceMap)
	}

	return messagingSettingRequestReferenceList
}
*/
