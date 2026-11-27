package responsemanagement_response

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

func testCarouselWhatsappConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":           "summer_sale_carousel",
		"namespace":      "",
		"language":       "en_US",
		"integration_id": "11111111-2222-3333-4444-555555555555",
		"category":       "marketing",
		"carousel": []interface{}{
			map[string]interface{}{
				"cards": []interface{}{
					map[string]interface{}{
						"header": []interface{}{
							map[string]interface{}{"type": "Image", "content": "https://example.com/card1.png"},
						},
						"buttons": []interface{}{
							map[string]interface{}{"type": "QuickReply", "content": "", "content_text": "Buy now", "payload": "buy-1"},
							map[string]interface{}{"type": "Url", "content": "https://example.com/1", "content_text": "Details", "payload": ""},
						},
						"card_body_text": []interface{}{
							map[string]interface{}{"content": "First card", "content_type": "text/plain"},
						},
					},
					map[string]interface{}{
						"header": []interface{}{
							map[string]interface{}{"type": "Image", "content": "https://example.com/card2.png"},
						},
						"buttons": []interface{}{
							map[string]interface{}{"type": "QuickReply", "content": "", "content_text": "Buy now", "payload": "buy-2"},
						},
						"card_body_text": []interface{}{},
					},
				},
			},
		},
	}
}

func TestUnitBuildWhatsappDefinitionCarousel(t *testing.T) {
	whatsApp := []interface{}{testCarouselWhatsappConfig()}

	got := buildWhatsappDefinition(whatsApp)

	if *got.IntegrationId != "11111111-2222-3333-4444-555555555555" || *got.Category != "marketing" {
		t.Fatalf("integration_id/category not built: %+v", got)
	}
	if got.TemplateStatus != nil || got.StatusInfo != nil {
		t.Fatalf("read-only fields must not be sent, got templateStatus=%v statusInfo=%v", got.TemplateStatus, got.StatusInfo)
	}
	if got.Carousel == nil || len(*got.Carousel.Cards) != 2 {
		t.Fatalf("expected 2 cards, got %+v", got.Carousel)
	}

	first := (*got.Carousel.Cards)[0]
	if *first.Header.VarType != "Image" || *first.Header.Content != "https://example.com/card1.png" {
		t.Errorf("card 1 header wrong: %+v", first.Header)
	}
	if len(*first.Buttons) != 2 || *(*first.Buttons)[0].Payload != "buy-1" || *(*first.Buttons)[1].Content != "https://example.com/1" {
		t.Errorf("card 1 buttons wrong: %+v", *first.Buttons)
	}
	// Empty strings are omitted so the API never receives "" for unset button fields.
	if (*first.Buttons)[0].Content != nil || (*first.Buttons)[1].Payload != nil {
		t.Errorf("empty button fields should be nil: %+v", *first.Buttons)
	}
	if *first.CardBodyText.Content != "First card" || *first.CardBodyText.ContentType != "text/plain" {
		t.Errorf("card 1 body text wrong: %+v", first.CardBodyText)
	}

	second := (*got.Carousel.Cards)[1]
	if second.CardBodyText != nil {
		t.Errorf("card 2 body text should be nil when not configured, got %+v", second.CardBodyText)
	}
	if *(*second.Buttons)[0].Payload != "buy-2" {
		t.Errorf("card order not preserved: %+v", *second.Buttons)
	}
}

func TestUnitBuildWhatsappDefinitionWithoutCarousel(t *testing.T) {
	whatsApp := []interface{}{map[string]interface{}{
		"name":           "plain_template",
		"namespace":      "ns",
		"language":       "en_US",
		"integration_id": "",
		"category":       "",
		"carousel":       []interface{}{},
	}}

	got := buildWhatsappDefinition(whatsApp)

	if got.IntegrationId != nil || got.Category != nil || got.Carousel != nil {
		t.Fatalf("non-carousel template must keep the existing payload, got %+v", got)
	}
	if *got.Name != "plain_template" || *got.Namespace != "ns" || *got.Language != "en_US" {
		t.Fatalf("existing fields not built: %+v", got)
	}
}

