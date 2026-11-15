package task_management_workitem

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitGetAllTaskManagementWorkitemFirstPageOmitsAfter guards against sending "after": "" on the first
// page of the /workbins/query call and of the per-workbin /workitems/query call that
// getAllTaskManagementWorkitemFn makes. The API rejects an empty after with a 400, so the key must be absent
// on the first request of each loop and only be set from the previous response on later pages.
func TestUnitGetAllTaskManagementWorkitemFirstPageOmitsAfter(t *testing.T) {
	var workbinBodies, workitemBodies []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		body := map[string]any{}
		require.NoError(t, json.Unmarshal(raw, &body))

		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/workitems/query") {
			workitemBodies = append(workitemBodies, body)
			if _, ok := body["after"]; !ok {
				_, _ = w.Write([]byte(`{"entities":[{"id":"wi1"}],"after":"wipage2"}`))
				return
			}
			_, _ = w.Write([]byte(`{"entities":[{"id":"wi2"}]}`))
			return
		}

		workbinBodies = append(workbinBodies, body)
		if _, ok := body["after"]; !ok {
			_, _ = w.Write([]byte(`{"entities":[{"id":"wb1"}],"after":"page2"}`))
			return
		}
		_, _ = w.Write([]byte(`{"entities":[{"id":"wb2"}]}`))
	}))
	defer srv.Close()

	cfg := platformclientv2.NewConfiguration()
	cfg.BasePath = srv.URL
	cfg.AccessToken = "test-token"

	_, _, err := getAllTaskManagementWorkitemFn(context.Background(), newTaskManagementWorkitemProxy(cfg))
	require.NoError(t, err)

	require.Len(t, workbinBodies, 2)
	assert.NotContains(t, workbinBodies[0], "after", "first page request must not include 'after'")
	assert.Equal(t, "page2", workbinBodies[1]["after"], "second page request must use the 'after' returned by the first page")

	// Two workbins, each paged through two /workitems/query requests
	require.Len(t, workitemBodies, 4)
	for _, first := range []int{0, 2} {
		assert.NotContains(t, workitemBodies[first], "after", "first workitems page request must not include 'after'")
		assert.Equal(t, "wipage2", workitemBodies[first+1]["after"], "second workitems page request must use the 'after' returned by the first page")
	}
}
