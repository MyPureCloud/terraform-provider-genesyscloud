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

// TestUnitGetAllTaskManagementWorkitemWorkbinListingOmitsAfter guards against sending "after": "" on the
// first page of the /workbins/query call that getAllTaskManagementWorkitemFn uses to list workbins. The API
// rejects an empty after with a 400, so the key must be absent on the first request and only be set from
// the previous response on later pages.
func TestUnitGetAllTaskManagementWorkitemWorkbinListingOmitsAfter(t *testing.T) {
	var workbinBodies []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		body := map[string]any{}
		require.NoError(t, json.Unmarshal(raw, &body))

		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/workitems/query") {
			_, _ = w.Write([]byte(`{"entities":[]}`))
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
}
