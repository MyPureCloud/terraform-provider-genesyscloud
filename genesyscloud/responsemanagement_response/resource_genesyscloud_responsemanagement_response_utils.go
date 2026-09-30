package responsemanagement_response

import (
	"fmt"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
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
		Form:          buildForm(d.Get("form").([]interface{})),
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
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResponseText.VarType, responseTextMap, "type")

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

// nestedBlockList safely reads a nested block (TypeList) out of a flattened resource map.
// It returns nil when the key is absent or holds a different type.
func nestedBlockList(targetMap map[string]interface{}, key string) []interface{} {
	if values, ok := targetMap[key].([]interface{}); ok {
		return values
	}
	return nil
}

// firstNestedBlock safely reads the single element of a MaxItems:1 nested block.
// The second return value reports whether the block was present.
func firstNestedBlock(targetMap map[string]interface{}, key string) (map[string]interface{}, bool) {
	list := nestedBlockList(targetMap, key)
	if len(list) == 0 {
		return nil, false
	}
	blockMap, ok := list[0].(map[string]interface{})
	return blockMap, ok
}

// nestedString safely reads a string out of a flattened resource map. Absent keys and empty
// strings both yield nil so the attribute is omitted from the API payload. Unlike
// resourcedata.BuildSDKStringValueIfNotNil this does not panic on a missing key.
func nestedString(targetMap map[string]interface{}, key string) *string {
	return resourcedata.GetNillableValueFromMap[string](targetMap, key, false)
}

// nestedBool safely reads a bool out of a flattened resource map, preserving an explicit false.
func nestedBool(targetMap map[string]interface{}, key string) *bool {
	return resourcedata.GetNillableValueFromMap[bool](targetMap, key, true)
}

