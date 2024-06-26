package responsemanagement_response

import (
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func getResponseFromResourceData(d *schema.ResourceData) platformclientv2.Response {
	interactionType := d.Get("interaction_type").(string)
	substitutionsSchema := d.Get("substitutions_schema_id").(string)
	responseType := d.Get("response_type").(string)
	messagingTemplate := d.Get("messaging_template").(*schema.Set)

	response := platformclientv2.Response{
		Name:          platformclientv2.String(d.Get("name").(string)),
		Libraries:     util.BuildSdkDomainEntityRefArr(d, "library_ids"),
		Texts:         buildResponseTexts(d.Get("texts").(*schema.Set)),
		Substitutions: buildResponseSubstitutions(d.Get("substitutions").(*schema.Set)),
		Assets:        buildAddressableEntityRefs(d.Get("asset_ids").(*schema.Set)),
		Footer:        buildFooterTemplate(d.Get("footer").(*schema.Set)),
	}

	if interactionType != "" {
		response.InteractionType = &interactionType
	}
	if substitutionsSchema != "" {
		response.SubstitutionsSchema = &platformclientv2.Jsonschemadocument{Id: &substitutionsSchema}
	}
	if responseType != "" {
		response.ResponseType = &responseType
	}
	// Need to check messaging template like this to avoid the responseType being giving a default value
	if messagingTemplate.Len() > 0 {
		response.MessagingTemplate = buildMessagingTemplate(messagingTemplate)
	}

	return response
}

func buildResponseTexts(responseTexts *schema.Set) *[]platformclientv2.Responsetext {
	if responseTexts == nil {
		return nil
	}

	sdkResponseTexts := make([]platformclientv2.Responsetext, 0)
	responseTextList := responseTexts.List()
	for _, responseText := range responseTextList {
		var sdkResponseText platformclientv2.Responsetext
		responseTextMap := responseText.(map[string]interface{})

		resourcedata.BuildSDKStringValueIfNotNil(&sdkResponseText.Content, responseTextMap, "content")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResponseText.ContentType, responseTextMap, "content_type")

		sdkResponseTexts = append(sdkResponseTexts, sdkResponseText)
	}
	return &sdkResponseTexts
}

func buildResponseSubstitutions(responseSubstitutions *schema.Set) *[]platformclientv2.Responsesubstitution {
	if responseSubstitutions == nil {
		return nil
	}

	sdkResponseSubstitutions := make([]platformclientv2.Responsesubstitution, 0)
	responseSubstitutionList := responseSubstitutions.List()
	for _, responseSubstitution := range responseSubstitutionList {
		var sdkResponseSubstitution platformclientv2.Responsesubstitution
		responseSubstitutionMap := responseSubstitution.(map[string]interface{})

		sdkResponseSubstitution.Id = platformclientv2.String(responseSubstitutionMap["id"].(string))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResponseSubstitution.Description, responseSubstitutionMap, "description")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResponseSubstitution.DefaultValue, responseSubstitutionMap, "default_value")

		sdkResponseSubstitutions = append(sdkResponseSubstitutions, sdkResponseSubstitution)
	}
	return &sdkResponseSubstitutions
}

func buildWhatsappDefinition(whatsappDefinition *schema.Set) *platformclientv2.Whatsappdefinition {
	if whatsappDefinition == nil {
		return nil
	}

	var sdkWhatsappDefinition platformclientv2.Whatsappdefinition
	whatsappDefinitionList := whatsappDefinition.List()
	if len(whatsappDefinitionList) > 0 {
		whatsappDefinitionMap := whatsappDefinitionList[0].(map[string]interface{})

		resourcedata.BuildSDKStringValueIfNotNil(&sdkWhatsappDefinition.Name, whatsappDefinitionMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkWhatsappDefinition.Namespace, whatsappDefinitionMap, "namespace")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkWhatsappDefinition.Language, whatsappDefinitionMap, "language")
	}

	return &sdkWhatsappDefinition
}

func buildFooterTemplate(footerTemplate *schema.Set) *platformclientv2.Footertemplate {
	if footerTemplate == nil {
		return nil
	}

	footerTemplateList := footerTemplate.List()
	var sdkFooterTemplate platformclientv2.Footertemplate
	if len(footerTemplateList) > 0 {
		footerTemplateMap := footerTemplateList[0].(map[string]interface{})

		resourcedata.BuildSDKStringValueIfNotNil(&sdkFooterTemplate.VarType, footerTemplateMap, "type")
		if applicableResources, exists := footerTemplateMap["applicable_resources"].([]interface{}); exists {
			applicableResourcesList := lists.InterfaceListToStrings(applicableResources)
			sdkFooterTemplate.ApplicableResources = &applicableResourcesList
		}
	}
	return &sdkFooterTemplate
}

