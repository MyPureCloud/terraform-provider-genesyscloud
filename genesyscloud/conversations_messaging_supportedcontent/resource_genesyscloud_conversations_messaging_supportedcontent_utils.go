package conversations_messaging_supportedcontent

import (
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_conversations_messaging_supportedcontent_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getSupportedContentFromResourceData maps data from schema ResourceData object to a platformclientv2.Supportedcontent
func getSupportedContentFromResourceData(d *schema.ResourceData) platformclientv2.Supportedcontent {
	var supportedContent platformclientv2.Supportedcontent
	supportedContent.Name = platformclientv2.String(d.Get("name").(string))
	if mediaTypes, ok := d.Get("media_types").([]interface{}); ok && len(mediaTypes) > 0 {
		supportedContent.MediaTypes = buildMediaTypes(d.Get("media_types").([]interface{}))
	}
	return supportedContent
}

// buildMediaTypes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatype
func buildMediaTypes(mediaTypes []interface{}) *platformclientv2.Mediatypes {
	var sdkMediaType platformclientv2.Mediatypes

	for _, mediaType := range mediaTypes {
		mediaTypesMap, ok := mediaType.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaType.Allow, mediaTypesMap, "allow", buildAllowedMediaTypeAccess)
	}
	return &sdkMediaType
}

// buildMediaTypeAccesss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatypeaccess
func buildAllowedMediaTypeAccess(mediaTypeAccess []interface{}) *platformclientv2.Mediatypeaccess {
	var sdkMediaTypeAccess platformclientv2.Mediatypeaccess
	for _, mediaType := range mediaTypeAccess {
		mediaTypeAccessMap, ok := mediaType.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaTypeAccess.Inbound, mediaTypeAccessMap, "inbound", buildInboundOutboundMediaTypes)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaTypeAccess.Outbound, mediaTypeAccessMap, "outbound", buildInboundOutboundMediaTypes)

	}

	return &sdkMediaTypeAccess
}

// buildMediaTypess maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatypes
func buildInboundOutboundMediaTypes(mediaTypes []interface{}) *[]platformclientv2.Mediatype {
	mediaTypesSlice := make([]platformclientv2.Mediatype, 0)
	for _, mediaType := range mediaTypes {
		var sdkMediaTypes platformclientv2.Mediatype
		mediaTypesMap, ok := mediaType.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkMediaTypes.VarType, mediaTypesMap, "type")

		mediaTypesSlice = append(mediaTypesSlice, sdkMediaTypes)
	}

	return &mediaTypesSlice
}

// flattenMediaTypes maps a Genesys Cloud *[]platformclientv2.Mediatype into a []interface{}
func flattenInboundOutboundMediaTypes(mediaTypes *[]platformclientv2.Mediatype) []interface{} {
	if len(*mediaTypes) == 0 {
		return nil
	}

	var mediaTypeList []interface{}
	for _, mediaType := range *mediaTypes {
		mediaTypeMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(mediaTypeMap, "type", mediaType.VarType)

		mediaTypeList = append(mediaTypeList, mediaTypeMap)
	}

	return mediaTypeList
}

// flattenMediaTypeAccesss maps a Genesys Cloud *[]platformclientv2.Mediatypeaccess into a []interface{}
func flattenAllowedMediaTypeAccess(mediaTypeAccess *platformclientv2.Mediatypeaccess) []interface{} {

	var mediaTypeAccessList []interface{}
	mediaTypeAccessMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaTypeAccessMap, "inbound", mediaTypeAccess.Inbound, flattenInboundOutboundMediaTypes)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaTypeAccessMap, "outbound", mediaTypeAccess.Outbound, flattenInboundOutboundMediaTypes)

	mediaTypeAccessList = append(mediaTypeAccessList, mediaTypeAccessMap)

	return mediaTypeAccessList
}

// flattenMediaTypess maps a Genesys Cloud *[]platformclientv2.Mediatypes into a []interface{}
func flattenMediaTypes(mediaTypes *platformclientv2.Mediatypes) []interface{} {
	var mediaTypesList []interface{}
	mediaTypesMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaTypesMap, "allow", mediaTypes.Allow, flattenAllowedMediaTypeAccess)

	mediaTypesList = append(mediaTypesList, mediaTypesMap)

	return mediaTypesList
}