// buildForm maps the `form` nested block into a Genesys Cloud *platformclientv2.Form
func buildForm(form []interface{}) *platformclientv2.Form {
	if len(form) == 0 {
		return nil
	}
	formMap, ok := form[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkForm platformclientv2.Form
	sdkForm.FormDescription = nestedString(formMap, "form_description")
	sdkForm.ShowSummary = nestedBool(formMap, "show_summary")
	sdkForm.FormPages = buildFormPages(nestedBlockList(formMap, "form_pages"))

	if receivedMessage, ok := firstNestedBlock(formMap, "received_message"); ok {
		sdkForm.ReceivedMessage = buildFormMessage(receivedMessage)
	}
	if replyMessage, ok := firstNestedBlock(formMap, "reply_message"); ok {
		sdkForm.ReplyMessage = buildFormMessage(replyMessage)
	}
	if introduction, ok := firstNestedBlock(formMap, "introduction"); ok {
		sdkForm.Introduction = buildFormIntroduction(introduction)
	}

	return &sdkForm
}

// buildFormMessage maps a `received_message`/`reply_message` block into a Genesys Cloud *platformclientv2.Formmessage
func buildFormMessage(formMessageMap map[string]interface{}) *platformclientv2.Formmessage {
	var sdkFormMessage platformclientv2.Formmessage

	sdkFormMessage.Title = nestedString(formMessageMap, "title")
	sdkFormMessage.Subtitle = nestedString(formMessageMap, "subtitle")
	sdkFormMessage.ImageUrl = nestedString(formMessageMap, "image_url")

	return &sdkFormMessage
}

// buildFormIntroduction maps an `introduction` block into a Genesys Cloud *platformclientv2.Formintroduction
func buildFormIntroduction(formIntroductionMap map[string]interface{}) *platformclientv2.Formintroduction {
	var sdkFormIntroduction platformclientv2.Formintroduction

	sdkFormIntroduction.Title = nestedString(formIntroductionMap, "title")
	sdkFormIntroduction.Subtitle = nestedString(formIntroductionMap, "subtitle")
	sdkFormIntroduction.ButtonText = nestedString(formIntroductionMap, "button_text")
	sdkFormIntroduction.ImageUrl = nestedString(formIntroductionMap, "image_url")

	return &sdkFormIntroduction
}

// buildFormPages maps the `form_pages` blocks into a Genesys Cloud *[]platformclientv2.Formpage
func buildFormPages(formPages []interface{}) *[]platformclientv2.Formpage {
	if formPages == nil {
		return nil
	}

	sdkFormPages := make([]platformclientv2.Formpage, 0, len(formPages))
	for _, formPage := range formPages {
		formPageMap, ok := formPage.(map[string]interface{})
		if !ok {
			continue
		}

		var sdkFormPage platformclientv2.Formpage
		sdkFormPage.Title = nestedString(formPageMap, "title")
		sdkFormPage.Subtitle = nestedString(formPageMap, "subtitle")
		sdkFormPage.PageComponents = buildFormPageComponents(nestedBlockList(formPageMap, "page_components"))

		sdkFormPages = append(sdkFormPages, sdkFormPage)
	}
	return &sdkFormPages
}

// buildFormPageComponents maps the `page_components` blocks into a Genesys Cloud *[]platformclientv2.Formpagecomponent
func buildFormPageComponents(pageComponents []interface{}) *[]platformclientv2.Formpagecomponent {
	if pageComponents == nil {
		return nil
	}

	sdkPageComponents := make([]platformclientv2.Formpagecomponent, 0, len(pageComponents))
	for _, pageComponent := range pageComponents {
		pageComponentMap, ok := pageComponent.(map[string]interface{})
		if !ok {
			continue
		}

		var sdkPageComponent platformclientv2.Formpagecomponent
		sdkPageComponent.FormComponentType = nestedString(pageComponentMap, "form_component_type")

		if listPicker, ok := firstNestedBlock(pageComponentMap, "list_picker"); ok {
			sdkPageComponent.ListPicker = buildFormListPicker(listPicker)
		}
		if datePicker, ok := firstNestedBlock(pageComponentMap, "date_picker"); ok {
			sdkPageComponent.DatePicker = buildFormDatePicker(datePicker)
		}
		if input, ok := firstNestedBlock(pageComponentMap, "input"); ok {
			sdkPageComponent.Input = buildFormInput(input)
		}
		if wheelPicker, ok := firstNestedBlock(pageComponentMap, "wheel_picker"); ok {
			sdkPageComponent.WheelPicker = buildFormWheelPicker(wheelPicker)
		}

		sdkPageComponents = append(sdkPageComponents, sdkPageComponent)
	}
	return &sdkPageComponents
}

// buildFormListPicker maps a `list_picker` block into a Genesys Cloud *platformclientv2.Formlistpicker
func buildFormListPicker(listPickerMap map[string]interface{}) *platformclientv2.Formlistpicker {
	var sdkListPicker platformclientv2.Formlistpicker
	sdkListPicker.Sections = buildFormListPickerSections(nestedBlockList(listPickerMap, "sections"))
	return &sdkListPicker
}

// buildFormListPickerSections maps the `sections` blocks into a Genesys Cloud *[]platformclientv2.Formlistpickersection
func buildFormListPickerSections(sections []interface{}) *[]platformclientv2.Formlistpickersection {
	if sections == nil {
		return nil
	}

	sdkSections := make([]platformclientv2.Formlistpickersection, 0, len(sections))
	for _, section := range sections {
		sectionMap, ok := section.(map[string]interface{})
		if !ok {
			continue
		}

		var sdkSection platformclientv2.Formlistpickersection
		sdkSection.Title = nestedString(sectionMap, "title")
		sdkSection.MultipleSelection = nestedBool(sectionMap, "multiple_selection")
		sdkSection.Items = buildFormListPickerItems(nestedBlockList(sectionMap, "items"))

		sdkSections = append(sdkSections, sdkSection)
	}
	return &sdkSections
}

// buildFormListPickerItems maps the list picker `items` blocks into a Genesys Cloud *[]platformclientv2.Formlistpickeritem
func buildFormListPickerItems(items []interface{}) *[]platformclientv2.Formlistpickeritem {
	if items == nil {
		return nil
	}

	sdkItems := make([]platformclientv2.Formlistpickeritem, 0, len(items))
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var sdkItem platformclientv2.Formlistpickeritem
		sdkItem.Title = nestedString(itemMap, "title")
		sdkItem.ImageUrl = nestedString(itemMap, "image_url")

		sdkItems = append(sdkItems, sdkItem)
	}
	return &sdkItems
}

