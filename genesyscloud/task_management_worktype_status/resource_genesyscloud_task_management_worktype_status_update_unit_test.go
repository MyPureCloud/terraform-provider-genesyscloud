package task_management_worktype_status

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newStatusUpdateData builds the ResourceData the update function sees: prior state plus the planned diff
// for the given config. This is how the SDK evaluates HasChange on an update, so it reflects a real apply.
func newStatusUpdateData(t *testing.T, priorState map[string]string, config map[string]interface{}) *schema.ResourceData {
	t.Helper()
	res := ResourceTaskManagementWorktypeStatus()

	state := &terraform.InstanceState{ID: "wt-1/st-1", Attributes: priorState}
	diff, err := res.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
	require.NoError(t, err)

	d, err := schema.InternalMap(res.Schema).Data(state, diff)
	require.NoError(t, err)
	return d
}

// statusUpdateBody returns the JSON the PATCH request would carry.
func statusUpdateBody(t *testing.T, d *schema.ResourceData) map[string]interface{} {
	t.Helper()
	update := buildWorktypeStatusUpdate(d)
	raw, err := json.Marshal(update)
	require.NoError(t, err)

	body := map[string]interface{}{}
	require.NoError(t, json.Unmarshal(raw, &body))
	return body
}

func baseStatusState(description string) map[string]string {
	state := map[string]string{
		"id":          "wt-1/st-1",
		"worktype_id": "wt-1",
		"name":        "my-status",
		"category":    "Waiting",
	}
	if description != "" {
		state["description"] = description
	}
	return state
}

func baseStatusConfig() map[string]interface{} {
	return map[string]interface{}{
		"worktype_id": "wt-1",
		"name":        "my-status",
		"category":    "Waiting",
	}
}

// Removing description from the config (the DEVTOOLING-1904 / issue #2663 case) must send an explicit empty
// description. Before the fix the field was dropped from the PATCH body, so the API kept the old value.
func TestUnitBuildWorktypeStatusUpdateClearsRemovedDescription(t *testing.T) {
	d := newStatusUpdateData(t, baseStatusState("old description"), baseStatusConfig())

	body := statusUpdateBody(t, d)

	desc, present := body["description"]
	assert.True(t, present, "description must be present in the PATCH body when it is removed")
	assert.Equal(t, "", desc)
}

func TestUnitBuildWorktypeStatusUpdateSendsChangedDescription(t *testing.T) {
	config := baseStatusConfig()
	config["description"] = "new description"
	d := newStatusUpdateData(t, baseStatusState("old description"), config)

	body := statusUpdateBody(t, d)

	assert.Equal(t, "new description", body["description"])
}

func TestUnitBuildWorktypeStatusUpdateSetsDescriptionFromEmpty(t *testing.T) {
	config := baseStatusConfig()
	config["description"] = "first description"
	d := newStatusUpdateData(t, baseStatusState(""), config)

	body := statusUpdateBody(t, d)

	assert.Equal(t, "first description", body["description"])
}

// An unchanged description is still sent, as it was before the fix, so nothing about the existing request changes.
func TestUnitBuildWorktypeStatusUpdateKeepsUnchangedDescription(t *testing.T) {
	config := baseStatusConfig()
	config["description"] = "same description"
	d := newStatusUpdateData(t, baseStatusState("same description"), config)

	body := statusUpdateBody(t, d)

	assert.Equal(t, "same description", body["description"])
}

// A status that never had a description and still has none must not start sending one.
func TestUnitBuildWorktypeStatusUpdateOmitsDescriptionWhenNeverSet(t *testing.T) {
	d := newStatusUpdateData(t, baseStatusState(""), baseStatusConfig())

	body := statusUpdateBody(t, d)

	_, present := body["description"]
	assert.False(t, present, "description must stay out of the body when it was never set")
}

// SetField switches the SDK to "only serialize fields that were set", so the other fields must still be sent.
func TestUnitBuildWorktypeStatusUpdateKeepsOtherFields(t *testing.T) {
	config := baseStatusConfig()
	config["description"] = "d"
	config["default_destination_status_id"] = "dest-1"
	config["destination_status_ids"] = []interface{}{"dest-1", "dest-2"}
	config["status_transition_delay_seconds"] = 90
	config["status_transition_time"] = "12:04:21"
	d := newStatusUpdateData(t, baseStatusState("d"), config)

	body := statusUpdateBody(t, d)

	assert.Equal(t, "my-status", body["name"])
	assert.Equal(t, "dest-1", body["defaultDestinationStatusId"])
	assert.Equal(t, []interface{}{"dest-1", "dest-2"}, body["destinationStatusIds"])
	assert.EqualValues(t, 90, body["statusTransitionDelaySeconds"])
	assert.Equal(t, "12:04:21", body["statusTransitionTime"])
}

// The update function rewrites composite "<worktypeId>/<statusId>" references after building the body. That must
// still reach the serialized request, because the SDK reads field values when it marshals.
func TestUnitBuildWorktypeStatusUpdateReflectsLaterIdRewrite(t *testing.T) {
	config := baseStatusConfig()
	config["default_destination_status_id"] = "wt-1/dest-1"
	d := newStatusUpdateData(t, baseStatusState(""), config)

	update := buildWorktypeStatusUpdate(d)
	require.NotNil(t, update.DefaultDestinationStatusId)
	_, id := SplitWorktypeStatusTerraformId(*update.DefaultDestinationStatusId)
	update.DefaultDestinationStatusId = &id

	raw, err := json.Marshal(update)
	require.NoError(t, err)
	body := map[string]interface{}{}
	require.NoError(t, json.Unmarshal(raw, &body))

	assert.Equal(t, "dest-1", body["defaultDestinationStatusId"])
}
