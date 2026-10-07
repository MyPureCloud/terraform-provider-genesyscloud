package task_management_worktype

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
Unit tests for open_status_default and closed_status_auto_terminate: the two fields that let a
genesyscloud_task_management_worktype configure the Open/Closed statuses Genesys Cloud
automatically creates for it. See GitHub issue #2201.

Covers: both fields set, only one field set, neither field set, and the false/unset error case.
*/

// newWorktypeTestResourceData builds ResourceData for ResourceTaskManagementWorktype with the
// minimal required fields set, optionally including open_status_default/closed_status_auto_terminate.
func newWorktypeTestResourceData(t *testing.T, tId string, extra map[string]interface{}) *schema.ResourceData {
	resourceDataMap := map[string]interface{}{
		"id":                 tId,
		"name":               "tf_worktype_" + uuid.NewString(),
		"default_workbin_id": uuid.NewString(),
		"schema_id":          uuid.NewString(),
	}
	for k, v := range extra {
		resourceDataMap[k] = v
	}

	d := schema.TestResourceDataRaw(t, ResourceTaskManagementWorktype().Schema, resourceDataMap)
	d.SetId(tId)
	return d
}

// newMockStatusProxy builds a TaskManagementWorktypeProxy whose getWorktypeStatusByCategoryAttr
// returns a fixed Open/Closed status id, and whose UpdateTaskManagementWorktype/auto-terminate
// PATCH calls are captured for assertions.
func newMockStatusProxy(openStatusId, closedStatusId string) (*TaskManagementWorktypeProxy, *[]platformclientv2.Worktypeupdate, *[]bool) {
	var capturedWorktypeUpdates []platformclientv2.Worktypeupdate
	var capturedAutoTerminateValues []bool

	proxy := &TaskManagementWorktypeProxy{}

	proxy.getWorktypeStatusByCategoryAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, worktypeId string, category string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
		id := openStatusId
		if category == closedStatusCategory {
			id = closedStatusId
		}
		return &platformclientv2.Workitemstatus{Id: &id, Category: &category}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	proxy.updateTaskManagementWorktypeAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, id string, update *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
		capturedWorktypeUpdates = append(capturedWorktypeUpdates, *update)
		return &platformclientv2.Worktype{Id: &id}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	proxy.patchWorktypeStatusAutoTerminateAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, worktypeId string, statusId string, autoTerminate bool) (*platformclientv2.APIResponse, error) {
		capturedAutoTerminateValues = append(capturedAutoTerminateValues, autoTerminate)
		return &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	return proxy, &capturedWorktypeUpdates, &capturedAutoTerminateValues
}

func TestUnitApplyAutoCreatedStatusSettings_BothFieldsSet(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, worktypeUpdates, autoTerminateValues := newMockStatusProxy(openStatusId, closedStatusId)

	d := newWorktypeTestResourceData(t, tId, map[string]interface{}{
		"open_status_default":          true,
		"closed_status_auto_terminate": true,
	})

	diags := applyAutoCreatedStatusSettings(context.Background(), proxy, d)
	require.False(t, diags.HasError(), diags)

	require.Len(t, *worktypeUpdates, 1, "expected exactly one worktype update call to set the default status")
	assert.Equal(t, openStatusId, *(*worktypeUpdates)[0].DefaultStatusId, "Open status id should be sent as the new default status")

	require.Len(t, *autoTerminateValues, 1, "expected exactly one auto-terminate patch call")
	assert.True(t, (*autoTerminateValues)[0])
}

