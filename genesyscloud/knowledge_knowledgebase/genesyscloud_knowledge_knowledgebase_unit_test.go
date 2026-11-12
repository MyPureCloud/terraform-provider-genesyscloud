package knowledge_knowledgebase

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

// errInUseByBotFlow mirrors the 424 the Knowledge API returns while the dependency-tracking graph
// still shows a bot flow referencing the knowledge base.
var errInUseByBotFlow = fmt.Errorf("API Error: 424 - KnowledgeBase abc in use by Bot flow status unknown")

type deleteHarness struct {
	deleteCalls  int
	rebuildCalls int
}

// runKnowledgebaseDelete wires a stub proxy and stubbed rebuild into deleteKnowledgeKnowledgebase and runs it.
// deleteFn decides the result of each DELETE call (1-based call number).
func runKnowledgebaseDelete(
	t *testing.T,
	retryTimeout time.Duration,
	deleteFn func(call int) (*platformclientv2.APIResponse, error),
) (*deleteHarness, diag.Diagnostics) {
	t.Helper()
	h := &deleteHarness{}

	proxy := &knowledgebaseProxy{}
	proxy.deleteKnowledgebaseAttr = func(ctx context.Context, p *knowledgebaseProxy, id string) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
		h.deleteCalls++
		resp, err := deleteFn(h.deleteCalls)
		return nil, resp, err
	}
	// Used by the post-delete confirmation loop: 404 means the knowledge base is gone.
	proxy.getKnowledgebaseByIdAttr = func(ctx context.Context, p *knowledgebaseProxy, id string) (*knowledgebaseResponse, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("not found")
	}

	origProxy, origRebuild, origTimeout := internalProxy, rebuildDatabaseFn, deleteRetryTimeout
	internalProxy = proxy
	rebuildDatabaseFn = func(*platformclientv2.Configuration) diag.Diagnostics {
		h.rebuildCalls++
		return nil
	}
	deleteRetryTimeout = retryTimeout
	t.Cleanup(func() {
		internalProxy, rebuildDatabaseFn, deleteRetryTimeout = origProxy, origRebuild, origTimeout
	})

	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceKnowledgeKnowledgebase().Schema, map[string]interface{}{
		"name": "unit-test-kb",
	})
	d.SetId(uuid.NewString())

	return h, deleteKnowledgeKnowledgebase(context.Background(), d, gcloud)
}

// First DELETE hits the bot-flow 424; after the rebuild the retried DELETE succeeds.
func TestUnitKnowledgebaseDeleteInUseThenSuccess(t *testing.T) {
	h, diags := runKnowledgebaseDelete(t, 30*time.Second, func(call int) (*platformclientv2.APIResponse, error) {
		if call == 1 {
			return &platformclientv2.APIResponse{StatusCode: http.StatusFailedDependency}, errInUseByBotFlow
		}
		return &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}, nil
	})

	assert.False(t, diags.HasError(), "delete should succeed once the bot-flow dependency clears")
	assert.Equal(t, 1, h.rebuildCalls, "dependency tracking should be rebuilt exactly once")
	assert.Equal(t, 2, h.deleteCalls, "expected the initial delete plus one successful retry")
}

// The bot-flow 424 never clears, so the retry loop must give up at the timeout and surface an error.
func TestUnitKnowledgebaseDeleteInUseUntilTimeout(t *testing.T) {
	h, diags := runKnowledgebaseDelete(t, 1*time.Second, func(call int) (*platformclientv2.APIResponse, error) {
		return &platformclientv2.APIResponse{StatusCode: http.StatusFailedDependency}, errInUseByBotFlow
	})

	assert.True(t, diags.HasError(), "delete should fail when the dependency never clears")
	assert.Equal(t, 1, h.rebuildCalls, "rebuild runs once, not on every retry")
	assert.GreaterOrEqual(t, h.deleteCalls, 2, "the delete should have been retried at least once")
}

// A bot-flow 424 followed by a different error on retry must stop immediately rather than keep retrying.
func TestUnitKnowledgebaseDeleteNonBotFlowErrorOnRetryFailsFast(t *testing.T) {
	h, diags := runKnowledgebaseDelete(t, 30*time.Second, func(call int) (*platformclientv2.APIResponse, error) {
		if call == 1 {
			return &platformclientv2.APIResponse{StatusCode: http.StatusFailedDependency}, errInUseByBotFlow
		}
		return &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("API Error: 500 - boom")
	})

	assert.True(t, diags.HasError())
	assert.Equal(t, 1, h.rebuildCalls)
	assert.Equal(t, 2, h.deleteCalls, "a non-bot-flow error must not be retried")
}

// An unrelated error on the very first DELETE must not trigger a rebuild or any retry.
func TestUnitKnowledgebaseDeleteNonBotFlowErrorFailsFast(t *testing.T) {
	h, diags := runKnowledgebaseDelete(t, 30*time.Second, func(call int) (*platformclientv2.APIResponse, error) {
		return &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("API Error: 500 - boom")
	})

	assert.True(t, diags.HasError())
	assert.Equal(t, 0, h.rebuildCalls, "rebuild is only for the bot-flow dependency error")
	assert.Equal(t, 1, h.deleteCalls)
}

func strPtr(s string) *string { return &s }

func timePtr(t time.Time) *time.Time { return &t }

// The cleanup helper must recognise test knowledge bases by name or description prefix, and nothing else.
func TestUnitIsTestKnowledgeBase(t *testing.T) {
	const prefix = "Test-Terraform-Knowledge-Base"

	tests := []struct {
		name string
		kb   platformclientv2.Knowledgebase
		want bool
	}{
		{"name_prefix_matches", platformclientv2.Knowledgebase{Id: strPtr("1"), Name: strPtr(prefix + "abc")}, true},
		{"description_prefix_matches", platformclientv2.Knowledgebase{Id: strPtr("1"), Name: strPtr("other"), Description: strPtr("test-knowledgebase-description1")}, true},
		{"description_matches_without_name", platformclientv2.Knowledgebase{Id: strPtr("1"), Description: strPtr("test-knowledgebase-description2")}, true},
		{"unrelated_knowledge_base", platformclientv2.Knowledgebase{Id: strPtr("1"), Name: strPtr("Production KB"), Description: strPtr("real content")}, false},
		{"no_id_is_never_matched", platformclientv2.Knowledgebase{Name: strPtr(prefix + "abc")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isTestKnowledgeBase(tt.kb, prefix))
		})
	}
}

// The cleanup helper must only delete leftovers. A knowledge base created moments ago can belong to a test that
// is running in another package at the same time (this caused 404s in those tests), so it must be left alone.
func TestUnitIsOlderThan(t *testing.T) {
	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	const minAge = 30 * time.Minute

	tests := []struct {
		name string
		kb   platformclientv2.Knowledgebase
		want bool
	}{
		{"created_seconds_ago_is_kept", platformclientv2.Knowledgebase{DateCreated: timePtr(now.Add(-10 * time.Second))}, false},
		{"created_just_under_threshold_is_kept", platformclientv2.Knowledgebase{DateCreated: timePtr(now.Add(-minAge + time.Second))}, false},
		{"created_exactly_at_threshold_is_cleaned", platformclientv2.Knowledgebase{DateCreated: timePtr(now.Add(-minAge))}, true},
		{"created_yesterday_is_cleaned", platformclientv2.Knowledgebase{DateCreated: timePtr(now.Add(-24 * time.Hour))}, true},
		{"missing_creation_date_is_kept", platformclientv2.Knowledgebase{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isOlderThan(tt.kb, now, minAge))
		})
	}
}