// buildFormDatePicker maps a `date_picker` block into a Genesys Cloud *platformclientv2.Formdatepicker
func buildFormDatePicker(datePickerMap map[string]interface{}) *platformclientv2.Formdatepicker {
	var sdkDatePicker platformclientv2.Formdatepicker

	sdkDatePicker.Title = nestedString(datePickerMap, "title")
	sdkDatePicker.Subtitle = nestedString(datePickerMap, "subtitle")
	sdkDatePicker.DateDisplayFormat = nestedString(datePickerMap, "date_display_format")

	return &sdkDatePicker
}

// buildFormInput maps an `input` block into a Genesys Cloud *platformclientv2.Input
func buildFormInput(inputMap map[string]interface{}) *platformclientv2.Input {
	var sdkInput platformclientv2.Input

	sdkInput.Title = nestedString(inputMap, "title")
	sdkInput.Subtitle = nestedString(inputMap, "subtitle")
	sdkInput.PlaceholderText = nestedString(inputMap, "placeholder_text")
	sdkInput.KeyboardType = nestedString(inputMap, "keyboard_type")
	sdkInput.AutoCompleteType = nestedString(inputMap, "auto_complete_type")
	sdkInput.IsMultipleLine = nestedBool(inputMap, "is_multiple_line")
	sdkInput.IsRequired = nestedBool(inputMap, "is_required")

	return &sdkInput
}

// buildFormWheelPicker maps a `wheel_picker` block into a Genesys Cloud *platformclientv2.Wheelpicker
func buildFormWheelPicker(wheelPickerMap map[string]interface{}) *platformclientv2.Wheelpicker {
	var sdkWheelPicker platformclientv2.Wheelpicker
	sdkWheelPicker.Items = buildFormWheelPickerItems(nestedBlockList(wheelPickerMap, "items"))
	return &sdkWheelPicker
}

// buildFormWheelPickerItems maps the wheel picker `items` blocks into a Genesys Cloud *[]platformclientv2.Wheelpickeritem
func buildFormWheelPickerItems(items []interface{}) *[]platformclientv2.Wheelpickeritem {
	if items == nil {
		return nil
	}

	sdkItems := make([]platformclientv2.Wheelpickeritem, 0, len(items))
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var sdkItem platformclientv2.Wheelpickeritem
		sdkItem.Title = nestedString(itemMap, "title")
		sdkItem.Value = nestedString(itemMap, "value")

		sdkItems = append(sdkItems, sdkItem)
	}
	return &sdkItems
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