func TestUnitFlattenWhatsappDefinitionCarouselRoundTrip(t *testing.T) {
	sdk := platformclientv2.Whatsappdefinition{
		Name:           platformclientv2.String("summer_sale_carousel"),
		Language:       platformclientv2.String("en_US"),
		IntegrationId:  platformclientv2.String("integration-1"),
		Category:       platformclientv2.String("MARKETING"),
		TemplateStatus: platformclientv2.String("APPROVED"),
		StatusInfo: &platformclientv2.Statusinfo{
			Code:    platformclientv2.String("OK"),
			Message: platformclientv2.String("Template approved"),
		},
		Carousel: &platformclientv2.Carousel{Cards: &[]platformclientv2.Carouselcard{
			{
				Header: &platformclientv2.Messageheader{VarType: platformclientv2.String("Image"), Content: platformclientv2.String("https://example.com/1.png")},
				Buttons: &[]platformclientv2.Button{
					{VarType: platformclientv2.String("QuickReply"), ContentText: platformclientv2.String("Buy"), Payload: platformclientv2.String("p1")},
				},
				CardBodyText: &platformclientv2.Cardbodytext{Content: platformclientv2.String("Body"), ContentType: platformclientv2.String("text/plain")},
			},
			{
				Header:  &platformclientv2.Messageheader{VarType: platformclientv2.String("Image"), Content: platformclientv2.String("https://example.com/2.png")},
				Buttons: &[]platformclientv2.Button{{VarType: platformclientv2.String("Url"), Content: platformclientv2.String("https://example.com"), ContentText: platformclientv2.String("Open")}},
			},
		}},
	}

	flattened := flattenWhatsappDefinition(&sdk)[0].(map[string]interface{})

	if flattened["integration_id"] != "integration-1" || flattened["category"] != "marketing" {
		t.Errorf("integration_id/category not flattened: %+v", flattened)
	}
	if flattened["template_status"] != "APPROVED" {
		t.Errorf("template_status not flattened: %v", flattened["template_status"])
	}
	wantStatus := []interface{}{map[string]interface{}{"code": "OK", "message": "Template approved"}}
	if !reflect.DeepEqual(flattened["status_info"], wantStatus) {
		t.Errorf("status_info = %+v, want %+v", flattened["status_info"], wantStatus)
	}

	cards := flattened["carousel"].([]interface{})[0].(map[string]interface{})["cards"].([]interface{})
	if len(cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(cards))
	}
	second := cards[1].(map[string]interface{})
	if _, ok := second["card_body_text"]; ok {
		t.Errorf("card_body_text should be absent when the API omits it: %+v", second)
	}

	// Store the flattened value in real resource data, read it back the way the provider does on update,
	// and compare with the original payload (minus read-only fields).
	d := schema.TestResourceDataRaw(t, ResourceResponsemanagementResponse().Schema, map[string]interface{}{})
	if err := d.Set("messaging_template", flattenMessagingTemplate(&platformclientv2.Messagingtemplate{WhatsApp: &sdk})); err != nil {
		t.Fatalf("setting messaging_template: %v", err)
	}
	rebuilt := buildMessagingTemplate(d.Get("messaging_template").([]interface{})).WhatsApp
	sdk.TemplateStatus, sdk.StatusInfo = nil, nil
	if !reflect.DeepEqual(rebuilt.Carousel, sdk.Carousel) {
		t.Errorf("carousel round trip mismatch:\n got  %+v\n want %+v", *rebuilt.Carousel.Cards, *sdk.Carousel.Cards)
	}
	if *rebuilt.IntegrationId != *sdk.IntegrationId || *rebuilt.Category != "marketing" {
		t.Errorf("round trip lost integration_id/category: %+v", rebuilt)
	}
}

func TestUnitFlattenWhatsappDefinitionWithoutCarousel(t *testing.T) {
	sdk := platformclientv2.Whatsappdefinition{
		Name:     platformclientv2.String("plain_template"),
		Language: platformclientv2.String("en_US"),
	}

	flattened := flattenWhatsappDefinition(&sdk)[0].(map[string]interface{})

	for _, key := range []string{"integration_id", "category", "carousel", "template_status", "status_info"} {
		if _, ok := flattened[key]; ok {
			t.Errorf("%s should be absent for a non-carousel template, got %v", key, flattened[key])
		}
	}
}

func TestUnitFlattenCarouselWithoutCards(t *testing.T) {
	if got := flattenCarousel(&platformclientv2.Carousel{}); got != nil {
		t.Errorf("expected nil for a carousel without cards, got %+v", got)
	}
}

