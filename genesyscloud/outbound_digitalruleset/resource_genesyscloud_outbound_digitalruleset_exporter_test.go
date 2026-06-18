package outbound_digitalruleset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v191/platformclientv2"
	respManagement "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/responsemanagement_response"
)

func TestOutboundDigitalrulesetExporter_ContentTemplateRefAttrs(t *testing.T) {
	exporter := OutboundDigitalrulesetExporter()
	if exporter == nil || exporter.RefAttrs == nil {
		t.Fatal("expected OutboundDigitalrulesetExporter RefAttrs to be defined")
	}

	for _, path := range []string{
		"rules.actions.set_content_template_action_settings.sms_content_template_id",
		"rules.actions.set_content_template_action_settings.email_content_template_id",
	} {
		settings := exporter.GetRefAttrSettings(path)
		if settings == nil {
			t.Fatalf("expected RefAttrs for %q", path)
		}
		if settings.RefType != respManagement.ResourceType {
			t.Fatalf("expected %q RefType %q, got %q", path, respManagement.ResourceType, settings.RefType)
		}
	}
}

func TestOutboundDigitalrulesetExporter_UnresolvedReferenceResolvers(t *testing.T) {
	exporter := OutboundDigitalrulesetExporter()
	if exporter.CustomAttributeResolver == nil {
		t.Fatal("expected CustomAttributeResolver to be defined")
	}

	for _, path := range []string{
		"contact_list_id",
		"rules.actions.set_content_template_action_settings.sms_content_template_id",
		"rules.actions.set_content_template_action_settings.email_content_template_id",
	} {
		if _, ok := exporter.CustomAttributeResolver[path]; !ok {
			t.Fatalf("expected CustomAttributeResolver for %q", path)
		}
	}
}

func TestSetContentTemplateActionSettingsSchema_OptionalFields(t *testing.T) {
	smsSchema := setContentTemplateActionSettingsResource.Schema["sms_content_template_id"]
	emailSchema := setContentTemplateActionSettingsResource.Schema["email_content_template_id"]

	if smsSchema.Required || emailSchema.Required {
		t.Fatal("content template IDs should be optional")
	}
}

func TestValidateSetContentTemplateActionSettings(t *testing.T) {
	settings := schema.NewSet(schema.HashResource(setContentTemplateActionSettingsResource), []interface{}{
		map[string]interface{}{
			"email_content_template_id": "0f31dd7a-d158-4ec2-bec1-5bc3a224eb17",
		},
	})
	if err := validateSetContentTemplateActionSettings(settings, 0, 0); err != nil {
		t.Fatalf("email-only settings should be valid: %v", err)
	}

	emptySettings := schema.NewSet(schema.HashResource(setContentTemplateActionSettingsResource), []interface{}{
		map[string]interface{}{},
	})
	if err := validateSetContentTemplateActionSettings(emptySettings, 0, 0); err == nil {
		t.Fatal("expected error when neither template ID is set")
	}
}

func TestFlattenSetContentTemplateActionSettings_OnlyEmailFromAPI(t *testing.T) {
	emailID := "64de0ae3-e359-498e-91e7-0fc1baf99aa1"
	settings := &platformclientv2.Setcontenttemplateactionsettings{
		EmailContentTemplateId: platformclientv2.String(emailID),
	}

	flat := flattenSetContentTemplateActionSettings(settings)
	if flat == nil || flat.Len() != 1 {
		t.Fatalf("expected one flattened block, got %v", flat)
	}

	item := flat.List()[0].(map[string]interface{})
	if item["email_content_template_id"] != emailID {
		t.Fatalf("expected email id %q, got %v", emailID, item["email_content_template_id"])
	}
	if _, ok := item["sms_content_template_id"]; ok {
		t.Fatalf("expected sms_content_template_id to be absent from API flatten, got %v", item["sms_content_template_id"])
	}
}
