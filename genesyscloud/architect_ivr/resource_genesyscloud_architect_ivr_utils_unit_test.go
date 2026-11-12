package architect_ivr

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func TestUnitBuildDnisFromResourceData(t *testing.T) {
	resourceSchema := ResourceArchitectIvrConfig().Schema

	t.Run("returns the configured numbers in order", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
			"name": "ivr",
			"dnis": []interface{}{"+13175550303", "+13175550301"},
		})
		dnis := buildDnisFromResourceData(d)
		assert.NotNil(t, dnis)
		assert.Equal(t, []string{"+13175550303", "+13175550301"}, *dnis)
	})

	t.Run("returns nil when dnis is not configured", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
			"name": "ivr",
		})
		assert.Nil(t, buildDnisFromResourceData(d))
	})
}

func TestUnitOrderDnisLikeCurrent(t *testing.T) {
	resourceSchema := ResourceArchitectIvrConfig().Schema

	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
		"name": "ivr",
		"dnis": []interface{}{"+13175550303", "+13175550301", "+13175550399"},
	})

	// The API returns a different order, drops +13175550399 and has two numbers that are not in the config yet.
	apiDnis := []string{"+13175550301", "+13175550305", "+13175550303", "+13175550304"}
	got := orderDnisLikeCurrent(d, apiDnis)

	assert.Equal(t, []interface{}{"+13175550303", "+13175550301", "+13175550304", "+13175550305"}, got)
}

func TestUnitOrderDnisLikeCurrentEmpty(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceArchitectIvrConfig().Schema, map[string]interface{}{
		"name": "ivr",
	})
	assert.Nil(t, orderDnisLikeCurrent(d, []string{}))
}

func TestUnitStateUpgraderArchitectIvrV1ToV2(t *testing.T) {
	rawState := map[string]interface{}{
		"name": "ivr",
		"dnis": []interface{}{"+13175550303", "+13175550301", "+13175550302"},
	}

	upgraded, err := stateUpgraderArchitectIvrV1ToV2(context.Background(), rawState, nil)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"+13175550301", "+13175550302", "+13175550303"}, upgraded["dnis"])
	assert.Equal(t, "ivr", upgraded["name"])

	// A state without dnis is left alone.
	upgraded, err = stateUpgraderArchitectIvrV1ToV2(context.Background(), map[string]interface{}{"name": "ivr"}, nil)
	assert.NoError(t, err)
	_, hasDnis := upgraded["dnis"]
	assert.False(t, hasDnis)
}

// A failed update must leave the prior dnis in state, not the planned ones.
func TestUnitResourceArchitectIvrUpdateFailureKeepsPriorState(t *testing.T) {
	tId := uuid.NewString()
	tName := "My Unit Test IVR"
	priorDnis := []string{"+13175550301", "+13175550302"}
	blockedDnis := "+13175550304"
	plannedDnis := []interface{}{priorDnis[0], priorDnis[1], blockedDnis}

	archProxy := &ArchitectIvrProxy{}
	archProxy.GetArchitectIvrAttr = func(ctx context.Context, a *ArchitectIvrProxy, id string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		return &platformclientv2.Ivr{Id: &tId, Name: &tName, Dnis: &priorDnis}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	archProxy.updateArchitectIvrAttr = func(ctx context.Context, a *ArchitectIvrProxy, id string, ivr platformclientv2.Ivr) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		// Use a status that is not retried so the test does not wait on the retry loop.
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusForbidden}, fmt.Errorf("API Error: 400 - DUPLICATE_VALUE")
	}
	internalProxy = archProxy
	defer func() { internalProxy = nil }()

	res := ResourceArchitectIvrConfig()
	prior := &terraform.InstanceState{
		ID: tId,
		Attributes: map[string]string{
			"id":     tId,
			"name":   tName,
			"dnis.#": "2",
			"dnis.0": priorDnis[0],
			"dnis.1": priorDnis[1],
		},
	}
	cfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"name": tName,
		"dnis": plannedDnis,
	})
	diff, err := res.Diff(context.Background(), prior, cfg, nil)
	assert.NoError(t, err)

	d, err := schema.InternalMap(res.Schema).Data(prior, diff)
	assert.NoError(t, err)
	// The planned value is what the update would have written.
	assert.Len(t, d.Get("dnis").([]interface{}), 3)

	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	diags := updateIvrConfig(context.Background(), d, gcloud)
	assert.True(t, diags.HasError())

	state := d.State()
	assert.NotNil(t, state)
	assert.Equal(t, "2", state.Attributes["dnis.#"])
	for key, value := range state.Attributes {
		assert.NotEqual(t, blockedDnis, value, "state should not contain the number that failed to be added (%s)", key)
	}
}
