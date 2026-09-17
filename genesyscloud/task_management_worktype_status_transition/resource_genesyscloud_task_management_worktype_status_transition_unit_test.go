package task_management_worktype_status_transition

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitBuildWorkitemStatusTransitionPatchClearsOptionalFieldsOnUpdate(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceTaskManagementWorktypeStatusTransition().Schema, map[string]interface{}{
		"worktype_id": "wt-id",
		"status_id":   "status-id",
	})
	d.SetId("wt-id/status-id transition")

	name := "Open"
	description := "status description"
	status := &platformclientv2.Workitemstatus{
		Name:        &name,
		Description: &description,
	}

	body := buildWorkitemStatusTransitionPatch(d, status, &[]string{}, nil, false)
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &result))

	assert.Equal(t, name, result["name"])
	require.Contains(t, result, "defaultDestinationStatusId")
	assert.Nil(t, result["defaultDestinationStatusId"])
	require.Contains(t, result, "statusTransitionDelaySeconds")
	assert.Nil(t, result["statusTransitionDelaySeconds"])
	require.Contains(t, result, "statusTransitionTime")
	assert.Nil(t, result["statusTransitionTime"])
	require.Contains(t, result, "destinationStatusIds")
	assert.Empty(t, result["destinationStatusIds"])
}

func TestUnitBuildWorkitemStatusTransitionPatchOmitsOptionalFieldsOnCreate(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceTaskManagementWorktypeStatusTransition().Schema, map[string]interface{}{
		"worktype_id": "wt-id",
		"status_id":   "status-id",
	})

	name := "Open"
	status := &platformclientv2.Workitemstatus{Name: &name}

	body := buildWorkitemStatusTransitionPatch(d, status, &[]string{"dest-1"}, nil, true)
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &result))

	assert.NotContains(t, result, "defaultDestinationStatusId")
	assert.NotContains(t, result, "statusTransitionDelaySeconds")
	assert.NotContains(t, result, "statusTransitionTime")
	assert.Equal(t, []interface{}{"dest-1"}, result["destinationStatusIds"])
}

func TestUnitWorkitemStatusTransitionMatchesConfigClearedDefault(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceTaskManagementWorktypeStatusTransition().Schema, map[string]interface{}{
		"worktype_id": "wt-id",
		"status_id":   "status-id",
	})

	assert.True(t, workitemStatusTransitionMatchesConfig(d, &platformclientv2.Workitemstatus{}))

	stillSet := "dest-status"
	assert.False(t, workitemStatusTransitionMatchesConfig(d, &platformclientv2.Workitemstatus{
		DefaultDestinationStatus: &platformclientv2.Workitemstatusreference{Id: &stillSet},
	}))
}

func TestUnitWorkitemStatusTransitionMatchesConfigDestinationIds(t *testing.T) {
	destA := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	destB := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	d := schema.TestResourceDataRaw(t, ResourceTaskManagementWorktypeStatusTransition().Schema, map[string]interface{}{
		"worktype_id":            "wt-id",
		"status_id":              "status-id",
		"destination_status_ids": []interface{}{"wt-id/" + destA, destB},
	})

	assert.True(t, workitemStatusTransitionMatchesConfig(d, &platformclientv2.Workitemstatus{
		DestinationStatuses: &[]platformclientv2.Workitemstatusreference{
			{Id: &destB},
			{Id: &destA},
		},
	}))

	other := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	assert.False(t, workitemStatusTransitionMatchesConfig(d, &platformclientv2.Workitemstatus{
		DestinationStatuses: &[]platformclientv2.Workitemstatusreference{
			{Id: &other},
		},
	}))
}

func TestUnitIsNoChangeForRecordError(t *testing.T) {
	resp := &platformclientv2.APIResponse{ErrorMessage: "API Error: 400 - No change for the record is obtained."}
	assert.True(t, isNoChangeForRecordError(nil, resp))
	assert.False(t, isNoChangeForRecordError(nil, &platformclientv2.APIResponse{ErrorMessage: "something else"}))
}

func TestUnitSetDestinationStatusIdsStateKeepsEmptyList(t *testing.T) {
	destID := "fd9ceb95-0c76-4962-956b-fd1f18654f4e"
	status := &platformclientv2.Workitemstatus{
		DestinationStatuses: &[]platformclientv2.Workitemstatusreference{{Id: &destID}},
	}

	d := schema.TestResourceDataRaw(t, ResourceTaskManagementWorktypeStatusTransition().Schema, map[string]interface{}{
		"worktype_id": "wt-id",
		"status_id":   "status-id",
	})
	setDestinationStatusIdsState(d, status, false)
	assert.Empty(t, d.Get("destination_status_ids"))
}

func TestUnitSetDestinationStatusIdsStatePopulatesOnImport(t *testing.T) {
	destID := "fd9ceb95-0c76-4962-956b-fd1f18654f4e"
	status := &platformclientv2.Workitemstatus{
		DestinationStatuses: &[]platformclientv2.Workitemstatusreference{{Id: &destID}},
	}

	d := schema.TestResourceDataRaw(t, ResourceTaskManagementWorktypeStatusTransition().Schema, map[string]interface{}{})
	setDestinationStatusIdsState(d, status, true)
	got := d.Get("destination_status_ids").([]interface{})
	require.Len(t, got, 1)
	assert.Equal(t, destID, got[0])
}