func TestUnitApplyAutoCreatedStatusSettings_OnlyOpenStatusDefaultSet(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, worktypeUpdates, autoTerminateValues := newMockStatusProxy(openStatusId, closedStatusId)

	d := newWorktypeTestResourceData(t, tId, map[string]interface{}{
		"open_status_default": true,
		// closed_status_auto_terminate intentionally omitted
	})

	diags := applyAutoCreatedStatusSettings(context.Background(), proxy, d)
	require.False(t, diags.HasError(), diags)

	require.Len(t, *worktypeUpdates, 1, "Open default status update should still happen")
	assert.Equal(t, openStatusId, *(*worktypeUpdates)[0].DefaultStatusId)

	assert.Empty(t, *autoTerminateValues, "auto-terminate should not be touched when closed_status_auto_terminate is not set")
}

func TestUnitApplyAutoCreatedStatusSettings_OnlyClosedAutoTerminateSet(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, worktypeUpdates, autoTerminateValues := newMockStatusProxy(openStatusId, closedStatusId)

	d := newWorktypeTestResourceData(t, tId, map[string]interface{}{
		// open_status_default intentionally omitted
		"closed_status_auto_terminate": true,
	})

	diags := applyAutoCreatedStatusSettings(context.Background(), proxy, d)
	require.False(t, diags.HasError(), diags)

	assert.Empty(t, *worktypeUpdates, "default status should not be touched when open_status_default is not set")

	require.Len(t, *autoTerminateValues, 1)
	assert.True(t, (*autoTerminateValues)[0])
}

func TestUnitApplyAutoCreatedStatusSettings_NeitherFieldSet(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, worktypeUpdates, autoTerminateValues := newMockStatusProxy(openStatusId, closedStatusId)

	d := newWorktypeTestResourceData(t, tId, nil)

	diags := applyAutoCreatedStatusSettings(context.Background(), proxy, d)
	require.False(t, diags.HasError(), diags)

	assert.Empty(t, *worktypeUpdates, "no API call should be made when open_status_default is omitted")
	assert.Empty(t, *autoTerminateValues, "no API call should be made when closed_status_auto_terminate is omitted")
}

func TestUnitApplyAutoCreatedStatusSettings_OpenStatusDefaultFalseReturnsError(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, worktypeUpdates, _ := newMockStatusProxy(openStatusId, closedStatusId)

	d := newWorktypeTestResourceData(t, tId, map[string]interface{}{
		"open_status_default": false,
	})
	// schema.TestResourceDataRaw's GetOkExists treats an explicit `false` the same as "present in
	// config" only when the raw config map actually contains the key, which it does here.
	require.NotNil(t, d.Get("open_status_default"))

	diags := applyAutoCreatedStatusSettings(context.Background(), proxy, d)
	require.True(t, diags.HasError(), "setting open_status_default to false should return an error")
	assert.Contains(t, diags[0].Summary+diags[0].Detail, "cannot be set to false")

	assert.Empty(t, *worktypeUpdates, "no worktype update should be attempted when open_status_default=false is rejected")
}

// TestUnitApplyAutoCreatedStatusSettings_OpenAlreadyDefaultIsToleratedAsSuccess reproduces a real
// API behavior observed against a live org: Genesys Cloud may already set the auto-created Open
// status as the worktype's default status on creation, so a PATCH re-asserting the same
// defaultStatusId returns 400 "No change for the record is obtained". That response must be
// treated as success, not surfaced as an error, since it means the desired state already holds.
func TestUnitApplyAutoCreatedStatusSettings_OpenAlreadyDefaultIsToleratedAsSuccess(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, _, _ := newMockStatusProxy(openStatusId, closedStatusId)
	proxy.updateTaskManagementWorktypeAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, id string, update *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
		resp := &platformclientv2.APIResponse{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: "No change for the record is obtained.",
		}
		return nil, resp, fmt.Errorf("API Error: 400 - No change for the record is obtained.")
	}

	d := newWorktypeTestResourceData(t, tId, map[string]interface{}{
		"open_status_default": true,
	})

	diags := applyAutoCreatedStatusSettings(context.Background(), proxy, d)
	require.False(t, diags.HasError(), "an already-default status must not be treated as an error: %v", diags)
}