func buildMessagingTemplate(messagingTemplate *schema.Set) *platformclientv2.Messagingtemplate {
	if messagingTemplate == nil {
		return nil
	}

	var sdkMessagingTemplate platformclientv2.Messagingtemplate
	messagingTemplateList := messagingTemplate.List()
	if len(messagingTemplateList) > 0 {
		messagingTemplateMap := messagingTemplateList[0].(map[string]interface{})

		if whatsApp := messagingTemplateMap["whats_app"]; whatsApp != nil {
			sdkMessagingTemplate.WhatsApp = buildWhatsappDefinition(whatsApp.(*schema.Set))
		}
	}

	return &sdkMessagingTemplate
}

func buildAddressableEntityRefs(addressableEntityRef *schema.Set) *[]platformclientv2.Addressableentityref {
	if addressableEntityRef == nil {
		return nil
	}

	strList := lists.SetToStringList(addressableEntityRef)
	if strList == nil {
		return nil
	}

	addressableEntityRefs := make([]platformclientv2.Addressableentityref, len(*strList))
	for i, id := range *strList {
		tempId := id
		addressableEntityRefs[i] = platformclientv2.Addressableentityref{Id: &tempId}
	}
	return &addressableEntityRefs
}

func flattenResponseTexts(responseTexts *[]platformclientv2.Responsetext) *schema.Set {
	if len(*responseTexts) == 0 {
		return nil
	}

	responseTextSet := schema.NewSet(schema.HashResource(responsetextResource), []interface{}{})
	for _, responseText := range *responseTexts {
		responseTextMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(responseTextMap, "content", responseText.Content)
		resourcedata.SetMapValueIfNotNil(responseTextMap, "content_type", responseText.ContentType)

		responseTextSet.Add(responseTextMap)
	}

	return responseTextSet
}

func flattenResponseSubstitutions(responseSubstitutions *[]platformclientv2.Responsesubstitution) *schema.Set {
	if len(*responseSubstitutions) == 0 {
		return nil
	}

	responseSubstitutionSet := schema.NewSet(schema.HashResource(substitutionResource), []interface{}{})
	for _, responseSubstitution := range *responseSubstitutions {
		responseSubstitutionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(responseSubstitutionMap, "id", responseSubstitution.Id)
		resourcedata.SetMapValueIfNotNil(responseSubstitutionMap, "description", responseSubstitution.Description)
		resourcedata.SetMapValueIfNotNil(responseSubstitutionMap, "default_value", responseSubstitution.DefaultValue)

		responseSubstitutionSet.Add(responseSubstitutionMap)
	}

	return responseSubstitutionSet
}

func flattenWhatsappDefinition(whatsappDefinition *platformclientv2.Whatsappdefinition) *schema.Set {
	if whatsappDefinition == nil {
		return nil
	}

	whatsappDefinitionSet := schema.NewSet(schema.HashResource(whatsappDefinitionResource), []interface{}{})
	whatsappDefinitionMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(whatsappDefinitionMap, "name", whatsappDefinition.Name)
	resourcedata.SetMapValueIfNotNil(whatsappDefinitionMap, "namespace", whatsappDefinition.Namespace)
	resourcedata.SetMapValueIfNotNil(whatsappDefinitionMap, "language", whatsappDefinition.Language)

	whatsappDefinitionSet.Add(whatsappDefinitionMap)

	return whatsappDefinitionSet
}

func flattenFooterTemplate(footerTemplate *platformclientv2.Footertemplate) *schema.Set {
	if footerTemplate == nil {
		return nil
	}

	footerTemplateSet := schema.NewSet(schema.HashResource(footerResource), []interface{}{})
	footerTemplateMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(footerTemplateMap, "type", footerTemplate.VarType)
	if footerTemplate.ApplicableResources != nil {
		footerTemplateMap["applicable_resources"] = lists.StringListToInterfaceList(*footerTemplate.ApplicableResources)
	}

	footerTemplateSet.Add(footerTemplateMap)
	return footerTemplateSet
}

func flattenMessagingTemplate(messagingTemplate *platformclientv2.Messagingtemplate) *schema.Set {
	if messagingTemplate == nil {
		return nil
	}

	messagingTemplateSet := schema.NewSet(schema.HashResource(messagingtemplateResource), []interface{}{})
	messagingTemplateMap := make(map[string]interface{})

	if messagingTemplate.WhatsApp != nil {
		messagingTemplateMap["whats_app"] = flattenWhatsappDefinition(messagingTemplate.WhatsApp)
	}

	messagingTemplateSet.Add(messagingTemplateMap)

	return messagingTemplateSet
}

func flattenAddressableEntityRefs(addressableEntityRefs *[]platformclientv2.Addressableentityref) *schema.Set {
	addressableEntityRefList := make([]interface{}, len(*addressableEntityRefs))
	for i, v := range *addressableEntityRefs {
		addressableEntityRefList[i] = *v.Id
	}
	return schema.NewSet(schema.HashString, addressableEntityRefList)
}
