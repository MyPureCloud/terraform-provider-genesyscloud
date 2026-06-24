package journey_action_map

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
)

// TestFlattenActionMapOpenAction verifies flattenActionMap sets open_action_fields.open_action
// as a TypeSet-compatible value ([]map), not *[]map, which previously panicked during export read.
func TestFlattenActionMapOpenAction(t *testing.T) {
	actionMap := sampleActionMapForFlatten()
	d := schema.TestResourceDataRaw(t, ResourceJourneyActionMap().Schema, map[string]interface{}{})

	flattenActionMap(d, actionMap)

	actionSet, ok := d.Get("action").(*schema.Set)
	if !ok || actionSet.Len() != 1 {
		t.Fatalf("action: got %T (#%v), want *schema.Set with 1 item", d.Get("action"), setLen(d.Get("action")))
	}

	actionBlock := actionSet.List()[0].(map[string]interface{})
	if actionBlock["media_type"] != "openAction" {
		t.Fatalf("media_type = %v, want openAction", actionBlock["media_type"])
	}

	openActionFieldsSet, ok := actionBlock["open_action_fields"].(*schema.Set)
	if !ok || openActionFieldsSet.Len() != 1 {
		t.Fatalf("open_action_fields: got %T (#%v)", actionBlock["open_action_fields"], setLen(actionBlock["open_action_fields"]))
	}

	oaf := openActionFieldsSet.List()[0].(map[string]interface{})
	openActionSet, ok := oaf["open_action"].(*schema.Set)
	if !ok || openActionSet.Len() != 1 {
		t.Fatalf("open_action: got %T (#%v)", oaf["open_action"], setLen(oaf["open_action"]))
	}

	ref := openActionSet.List()[0].(map[string]interface{})
	if ref["id"] != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("open_action id = %v", ref["id"])
	}
	if ref["name"] != "export-repro-open-action" {
		t.Fatalf("open_action name = %v", ref["name"])
	}
}

func setLen(v interface{}) int {
	if s, ok := v.(*schema.Set); ok {
		return s.Len()
	}
	return -1
}

func sampleActionMapForFlatten() *platformclientv2.Actionmap {
	isActive := true
	displayName := "export-repro-action-map"
	weight := 2
	ignoreFrequencyCap := false
	activationType := "immediate"
	mediaType := "openAction"
	isPacingEnabled := true
	startDate := time.Date(2022, 7, 4, 12, 0, 0, 0, time.UTC)
	openActionID := "00000000-0000-0000-0000-000000000001"
	openActionName := "export-repro-open-action"

	return &platformclientv2.Actionmap{
		IsActive:           &isActive,
		DisplayName:        &displayName,
		Weight:             &weight,
		IgnoreFrequencyCap: &ignoreFrequencyCap,
		StartDate:          &startDate,
		Activation: &platformclientv2.Activation{
			VarType: &activationType,
		},
		Action: &platformclientv2.Actionmapaction{
			MediaType:       &mediaType,
			IsPacingEnabled: &isPacingEnabled,
			OpenActionFields: &platformclientv2.Openactionfields{
				OpenAction: &platformclientv2.Domainentityref{
					Id:   &openActionID,
					Name: &openActionName,
				},
			},
		},
	}
}