// TestUnitUpdateTaskManagementWorktype_SkipsBasePatchWhenOnlyVirtualFieldsChanged reproduces a
// real API failure observed against a live org: getWorktypeupdateFromResourceData always sets
// Name (no HasChange guard), so when the only changed fields in an apply are
// open_status_default/closed_status_auto_terminate (which are not base Worktypeupdate fields),
// the base update call would send a PATCH containing only the unchanged Name, and Genesys Cloud
// rejects it with 400 "No change for the record is obtained". The base update must be skipped
// entirely in that case.
func TestUnitUpdateTaskManagementWorktype_SkipsBasePatchWhenOnlyVirtualFieldsChanged(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, _, autoTerminateValues := newMockStatusProxy(openStatusId, closedStatusId)

	baseUpdateCalls := 0
	proxy.updateTaskManagementWorktypeAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, id string, update *platformclientv2.Worktypeupdate) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
		baseUpdateCalls++
		return &platformclientv2.Worktype{Id: &id, Name: update.Name}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getTaskManagementWorktypeByIdAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, id string) (*platformclientv2.Worktype, *platformclientv2.APIResponse, error) {
		name := "tf_worktype"
		return &platformclientv2.Worktype{Id: &id, Name: &name}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	sch := ResourceTaskManagementWorktype().Schema
	state := &terraform.InstanceState{
		Attributes: map[string]string{
			"name":                         "tf_worktype",
			"default_workbin_id":           uuid.NewString(),
			"schema_id":                    uuid.NewString(),
			"closed_status_auto_terminate": "false",
		},
	}
	// Only closed_status_auto_terminate changes; no base worktype field is in this diff.
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"closed_status_auto_terminate": {Old: "false", New: "true"},
		},
	}
	d, err := schema.InternalMap(sch).Data(state, diff)
	require.NoError(t, err)
	d.SetId(tId)

	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	diags := updateTaskManagementWorktype(context.Background(), d, gcloud)
	require.False(t, diags.HasError(), diags)

	assert.Zero(t, baseUpdateCalls, "the base worktype update PATCH must be skipped when only virtual fields changed")
	require.Len(t, *autoTerminateValues, 1, "the auto-terminate status update must still happen")
	assert.True(t, (*autoTerminateValues)[0])
}

// TestUnitApplyAutoCreatedStatusSettings_StatusWithNilIdDoesNotPanic guards against a nil pointer
// dereference: the proxy's status lookup is a swappable seam (used directly by tests and
// indirectly by future refactors), and nothing in its signature guarantees a returned status has
// a non-nil Id. Both apply paths must surface a clear error instead of panicking if that ever
// happens.
func TestUnitApplyAutoCreatedStatusSettings_StatusWithNilIdDoesNotPanic(t *testing.T) {
	tId := uuid.NewString()

	proxy := &TaskManagementWorktypeProxy{}
	proxy.getWorktypeStatusByCategoryAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, worktypeId string, category string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
		// Status found, but with a nil Id - e.g. a partially populated API response.
		return &platformclientv2.Workitemstatus{Category: &category}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	d := newWorktypeTestResourceData(t, tId, map[string]interface{}{
		"open_status_default":          true,
		"closed_status_auto_terminate": true,
	})

	require.NotPanics(t, func() {
		diags := applyAutoCreatedStatusSettings(context.Background(), proxy, d)
		require.True(t, diags.HasError(), "a status with a nil Id must surface as an error, not succeed silently")
	})
}

// TestUnitApplyAutoCreatedStatusSettings_NilStatusDoesNotPanic guards against a nil pointer
// dereference if the proxy's status lookup ever returns a nil status alongside a nil error - a
// combination the function signature does not forbid, even though the current real
// implementation never produces it.
func TestUnitApplyAutoCreatedStatusSettings_NilStatusDoesNotPanic(t *testing.T) {
	tId := uuid.NewString()

	proxy := &TaskManagementWorktypeProxy{}
	proxy.getWorktypeStatusByCategoryAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, worktypeId string, category string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	d := newWorktypeTestResourceData(t, tId, map[string]interface{}{
		"open_status_default": true,
	})

	require.NotPanics(t, func() {
		diags := applyAutoCreatedStatusSettings(context.Background(), proxy, d)
		require.True(t, diags.HasError(), "a nil status must surface as an error, not succeed silently")
	})
}