func buildAddressableEntityRefs(addressableEntityRef *schema.Set) *[]platformclientv2.Rmsassetaddressableref {
	if addressableEntityRef == nil {
		return nil
	}

	strList := lists.SetToStringList(addressableEntityRef)
	if strList == nil {
		return nil
	}

	addressableEntityRefs := make([]platformclientv2.Rmsassetaddressableref, len(*strList))
	for i, id := range *strList {
		tempId := id
		addressableEntityRefs[i] = platformclientv2.Rmsassetaddressableref{Id: &tempId}
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
		resourcedata.SetMapValueIfNotNil(responseTextMap, "type", responseText.VarType)

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

// flattenForm maps a Genesys Cloud *platformclientv2.Form into the `form` nested block
func flattenForm(form *platformclientv2.Form) []interface{} {
	if form == nil {
		return nil
	}

	formMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(formMap, "form_description", form.FormDescription)
	resourcedata.SetMapValueIfNotNil(formMap, "show_summary", form.ShowSummary)
	if form.ReceivedMessage != nil {
		formMap["received_message"] = flattenFormMessage(form.ReceivedMessage)
	}
	if form.ReplyMessage != nil {
		formMap["reply_message"] = flattenFormMessage(form.ReplyMessage)
	}
	if form.Introduction != nil {
		formMap["introduction"] = flattenFormIntroduction(form.Introduction)
	}
	if form.FormPages != nil {
		formMap["form_pages"] = flattenFormPages(form.FormPages)
	}

	return []interface{}{formMap}
}

// flattenFormMessage maps a Genesys Cloud *platformclientv2.Formmessage into a nested block
func flattenFormMessage(formMessage *platformclientv2.Formmessage) []interface{} {
	if formMessage == nil {
		return nil
	}

	formMessageMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(formMessageMap, "title", formMessage.Title)
	resourcedata.SetMapValueIfNotNil(formMessageMap, "subtitle", formMessage.Subtitle)
	resourcedata.SetMapValueIfNotNil(formMessageMap, "image_url", formMessage.ImageUrl)

	return []interface{}{formMessageMap}
}

// flattenFormIntroduction maps a Genesys Cloud *platformclientv2.Formintroduction into a nested block
func flattenFormIntroduction(formIntroduction *platformclientv2.Formintroduction) []interface{} {
	if formIntroduction == nil {
		return nil
	}

	formIntroductionMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(formIntroductionMap, "title", formIntroduction.Title)
	resourcedata.SetMapValueIfNotNil(formIntroductionMap, "subtitle", formIntroduction.Subtitle)
	resourcedata.SetMapValueIfNotNil(formIntroductionMap, "button_text", formIntroduction.ButtonText)
	resourcedata.SetMapValueIfNotNil(formIntroductionMap, "image_url", formIntroduction.ImageUrl)

	return []interface{}{formIntroductionMap}
}

// flattenFormPages maps a Genesys Cloud *[]platformclientv2.Formpage into the `form_pages` nested blocks
func flattenFormPages(formPages *[]platformclientv2.Formpage) []interface{} {
	if formPages == nil {
		return nil
	}

	formPageList := make([]interface{}, 0, len(*formPages))
	for _, formPage := range *formPages {
		formPageMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(formPageMap, "title", formPage.Title)
		resourcedata.SetMapValueIfNotNil(formPageMap, "subtitle", formPage.Subtitle)
		if formPage.PageComponents != nil {
			formPageMap["page_components"] = flattenFormPageComponents(formPage.PageComponents)
		}

		formPageList = append(formPageList, formPageMap)
	}
	return formPageList
}

// flattenFormPageComponents maps a Genesys Cloud *[]platformclientv2.Formpagecomponent into the `page_components` nested blocks
func flattenFormPageComponents(pageComponents *[]platformclientv2.Formpagecomponent) []interface{} {
	if pageComponents == nil {
		return nil
	}

	pageComponentList := make([]interface{}, 0, len(*pageComponents))
	for _, pageComponent := range *pageComponents {
		pageComponentMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(pageComponentMap, "form_component_type", pageComponent.FormComponentType)
		if pageComponent.ListPicker != nil {
			pageComponentMap["list_picker"] = flattenFormListPicker(pageComponent.ListPicker)
		}
		if pageComponent.DatePicker != nil {
			pageComponentMap["date_picker"] = flattenFormDatePicker(pageComponent.DatePicker)
		}
		if pageComponent.Input != nil {
			pageComponentMap["input"] = flattenFormInput(pageComponent.Input)
		}
		if pageComponent.WheelPicker != nil {
			pageComponentMap["wheel_picker"] = flattenFormWheelPicker(pageComponent.WheelPicker)
		}

		pageComponentList = append(pageComponentList, pageComponentMap)
	}
	return pageComponentList
}

// flattenFormListPicker maps a Genesys Cloud *platformclientv2.Formlistpicker into a nested block
func flattenFormListPicker(listPicker *platformclientv2.Formlistpicker) []interface{} {
	if listPicker == nil {
		return nil
	}

	listPickerMap := make(map[string]interface{})
	if listPicker.Sections != nil {
		listPickerMap["sections"] = flattenFormListPickerSections(listPicker.Sections)
	}

	return []interface{}{listPickerMap}
}

// flattenFormListPickerSections maps a Genesys Cloud *[]platformclientv2.Formlistpickersection into the `sections` nested blocks
func flattenFormListPickerSections(sections *[]platformclientv2.Formlistpickersection) []interface{} {
	if sections == nil {
		return nil
	}

	sectionList := make([]interface{}, 0, len(*sections))
	for _, section := range *sections {
		sectionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(sectionMap, "title", section.Title)
		resourcedata.SetMapValueIfNotNil(sectionMap, "multiple_selection", section.MultipleSelection)
		if section.Items != nil {
			sectionMap["items"] = flattenFormListPickerItems(section.Items)
		}

		sectionList = append(sectionList, sectionMap)
	}
	return sectionList
}

// flattenFormListPickerItems maps a Genesys Cloud *[]platformclientv2.Formlistpickeritem into the `items` nested blocks
func flattenFormListPickerItems(items *[]platformclientv2.Formlistpickeritem) []interface{} {
	if items == nil {
		return nil
	}

	itemList := make([]interface{}, 0, len(*items))
	for _, item := range *items {
		itemMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(itemMap, "title", item.Title)
		resourcedata.SetMapValueIfNotNil(itemMap, "image_url", item.ImageUrl)

		itemList = append(itemList, itemMap)
	}
	return itemList
}

// flattenFormDatePicker maps a Genesys Cloud *platformclientv2.Formdatepicker into a nested block
func flattenFormDatePicker(datePicker *platformclientv2.Formdatepicker) []interface{} {
	if datePicker == nil {
		return nil
	}

	datePickerMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(datePickerMap, "title", datePicker.Title)
	resourcedata.SetMapValueIfNotNil(datePickerMap, "subtitle", datePicker.Subtitle)
	resourcedata.SetMapValueIfNotNil(datePickerMap, "date_display_format", datePicker.DateDisplayFormat)

	return []interface{}{datePickerMap}
}

// flattenFormInput maps a Genesys Cloud *platformclientv2.Input into a nested block
func flattenFormInput(input *platformclientv2.Input) []interface{} {
	if input == nil {
		return nil
	}

	inputMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(inputMap, "title", input.Title)
	resourcedata.SetMapValueIfNotNil(inputMap, "subtitle", input.Subtitle)
	resourcedata.SetMapValueIfNotNil(inputMap, "placeholder_text", input.PlaceholderText)
	resourcedata.SetMapValueIfNotNil(inputMap, "keyboard_type", input.KeyboardType)
	resourcedata.SetMapValueIfNotNil(inputMap, "auto_complete_type", input.AutoCompleteType)
	resourcedata.SetMapValueIfNotNil(inputMap, "is_multiple_line", input.IsMultipleLine)
	resourcedata.SetMapValueIfNotNil(inputMap, "is_required", input.IsRequired)

	return []interface{}{inputMap}
}

// flattenFormWheelPicker maps a Genesys Cloud *platformclientv2.Wheelpicker into a nested block
func flattenFormWheelPicker(wheelPicker *platformclientv2.Wheelpicker) []interface{} {
	if wheelPicker == nil {
		return nil
	}

	wheelPickerMap := make(map[string]interface{})
	if wheelPicker.Items != nil {
		wheelPickerMap["items"] = flattenFormWheelPickerItems(wheelPicker.Items)
	}

	return []interface{}{wheelPickerMap}
}

// flattenFormWheelPickerItems maps a Genesys Cloud *[]platformclientv2.Wheelpickeritem into the `items` nested blocks
func flattenFormWheelPickerItems(items *[]platformclientv2.Wheelpickeritem) []interface{} {
	if items == nil {
		return nil
	}

	itemList := make([]interface{}, 0, len(*items))
	for _, item := range *items {
		itemMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(itemMap, "title", item.Title)
		resourcedata.SetMapValueIfNotNil(itemMap, "value", item.Value)

		itemList = append(itemList, itemMap)
	}
	return itemList
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

func flattenAddressableEntityRefs(addressableEntityRefs *[]platformclientv2.Rmsassetaddressableref) *schema.Set {
	addressableEntityRefList := make([]interface{}, len(*addressableEntityRefs))
	for i, v := range *addressableEntityRefs {
		addressableEntityRefList[i] = *v.Id
	}
	return schema.NewSet(schema.HashString, addressableEntityRefList)
}

func GenerateResponseManagementResponseResource(
	resourceLabel string,
	name string,
	libraryIds []string,
	interactionType string,
	schema string,
	responseType string,
	assetIds []string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_responsemanagement_response" "%s" {
			name = "%s"
			library_ids = [%s]
			interaction_type = %s
			substitutions_schema_id = %s
			response_type = %s
			asset_ids = [%s]
			%s
		}
	`, resourceLabel, name, strings.Join(libraryIds, ", "), interactionType, schema, responseType, strings.Join(assetIds, ", "), strings.Join(nestedBlocks, "\n"))
}

func GenerateTextsBlock(content string, contentType string, cType string) string {
	return fmt.Sprintf(`
		texts {
			content = "%s"
			content_type = "%s"
			type = %s
		}
	`, content, contentType, cType)
}
