package user

import (
	"context"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	featureToggles "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func TestUnitGetVoicemailUserpoliciesByUserIdCacheHit(t *testing.T) {
	tfexporter_state.ActivateExporterState()

	userID := "user-voicemail-cache-test"
	cached := platformclientv2.Voicemailuserpolicy{
		Enabled: platformclientv2.Bool(true),
	}
	rc.SetCache(userVoicemailPolicyCache, userID, cached)

	policy, _, err := getVoicemailUserpoliciesByUserIdFn(context.Background(), &userProxy{}, userID)
	require.NoError(t, err)
	require.NotNil(t, policy)
	assert.True(t, *policy.Enabled)
}

func TestUnitGetUserRoutingUtilizationRawCacheHit(t *testing.T) {
	tfexporter_state.ActivateExporterState()

	userID := "user-utilization-cache-test"
	cached := []byte(`{"level":"Organization"}`)
	rc.SetCache(userRoutingUtilizationCache, userID, cached)

	rawBody, _, err := getUserRoutingUtilizationRaw(context.Background(), userID, &userProxy{})
	require.NoError(t, err)
	assert.JSONEq(t, string(cached), string(rawBody))
}

func TestUnitGetUsersByIds(t *testing.T) {
	str := func(s string) *string { return &s }
	users := map[string]*platformclientv2.User{
		"active":   {Id: str("active"), Email: str("a@x.com"), Name: str("A"), State: str("active")},
		"inactive": {Id: str("inactive"), Email: str("i@x.com"), Name: str("I"), State: str("inactive")},
		"deleted":  {Id: str("deleted"), Email: str("d@x.com"), Name: str("D"), State: str("deleted")},
	}
	proxy := &userProxy{userCache: rc.NewResourceCache[platformclientv2.User]()}
	proxy.getUserByIdAttr = func(_ context.Context, _ *userProxy, id string, expand []string, state string) (*platformclientv2.User, *platformclientv2.APIResponse, error) {
		assert.Equal(t, "any", state)
		assert.Equal(t, userExportExpands, expand)
		if u, ok := users[id]; ok {
			return u, nil, nil
		}
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, assert.AnError
	}
	internalProxy = proxy
	t.Setenv(userTestsActiveEnvVar, "true")
	defer func() { internalProxy = nil }()
	ids := []string{"active", "inactive", "deleted", "missing"}

	getByIds := func() (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
		return resourceExporter.GetResourcesByIds(context.Background(), ids, func(ctx context.Context, id string) (*resourceExporter.ResourceMeta, diag.Diagnostics) {
			return getUserMetaById(ctx, nil, id)
		})
	}
	res, diags := getByIds()
	require.False(t, diags.HasError())
	assert.Len(t, res, 2) // deleted and missing users are left out, like the list path
	expected, err := userResourceMeta(*users["active"])
	require.NoError(t, err)
	assert.Equal(t, expected, res["active"])
	assert.NotNil(t, rc.GetCacheItem(proxy.userCache, "active"), "cached like the list path so readUser avoids another call")

	t.Setenv(featureToggles.SkipInactiveUserExportToggleName(), "true")
	res, _ = getByIds()
	assert.Len(t, res, 1)
}