// TestUnitReadAutoCreatedStatusSettings_StatusWithNilIdDoesNotPanic mirrors the two cases above
// for the read path, which dereferences openStatus.Id/closedStatus.AutoTerminateWorkitem directly.
func TestUnitReadAutoCreatedStatusSettings_StatusWithNilIdDoesNotPanic(t *testing.T) {
	tId := uuid.NewString()

	proxy := &TaskManagementWorktypeProxy{}
	proxy.getWorktypeStatusByCategoryAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, worktypeId string, category string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	d := newWorktypeTestResourceData(t, tId, nil)
	worktype := &platformclientv2.Worktype{Id: &tId}

	require.NotPanics(t, func() {
		readAutoCreatedStatusSettings(context.Background(), proxy, d, worktype)
	})
}

func TestUnitApplyAutoCreatedStatusSettingsOnUpdate_OnlyActsOnChangedFields(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, worktypeUpdates, autoTerminateValues := newMockStatusProxy(openStatusId, closedStatusId)

	sch := ResourceTaskManagementWorktype().Schema
	state := &terraform.InstanceState{
		Attributes: map[string]string{
			"name":                         "tf_worktype",
			"default_workbin_id":           uuid.NewString(),
			"schema_id":                    uuid.NewString(),
			"open_status_default":          "true", // already true, unchanged in this apply
			"closed_status_auto_terminate": "false",
		},
	}
	// Only closed_status_auto_terminate is present in the diff (false -> true).
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"closed_status_auto_terminate": {Old: "false", New: "true"},
		},
	}
	d, err := schema.InternalMap(sch).Data(state, diff)
	require.NoError(t, err)
	d.SetId(tId)

	require.False(t, d.HasChange("open_status_default"), "open_status_default should NOT be flagged as changed")
	require.True(t, d.HasChange("closed_status_auto_terminate"), "closed_status_auto_terminate should be flagged as changed")

	diags := applyAutoCreatedStatusSettingsOnUpdate(context.Background(), proxy, d)
	require.False(t, diags.HasError(), diags)

	assert.Empty(t, *worktypeUpdates, "open_status_default did not change, so the default status API must not be called")
	require.Len(t, *autoTerminateValues, 1, "closed_status_auto_terminate changed, so the auto-terminate API must be called exactly once")
	assert.True(t, (*autoTerminateValues)[0])
}

func TestUnitApplyAutoCreatedStatusSettingsOnUpdate_NoChangesMeansNoApiCalls(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, worktypeUpdates, autoTerminateValues := newMockStatusProxy(openStatusId, closedStatusId)

	sch := ResourceTaskManagementWorktype().Schema
	state := &terraform.InstanceState{
		Attributes: map[string]string{
			"name":                         "tf_worktype",
			"default_workbin_id":           uuid.NewString(),
			"schema_id":                    uuid.NewString(),
			"open_status_default":          "true",
			"closed_status_auto_terminate": "true",
		},
	}
	diff := &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{}}
	d, err := schema.InternalMap(sch).Data(state, diff)
	require.NoError(t, err)
	d.SetId(tId)

	diags := applyAutoCreatedStatusSettingsOnUpdate(context.Background(), proxy, d)
	require.False(t, diags.HasError(), diags)

	assert.Empty(t, *worktypeUpdates, "no field changed, so no worktype update API call should happen")
	assert.Empty(t, *autoTerminateValues, "no field changed, so no auto-terminate API call should happen")
}

