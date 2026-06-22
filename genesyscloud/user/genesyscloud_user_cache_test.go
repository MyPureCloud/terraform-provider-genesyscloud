package user

import (
	"context"
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
