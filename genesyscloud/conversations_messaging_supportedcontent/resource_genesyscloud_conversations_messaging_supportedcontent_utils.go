package supported_content

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
		supportedContent.MediaTypes = buildMediaTypes(d.Get("media_types").(*schema.Set))
	}
	return supportedContent
}

// buildMediaTypes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatype
func buildMediaTypes(mediaTypes *schema.Set) *platformclientv2.Mediatypes {
	var sdkMediaType platformclientv2.Mediatypes
	if mediaTypes == nil {
		return nil
	}

	mediaTypeList := mediaTypes.List()
	if len(mediaTypeList) > 0 {
		mediaTypesMap := mediaTypeList[0].(map[string]interface{})

		if mediaAllow := mediaTypesMap["allow"]; mediaAllow != nil {
			sdkMediaType.Allow = buildAllowedMediaTypeAccess(mediaAllow.(*schema.Set))
		}
	}
	return &sdkMediaType
}

// buildMediaTypeAccesss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatypeaccess
func buildAllowedMediaTypeAccess(mediaTypeAccess *schema.Set) *platformclientv2.Mediatypeaccess {
	var sdkMediaTypeAccess platformclientv2.Mediatypeaccess
	if mediaTypeAccess == nil {
		return nil
	}

	mediaTypeAccessList := mediaTypeAccess.List()

	if len(mediaTypeAccessList) > 0 {
		mediaTypeAccessMap := mediaTypeAccessList[0].(map[string]interface{})

		if mediaInbound := mediaTypeAccessMap["inbound"].([]interface{}); len(mediaInbound) > 0 {
			sdkMediaTypeAccess.Inbound = buildInboundOutboundMediaTypes(mediaInbound)
		}

		if mediaOutbound := mediaTypeAccessMap["outbound"].([]interface{}); len(mediaOutbound) > 0 {
			sdkMediaTypeAccess.Outbound = buildInboundOutboundMediaTypes(mediaOutbound)
		}
	}
	return &sdkMediaTypeAccess
}

// buildMediaTypess maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediatypes
func buildInboundOutboundMediaTypes(mediaTypes []interface{}) *[]platformclientv2.Mediatype {
	if mediaTypes == nil {
		return nil
	}

	mediaTypesSlice := make([]platformclientv2.Mediatype, 0)
	for _, mediaType := range mediaTypes {
		var sdkMediaTypes platformclientv2.Mediatype
		mediaTypesMap, ok := mediaType.(map[string]interface{})
		if !ok {
			continue
		}

		if fieldType := mediaTypesMap["type"].(string); fieldType != "" {
			sdkMediaTypes.VarType = &fieldType
		}

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
func flattenMediaTypes(mediaTypes *platformclientv2.Mediatypes) *schema.Set {
	if mediaTypes == nil {
		return nil
	}

	mediaAllowedSet := schema.NewSet(schema.HashResource(mediaTypesResource), []interface{}{})
	mediaTypesMap := make(map[string]interface{})

	if mediaTypes.Allow != nil {
		mediaTypesMap["allow"] = *mediaTypes.Allow
	}

	mediaAllowedSet.Add(mediaTypesMap)
	return mediaAllowedSet
}