func TestUnitReadAutoCreatedStatusSettings_PopulatesBothFields(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, _, _ := newMockStatusProxy(openStatusId, closedStatusId)
	// Override the Closed status lookup to also report AutoTerminateWorkitem=true.
	proxy.getWorktypeStatusByCategoryAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, worktypeId string, category string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
		if category == closedStatusCategory {
			autoTerminate := true
			return &platformclientv2.Workitemstatus{Id: &closedStatusId, Category: &category, AutoTerminateWorkitem: &autoTerminate}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
		}
		return &platformclientv2.Workitemstatus{Id: &openStatusId, Category: &category}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	d := newWorktypeTestResourceData(t, tId, nil)

	worktype := &platformclientv2.Worktype{
		Id:            &tId,
		DefaultStatus: &platformclientv2.Workitemstatusreference{Id: &openStatusId},
	}

	readAutoCreatedStatusSettings(context.Background(), proxy, d, worktype)

	assert.True(t, d.Get("open_status_default").(bool), "Open status id matches worktype.DefaultStatus.Id, so open_status_default should read true")
	assert.True(t, d.Get("closed_status_auto_terminate").(bool))
}

func TestUnitReadAutoCreatedStatusSettings_OpenIsNotDefault(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()
	someOtherStatusId := uuid.NewString()

	proxy, _, _ := newMockStatusProxy(openStatusId, closedStatusId)

	d := newWorktypeTestResourceData(t, tId, nil)

	// A different status (not Open) is the worktype's default.
	worktype := &platformclientv2.Worktype{
		Id:            &tId,
		DefaultStatus: &platformclientv2.Workitemstatusreference{Id: &someOtherStatusId},
	}

	readAutoCreatedStatusSettings(context.Background(), proxy, d, worktype)

	assert.False(t, d.Get("open_status_default").(bool), "Open status id does not match worktype.DefaultStatus.Id, so open_status_default should read false")
}

func TestUnitReadAutoCreatedStatusSettings_LookupFailureDoesNotPanicOrError(t *testing.T) {
	tId := uuid.NewString()

	proxy := &TaskManagementWorktypeProxy{}
	proxy.getWorktypeStatusByCategoryAttr = func(ctx context.Context, p *TaskManagementWorktypeProxy, worktypeId string, category string) (*platformclientv2.Workitemstatus, *platformclientv2.APIResponse, error) {
		// Simulate an org where disable_default_status_creation=true, so there is no Open/Closed status.
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, assert.AnError
	}

	d := newWorktypeTestResourceData(t, tId, nil)
	worktype := &platformclientv2.Worktype{Id: &tId}

	// Must not panic when statuses cannot be found; the two fields simply keep their existing
	// (zero) value rather than surfacing an error that would break reading the worktype itself.
	require.NotPanics(t, func() {
		readAutoCreatedStatusSettings(context.Background(), proxy, d, worktype)
	})
}

func TestUnitApplyAutoCreatedStatusSettingsOnUpdate_OpenStatusDefaultChangedToFalseErrors(t *testing.T) {
	tId := uuid.NewString()
	openStatusId := uuid.NewString()
	closedStatusId := uuid.NewString()

	proxy, worktypeUpdates, _ := newMockStatusProxy(openStatusId, closedStatusId)

	sch := ResourceTaskManagementWorktype().Schema
	state := &terraform.InstanceState{
		Attributes: map[string]string{
			"name":                "tf_worktype",
			"default_workbin_id":  uuid.NewString(),
			"schema_id":           uuid.NewString(),
			"open_status_default": "true",
		},
	}
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"open_status_default": {Old: "true", New: "false"},
		},
	}
	d, err := schema.InternalMap(sch).Data(state, diff)
	require.NoError(t, err)
	d.SetId(tId)

	diags := applyAutoCreatedStatusSettingsOnUpdate(context.Background(), proxy, d)
	require.True(t, diags.HasError(), "changing open_status_default to false should error")
	assert.Empty(t, *worktypeUpdates)
}
