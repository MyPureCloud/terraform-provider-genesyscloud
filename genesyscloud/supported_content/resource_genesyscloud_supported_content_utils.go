package supported_content

import (
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_supported_content_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getSupportedContentFromResourceData maps data from schema ResourceData object to a platformclientv2.Supportedcontent
func getSupportedContentFromResourceData(d *schema.ResourceData) platformclientv2.Supportedcontent {
	return platformclientv2.Supportedcontent{
		Name:       platformclientv2.String(d.Get("name").(string)),
		MediaTypes: buildMediaTypes(d.Get("media_types").([]interface{})),
	}
}

// buildMediaTypes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatype
func buildMediaTypes(mediaTypes []interface{}) *platformclientv2.Mediatypes {
	var sdkMediaType platformclientv2.Mediatypes

	for _, mediaType := range mediaTypes {
		mediaTypesMap, ok := mediaType.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaType.Allow, mediaTypesMap, "allow", buildMediaTypeAccesss)
	}
	return &sdkMediaType
}

// buildMediaTypeAccesss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatypeaccess
func buildMediaTypeAccesss(mediaTypeAccesss []interface{}) *platformclientv2.Mediatypeaccess {
	var sdkMediaTypeAccess platformclientv2.Mediatypeaccess
	for _, mediaTypeAccess := range mediaTypeAccesss {
		mediaTypeAccesssMap, ok := mediaTypeAccess.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaTypeAccess.Inbound, mediaTypeAccesssMap, "inbound", buildMediaTypess)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaTypeAccess.Outbound, mediaTypeAccesssMap, "outbound", buildMediaTypess)

	}

	return &sdkMediaTypeAccess
}

// buildMediaTypess maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatypes
func buildMediaTypess(mediaTypess []interface{}) *[]platformclientv2.Mediatype {
	mediaTypessSlice := make([]platformclientv2.Mediatype, 0)
	for _, mediaTypes := range mediaTypess {
		var sdkMediaTypes platformclientv2.Mediatype
		mediaTypessMap, ok := mediaTypes.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkMediaTypes.VarType, mediaTypessMap, "type")

		mediaTypessSlice = append(mediaTypessSlice, sdkMediaTypes)
	}

	return &mediaTypessSlice
}

// flattenMediaTypes maps a Genesys Cloud *[]platformclientv2.Mediatype into a []interface{}
func flattenMediaTypes(mediaTypes *[]platformclientv2.Mediatype) []interface{} {
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
func flattenMediaTypeAccesss(mediaTypeAccesss *platformclientv2.Mediatypeaccess) []interface{} {

	var mediaTypeAccessList []interface{}
	mediaTypeAccessMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaTypeAccessMap, "inbound", mediaTypeAccesss.Inbound, flattenMediaTypes)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaTypeAccessMap, "outbound", mediaTypeAccesss.Outbound, flattenMediaTypes)

	mediaTypeAccessList = append(mediaTypeAccessList, mediaTypeAccessMap)

	return mediaTypeAccessList
}

// flattenMediaTypess maps a Genesys Cloud *[]platformclientv2.Mediatypes into a []interface{}
func flattenMediaTypess(mediaTypess *platformclientv2.Mediatypes) []interface{} {
	var mediaTypesList []interface{}
	mediaTypesMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaTypesMap, "allow", mediaTypess.Allow, flattenMediaTypeAccesss)

	mediaTypesList = append(mediaTypesList, mediaTypesMap)

	return mediaTypesList
}
