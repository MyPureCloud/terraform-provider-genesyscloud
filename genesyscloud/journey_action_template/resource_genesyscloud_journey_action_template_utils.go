package journey_action_template

import (
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"terraform-provider-genesyscloud/genesyscloud/util/stringmap"
	"terraform-provider-genesyscloud/genesyscloud/util/typeconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

// All buildSdkPatch*  functions are helper method which maps Create operation of journeyApi's Actiontemplates
func buildSdkPatchActionTemplate(patchActionTemplate *schema.ResourceData) *platformclientv2.Patchactiontemplate {
	name := patchActionTemplate.Get("name").(string)
	description := patchActionTemplate.Get("description").(string)
	mediaType := patchActionTemplate.Get("media_type").(string)
	state := patchActionTemplate.Get("state").(string)
	contentOffer := resourcedata.BuildSdkListFirstElement(patchActionTemplate, "content_offer", buildSdkPatchContentOffer, true)

	sdkPatchActionTemplate := platformclientv2.Patchactiontemplate{}
	sdkPatchActionTemplate.SetField("Name", &name)
	sdkPatchActionTemplate.SetField("Description", &description)
	sdkPatchActionTemplate.SetField("MediaType", &mediaType)
	sdkPatchActionTemplate.SetField("State", &state)
	sdkPatchActionTemplate.SetField("ContentOffer", contentOffer)
	return &sdkPatchActionTemplate
}

func buildSdkPatchContentOffer(patchContentOffer map[string]interface{}) *platformclientv2.Patchcontentoffer {
	imageUrl := patchContentOffer["image_url"].(string)
	displayMode := patchContentOffer["display_mode"].(string)
	layoutMode := patchContentOffer["layout_mode"].(string)
	title := patchContentOffer["title"].(string)
	headline := patchContentOffer["headline"].(string)
	body := patchContentOffer["body"].(string)
	callToAction := stringmap.BuildSdkListFirstElement(patchContentOffer, "call_to_action", buildSdkPatchCallToAction, true)
	style := stringmap.BuildSdkListFirstElement(patchContentOffer, "style", buildSdkPatchContentOfferStylingConfiguration, true)

	sdkPatchActionTemplate := platformclientv2.Patchcontentoffer{}
	sdkPatchActionTemplate.SetField("ImageUrl", &imageUrl)
	sdkPatchActionTemplate.SetField("DisplayMode", &displayMode)
	sdkPatchActionTemplate.SetField("LayoutMode", &layoutMode)
	sdkPatchActionTemplate.SetField("Title", &title)
	sdkPatchActionTemplate.SetField("Headline", &headline)
	sdkPatchActionTemplate.SetField("Body", &body)
	sdkPatchActionTemplate.SetField("CallToAction", callToAction)
	sdkPatchActionTemplate.SetField("Style", style)
	return &sdkPatchActionTemplate
}

func buildSdkPatchContentOfferStylingConfiguration(patchContentOfferStyle map[string]interface{}) *platformclientv2.Patchcontentofferstylingconfiguration {
	position := stringmap.BuildSdkListFirstElement(patchContentOfferStyle, "position", buildSdkPatchPosition, true)
	offer := stringmap.BuildSdkListFirstElement(patchContentOfferStyle, "offer", buildSdkPatchOffer, true)
	closeButton := stringmap.BuildSdkListFirstElement(patchContentOfferStyle, "close_button", buildSdkPatchCloseButton, true)
	ctaButton := stringmap.BuildSdkListFirstElement(patchContentOfferStyle, "cta_button", buildSdkPatchCtaButton, true)
	title := stringmap.BuildSdkListFirstElement(patchContentOfferStyle, "title", buildSdkPatchTitleOrHeadlineOrBody, true)
	headline := stringmap.BuildSdkListFirstElement(patchContentOfferStyle, "headline", buildSdkPatchTitleOrHeadlineOrBody, true)
	body := stringmap.BuildSdkListFirstElement(patchContentOfferStyle, "body", buildSdkPatchTitleOrHeadlineOrBody, true)

	sdkPatchContentOfferStyle := platformclientv2.Patchcontentofferstylingconfiguration{}
	sdkPatchContentOfferStyle.SetField("Position", position)
	sdkPatchContentOfferStyle.SetField("Offer", offer)
	sdkPatchContentOfferStyle.SetField("CloseButton", closeButton)
	sdkPatchContentOfferStyle.SetField("CtaButton", ctaButton)
	sdkPatchContentOfferStyle.SetField("Title", title)
	sdkPatchContentOfferStyle.SetField("Headline", headline)
	sdkPatchContentOfferStyle.SetField("Body", body)
	return &sdkPatchContentOfferStyle
}

func buildSdkPatchPosition(patchContentPositionProp map[string]interface{}) *platformclientv2.Patchcontentpositionproperties {
	top := patchContentPositionProp["top"].(string)
	bottom := patchContentPositionProp["bottom"].(string)
	left := patchContentPositionProp["left"].(string)
	right := patchContentPositionProp["right"].(string)

	sdkPatchContentPositionProp := &platformclientv2.Patchcontentpositionproperties{}
	sdkPatchContentPositionProp.SetField("Top", &top)
	sdkPatchContentPositionProp.SetField("Bottom", &bottom)
	sdkPatchContentPositionProp.SetField("Left", &left)
	sdkPatchContentPositionProp.SetField("Right", &right)
	return sdkPatchContentPositionProp
}

func buildSdkPatchOffer(patchContentOfferStyleProp map[string]interface{}) *platformclientv2.Patchcontentofferstyleproperties {
	padding := patchContentOfferStyleProp["padding"].(string)
	color := patchContentOfferStyleProp["color"].(string)
	backgroundColor := patchContentOfferStyleProp["background_color"].(string)

	sdkPatchContentOfferStyleProp := &platformclientv2.Patchcontentofferstyleproperties{}
	sdkPatchContentOfferStyleProp.SetField("Padding", &padding)
	sdkPatchContentOfferStyleProp.SetField("Color", &color)
	sdkPatchContentOfferStyleProp.SetField("BackgroundColor", &backgroundColor)
	return sdkPatchContentOfferStyleProp
}

func buildSdkPatchCloseButton(patchCloseButtonStyleProp map[string]interface{}) *platformclientv2.Patchclosebuttonstyleproperties {
	color := patchCloseButtonStyleProp["color"].(string)
	opacity64 := patchCloseButtonStyleProp["opacity"].(float64)
	opacity := typeconv.Float64to32(&opacity64)

	skdPatchCloseButtonStyleProp := &platformclientv2.Patchclosebuttonstyleproperties{}
	skdPatchCloseButtonStyleProp.SetField("Color", &color)
	skdPatchCloseButtonStyleProp.SetField("Opacity", opacity)
	return skdPatchCloseButtonStyleProp
}

func buildSdkPatchCtaButton(patchCtaButtonStyleProp map[string]interface{}) *platformclientv2.Patchctabuttonstyleproperties {
	color := patchCtaButtonStyleProp["color"].(string)
	font := patchCtaButtonStyleProp["font"].(string)
	fontSize := patchCtaButtonStyleProp["font_size"].(string)
	textAlign := patchCtaButtonStyleProp["text_align"].(string)
	backgroundColor := patchCtaButtonStyleProp["background_color"].(string)

	sdkPatchCtaButtonStyleProp := &platformclientv2.Patchctabuttonstyleproperties{}
	sdkPatchCtaButtonStyleProp.SetField("Color", &color)
	sdkPatchCtaButtonStyleProp.SetField("Font", &font)
	sdkPatchCtaButtonStyleProp.SetField("FontSize", &fontSize)
	sdkPatchCtaButtonStyleProp.SetField("TextAlign", &textAlign)
	sdkPatchCtaButtonStyleProp.SetField("BackgroundColor", &backgroundColor)
	return sdkPatchCtaButtonStyleProp
}

func buildSdkPatchTitleOrHeadlineOrBody(patchTextStyleProp map[string]interface{}) *platformclientv2.Patchtextstyleproperties {
	color := patchTextStyleProp["color"].(string)
	font := patchTextStyleProp["font"].(string)
	fontSize := patchTextStyleProp["font_size"].(string)
	textAlign := patchTextStyleProp["text_align"].(string)

	sdkPatchTextStyleProp := &platformclientv2.Patchtextstyleproperties{}
	sdkPatchTextStyleProp.SetField("Color", &color)
	sdkPatchTextStyleProp.SetField("Font", &font)
	sdkPatchTextStyleProp.SetField("FontSize", &fontSize)
	sdkPatchTextStyleProp.SetField("TextAlign", &textAlign)
	return sdkPatchTextStyleProp
}

func buildSdkPatchCallToAction(patchCallToAction map[string]interface{}) *platformclientv2.Patchcalltoaction {
	text := patchCallToAction["text"].(string)
	url := patchCallToAction["url"].(string)
	targetUrl := patchCallToAction["target"].(string)

	sdkPatchCallToAction := &platformclientv2.Patchcalltoaction{}
	sdkPatchCallToAction.SetField("Text", &text)
	sdkPatchCallToAction.SetField("Url", &url)
	sdkPatchCallToAction.SetField("Target", &targetUrl)
	return sdkPatchCallToAction
}

// All buildSdk* (not buildSdkPatch*) functions are helper method which maps Create operation of journeyApi's Actiontemplates
func buildSdkActionTemplate(actionTemplate *schema.ResourceData) *platformclientv2.Actiontemplate {
	name := actionTemplate.Get("name").(string)
	description := actionTemplate.Get("description").(string)
	mediaType := actionTemplate.Get("media_type").(string)
	state := actionTemplate.Get("state").(string)
	contentOffer := resourcedata.BuildSdkListFirstElement(actionTemplate, "content_offer", buildSdkContentOffer, true)

	return &platformclientv2.Actiontemplate{
		Name:         &name,
		Description:  &description,
		MediaType:    &mediaType,
		State:        &state,
		ContentOffer: contentOffer,
	}
}

func buildSdkContentOffer(contentOffer map[string]interface{}) *platformclientv2.Contentoffer {
	imageUrl := contentOffer["image_url"].(string)
	displayMode := contentOffer["display_mode"].(string)
	layoutMode := contentOffer["layout_mode"].(string)
	title := contentOffer["title"].(string)
	headline := contentOffer["headline"].(string)
	body := contentOffer["body"].(string)
	callToAction := stringmap.BuildSdkListFirstElement(contentOffer, "call_to_action", buildSdkCallToAction, true)
	style := stringmap.BuildSdkListFirstElement(contentOffer, "style", buildSdkContentOfferStylingConfiguration, true)

	return &platformclientv2.Contentoffer{
		ImageUrl:     &imageUrl,
		DisplayMode:  &displayMode,
		LayoutMode:   &layoutMode,
		Title:        &title,
		Headline:     &headline,
		Body:         &body,
		CallToAction: callToAction,
		Style:        style,
	}
}

func buildSdkCallToAction(callToAction map[string]interface{}) *platformclientv2.Calltoaction {
	text := callToAction["text"].(string)
	url := callToAction["url"].(string)
	targetUrl := callToAction["target"].(string)

	return &platformclientv2.Calltoaction{
		Text:   &text,
		Url:    &url,
		Target: &targetUrl,
	}
}

func buildSdkContentOfferStylingConfiguration(contentOfferStylingConfig map[string]interface{}) *platformclientv2.Contentofferstylingconfiguration {
	position := stringmap.BuildSdkListFirstElement(contentOfferStylingConfig, "position", buildSdkContentPositionProperties, true)
	offer := stringmap.BuildSdkListFirstElement(contentOfferStylingConfig, "offer", buildSdkContentOfferStyleProperties, true)
	closeButton := stringmap.BuildSdkListFirstElement(contentOfferStylingConfig, "close_button", buildSdkCloseButtonStyleProperties, true)
	ctaButton := stringmap.BuildSdkListFirstElement(contentOfferStylingConfig, "cta_button", buildSdkCtaButtonStyleProperties, true)
	title := stringmap.BuildSdkListFirstElement(contentOfferStylingConfig, "title", buildSdkTextStyleProperties, true)
	headline := stringmap.BuildSdkListFirstElement(contentOfferStylingConfig, "headline", buildSdkTextStyleProperties, true)
	body := stringmap.BuildSdkListFirstElement(contentOfferStylingConfig, "body", buildSdkTextStyleProperties, true)

	return &platformclientv2.Contentofferstylingconfiguration{
		Position:    position,
		Offer:       offer,
		CloseButton: closeButton,
		CtaButton:   ctaButton,
		Title:       title,
		Headline:    headline,
		Body:        body,
	}
}

func buildSdkContentPositionProperties(contentPositionProperties map[string]interface{}) *platformclientv2.Contentpositionproperties {
	top := contentPositionProperties["top"].(string)
	bottom := contentPositionProperties["bottom"].(string)
	left := contentPositionProperties["left"].(string)
	right := contentPositionProperties["right"].(string)
	return &platformclientv2.Contentpositionproperties{
		Top:    &top,
		Bottom: &bottom,
		Left:   &left,
		Right:  &right,
	}
}

func buildSdkContentOfferStyleProperties(contentPositionProperties map[string]interface{}) *platformclientv2.Contentofferstyleproperties {
	padding := contentPositionProperties["padding"].(string)
	color := contentPositionProperties["color"].(string)
	backGroundColor := contentPositionProperties["background_color"].(string)
	return &platformclientv2.Contentofferstyleproperties{
		Padding:         &padding,
		Color:           &color,
		BackgroundColor: &backGroundColor,
	}
}

func buildSdkCtaButtonStyleProperties(contentPositionProperties map[string]interface{}) *platformclientv2.Ctabuttonstyleproperties {
	color := contentPositionProperties["color"].(string)
	font := contentPositionProperties["font"].(string)
	fontSize := contentPositionProperties["font_size"].(string)
	textAlign := contentPositionProperties["text_align"].(string)
	backgoundColor := contentPositionProperties["background_color"].(string)
	return &platformclientv2.Ctabuttonstyleproperties{
		Color:           &color,
		Font:            &font,
		FontSize:        &fontSize,
		TextAlign:       &textAlign,
		BackgroundColor: &backgoundColor,
	}
}

func buildSdkCloseButtonStyleProperties(contentPositionProperties map[string]interface{}) *platformclientv2.Closebuttonstyleproperties {
	color := contentPositionProperties["color"].(string)
	opacity64 := contentPositionProperties["opacity"].(float64)
	opacity := typeconv.Float64to32(&opacity64)
	return &platformclientv2.Closebuttonstyleproperties{
		Color:   &color,
		Opacity: opacity,
	}
}

func buildSdkTextStyleProperties(contentPositionProperties map[string]interface{}) *platformclientv2.Textstyleproperties {
	color := contentPositionProperties["color"].(string)
	font := contentPositionProperties["font"].(string)
	fontSize := contentPositionProperties["font_size"].(string)
	textAlign := contentPositionProperties["text_align"].(string)
	return &platformclientv2.Textstyleproperties{
		Color:     &color,
		Font:      &font,
		FontSize:  &fontSize,
		TextAlign: &textAlign,
	}
}

// All flatten* functions are helper method which maps Read operation of journeyApi's Actiontemplates
func flattenActionTemplate(data *schema.ResourceData, actionTemplate *platformclientv2.Actiontemplate) {
	data.Set("name", *actionTemplate.Name)
	resourcedata.SetNillableValue(data, "description", actionTemplate.Description)
	data.Set("media_type", *actionTemplate.MediaType)
	data.Set("state", *actionTemplate.State)
	resourcedata.SetNillableValue(data, "content_offer", lists.FlattenAsList(actionTemplate.ContentOffer, flattenActionTemplateContentOffer))
}

func flattenActionTemplateContentOffer(resource *platformclientv2.Contentoffer) map[string]interface{} {
	actionTemplateContentOfferMap := make(map[string]interface{})
	actionTemplateContentOfferMap["image_url"] = resource.ImageUrl
	actionTemplateContentOfferMap["display_mode"] = resource.DisplayMode
	actionTemplateContentOfferMap["layout_mode"] = resource.LayoutMode
	actionTemplateContentOfferMap["title"] = resource.Title
	actionTemplateContentOfferMap["headline"] = resource.Headline
	actionTemplateContentOfferMap["body"] = resource.Body
	stringmap.SetValueIfNotNil(actionTemplateContentOfferMap, "call_to_action", lists.FlattenAsList(resource.CallToAction, flattenCallToAction))
	stringmap.SetValueIfNotNil(actionTemplateContentOfferMap, "style", lists.FlattenAsList(resource.Style, flattenStyle))
	return actionTemplateContentOfferMap
}

func flattenCallToAction(resource *platformclientv2.Calltoaction) map[string]interface{} {
	callToActionMap := make(map[string]interface{})
	callToActionMap["text"] = resource.Text
	callToActionMap["url"] = resource.Url
	callToActionMap["target"] = resource.Target
	return callToActionMap
}

func flattenStyle(resource *platformclientv2.Contentofferstylingconfiguration) map[string]interface{} {
	styleMap := make(map[string]interface{})
	stringmap.SetValueIfNotNil(styleMap, "position", lists.FlattenAsList(resource.Position, flattenPositionProperties))
	stringmap.SetValueIfNotNil(styleMap, "offer", lists.FlattenAsList(resource.Offer, flattenOfferProperties))
	stringmap.SetValueIfNotNil(styleMap, "close_button", lists.FlattenAsList(resource.CloseButton, flattenCloseButtonProperties))
	stringmap.SetValueIfNotNil(styleMap, "cta_button", lists.FlattenAsList(resource.CtaButton, flattenCtaButtonProperties))
	stringmap.SetValueIfNotNil(styleMap, "title", lists.FlattenAsList(resource.Title, flattenTextStyleProperties))
	stringmap.SetValueIfNotNil(styleMap, "headline", lists.FlattenAsList(resource.Headline, flattenTextStyleProperties))
	stringmap.SetValueIfNotNil(styleMap, "body", lists.FlattenAsList(resource.Body, flattenTextStyleProperties))
	return styleMap
}

func flattenPositionProperties(resource *platformclientv2.Contentpositionproperties) map[string]interface{} {
	positionMap := make(map[string]interface{})
	positionMap["top"] = resource.Top
	positionMap["bottom"] = resource.Bottom
	positionMap["left"] = resource.Left
	positionMap["right"] = resource.Right
	return positionMap
}

func flattenOfferProperties(resource *platformclientv2.Contentofferstyleproperties) map[string]interface{} {
	offerMap := make(map[string]interface{})
	offerMap["padding"] = resource.Padding
	offerMap["color"] = resource.Color
	offerMap["background_color"] = resource.BackgroundColor
	return offerMap
}

func flattenCloseButtonProperties(resource *platformclientv2.Closebuttonstyleproperties) map[string]interface{} {
	closeButtonMap := make(map[string]interface{})
	closeButtonMap["color"] = resource.Color
	closeButtonMap["opacity"] = *typeconv.Float32to64(resource.Opacity)
	return closeButtonMap
}

func flattenCtaButtonProperties(resource *platformclientv2.Ctabuttonstyleproperties) map[string]interface{} {
	ctaButtonMap := make(map[string]interface{})
	ctaButtonMap["color"] = resource.Color
	ctaButtonMap["font"] = resource.Font
	ctaButtonMap["font_size"] = resource.FontSize
	ctaButtonMap["text_align"] = resource.TextAlign
	ctaButtonMap["background_color"] = resource.BackgroundColor
	return ctaButtonMap
}

func flattenTextStyleProperties(resource *platformclientv2.Textstyleproperties) map[string]interface{} {
	textStyleMap := make(map[string]interface{})
	textStyleMap["color"] = resource.Color
	textStyleMap["font"] = resource.Font
	textStyleMap["font_size"] = resource.FontSize
	textStyleMap["text_align"] = resource.TextAlign
	return textStyleMap
}