func TestUnitShouldReplaceCarouselTemplate(t *testing.T) {
	carouselState := func(status string) map[string]interface{} {
		m := testCarouselWhatsappConfig()
		m["template_status"] = status
		return m
	}
	withChange := func(key string, value interface{}) map[string]interface{} {
		m := testCarouselWhatsappConfig()
		m[key] = value
		return m
	}
	otherCarousel := []interface{}{map[string]interface{}{"cards": []interface{}{}}}

	tests := []struct {
		name         string
		old          map[string]interface{}
		new          map[string]interface{}
		textsChanged bool
		want         bool
	}{
		{"unchanged carousel stays in place", carouselState("APPROVED"), testCarouselWhatsappConfig(), false, false},
		{"name change replaces", carouselState("APPROVED"), withChange("name", "renamed"), false, true},
		{"language change replaces", carouselState("PENDING"), withChange("language", "pl_PL"), false, true},
		{"integration change replaces", carouselState("ACCEPTED"), withChange("integration_id", "other"), false, true},
		{"category change replaces", carouselState("REJECTED"), withChange("category", ""), false, true},
		{"card change replaces", carouselState("APPROVED"), withChange("carousel", otherCarousel), false, true},
		{"carousel removal replaces", carouselState("APPROVED"), withChange("carousel", []interface{}{}), false, true},
		{"whats_app removal replaces", carouselState("APPROVED"), nil, false, true},
		{"texts change replaces", carouselState("APPROVED"), testCarouselWhatsappConfig(), true, true},
		{"ERROR status allows in-place edit", carouselState("ERROR"), withChange("name", "renamed"), true, false},
		{"plain template is never replaced", map[string]interface{}{"name": "a", "language": "en_US", "carousel": []interface{}{}}, map[string]interface{}{"name": "b", "language": "en_US", "carousel": []interface{}{}}, true, false},
		{"no previous whats_app", nil, testCarouselWhatsappConfig(), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldReplaceCarouselTemplate(tt.old, tt.new, tt.textsChanged); got != tt.want {
				t.Errorf("shouldReplaceCarouselTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnitFirstWhatsappDefinition(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, ResourceResponsemanagementResponse().Schema, map[string]interface{}{})
	if got := firstWhatsappDefinition(resourceData.Get("messaging_template")); got != nil {
		t.Errorf("expected nil without messaging_template, got %+v", got)
	}

	sdk := &platformclientv2.Messagingtemplate{WhatsApp: &platformclientv2.Whatsappdefinition{
		Name:     platformclientv2.String("tpl"),
		Language: platformclientv2.String("en_US"),
	}}
	if err := resourceData.Set("messaging_template", flattenMessagingTemplate(sdk)); err != nil {
		t.Fatalf("setting messaging_template: %v", err)
	}
	got := firstWhatsappDefinition(resourceData.Get("messaging_template"))
	if got == nil || got["name"] != "tpl" {
		t.Errorf("expected whats_app map with name tpl, got %+v", got)
	}

	if got := firstWhatsappDefinition("not a set"); got != nil {
		t.Errorf("expected nil for an unexpected type, got %+v", got)
	}
}

func TestUnitCarouselReplacementKeys(t *testing.T) {
	messagingTemplateKeys := []string{
		"messaging_template.0.whats_app.0.carousel.0.cards.0.buttons.1.content_text",
		"messaging_template.0.whats_app.0.name",
		"messaging_template.0.whats_app.0.template_status",
		"messaging_template.0.whats_app.0.status_info.0.code",
		"messaging_template.0.whats_app.0.namespace",
		"messaging_template.0.whats_app.#",
	}
	textsKeys := []string{"texts.1234.content"}

	got := carouselReplacementKeys(messagingTemplateKeys, textsKeys)

	want := []string{
		"messaging_template.0.whats_app.#",
		"messaging_template.0.whats_app.0.carousel.0.cards.0.buttons.1.content_text",
		"messaging_template.0.whats_app.0.name",
		"texts.1234.content",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("carouselReplacementKeys() = %v, want %v", got, want)
	}
}

func TestUnitResponsemanagementResponseExporterWhatsappIntegrationRef(t *testing.T) {
	refAttr, ok := ResponsemanagementResponseExporter().RefAttrs["messaging_template.whats_app.integration_id"]
	if !ok {
		t.Fatal("expected messaging_template.whats_app.integration_id to be an exporter reference")
	}
	if want := "genesyscloud_conversations_messaging_integrations_whatsapp"; refAttr.RefType != want {
		t.Errorf("RefType = %q, want %q", refAttr.RefType, want)
	}
}
