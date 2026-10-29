//go:build unit
// +build unit

package integration_action

import (
	"context"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

// TestUnitUpdateDraftPreservesActionId is a regression test for DEVTOOLING-1886 / GitHub #2638.
//
// The bug: updating a function-data-action's zip created the draft via the drafts-COLLECTION
// endpoint (createIntegrationActionDraft -> POST /api/v2/integrations/actions/drafts), which
// mints a BRAND-NEW action id, and then overwrote state with that new id. The fix routes the
// update through the id-SCOPED draft endpoint (createIntegrationActionDraftFromExisting ->
// POST /api/v2/integrations/actions/{actionId}/draft), which preserves the original id.
//
// This test exercises the proxy contract directly: it asserts that
//   - createIntegrationActionDraftFromExisting forwards the EXISTING action id unchanged, and
//   - the collection-draft create (createIntegrationActionDraft) is NOT what the update relies on
//     to obtain the id.
func TestUnitUpdateDraftPreservesActionId(t *testing.T) {
	const existingID = "custom_-_11111111-2222-3333-4444-555555555555"

	var fromExistingCalledWith string
	var collectionCreateCalled bool

	proxy := &integrationActionsProxy{
		// id-scoped draft create: record the id it was called with and echo it back.
		createIntegrationActionDraftFromExistingAttr: func(_ context.Context, _ *integrationActionsProxy, actionId string) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
			fromExistingCalledWith = actionId
			id := actionId
			return &platformclientv2.Action{Id: &id}, nil, nil
		},
		// collection draft create: if this is ever used to obtain the id on update, flag it.
		createIntegrationActionDraftAttr: func(_ context.Context, _ *integrationActionsProxy, _ *IntegrationAction) (*IntegrationAction, *platformclientv2.APIResponse, error) {
			collectionCreateCalled = true
			newID := "custom_-_99999999-9999-9999-9999-999999999999"
			return &IntegrationAction{Id: &newID}, nil, nil
		},
	}

	action, _, err := proxy.createIntegrationActionDraftFromExisting(context.Background(), existingID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fromExistingCalledWith != existingID {
		t.Fatalf("expected id-scoped draft create to be called with %q, got %q", existingID, fromExistingCalledWith)
	}
	if action == nil || action.Id == nil || *action.Id != existingID {
		t.Fatalf("expected draft create to preserve action id %q, got %+v", existingID, action)
	}
	if collectionCreateCalled {
		t.Fatal("collection draft create must not be used to obtain the action id on update (it mints a new id)")
	}
}

// TestUnitPatchDraftPassesVersion verifies the draft PATCH forwards the action id and the
// version required for optimistic concurrency (a missing version produced a 400 from the API).
func TestUnitPatchDraftPassesVersion(t *testing.T) {
	const existingID = "custom_-_aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	wantVersion := 7

	var gotID string
	var gotVersion *int

	proxy := &integrationActionsProxy{
		patchIntegrationActionDraftAttr: func(_ context.Context, _ *integrationActionsProxy, actionId string, body platformclientv2.Updatedraftinput) (*platformclientv2.Action, *platformclientv2.APIResponse, error) {
			gotID = actionId
			gotVersion = body.Version
			id := actionId
			return &platformclientv2.Action{Id: &id}, nil, nil
		},
	}

	_, _, err := proxy.patchIntegrationActionDraft(context.Background(), existingID, platformclientv2.Updatedraftinput{
		Version: &wantVersion,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != existingID {
		t.Fatalf("expected patch to target id %q, got %q", existingID, gotID)
	}
	if gotVersion == nil || *gotVersion != wantVersion {
		t.Fatalf("expected patch to forward version %d, got %v", wantVersion, gotVersion)
	}
}
