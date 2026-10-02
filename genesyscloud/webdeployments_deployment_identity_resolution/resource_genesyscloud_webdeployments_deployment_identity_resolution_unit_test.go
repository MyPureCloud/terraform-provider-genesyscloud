package webdeployments_deployment_identity_resolution

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	webDeploymentsDeployment "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/webdeployments_deployment"
	"github.com/stretchr/testify/assert"
)

// bypassConsistencyChecker disables the consistency checker for tests that deliberately
// simulate state the API disagrees with. Without it the checker keeps returning a
// retryable error and the read spins until the retry timeout.
func bypassConsistencyChecker(t *testing.T) {
	t.Setenv(feature_toggles.CCToggleName(), "true")
	t.Setenv("CONSISTENCY_CHECKS", "0")
}

// TestUnitResourceWebDeploymentIdentityResolutionUpdate asserts the PUT payload omits
// division and external source when they are unset, rather than sending empty objects
// (contacts-service rejects an empty division object with a 422).
func TestUnitResourceWebDeploymentIdentityResolutionUpdate(t *testing.T) {
	tDeploymentId := uuid.NewString()

	proxy := &webDeploymentIdentityResolutionProxy{}
	proxy.putWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, deploymentId string, config platformclientv2.Deploymentidentityresolutionconfig) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tDeploymentId, deploymentId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, false, *config.ResolveIdentities)
		assert.Nil(t, config.Division, "division must be omitted when division_id is unset")
		assert.Nil(t, config.ExternalSource, "external source must be omitted when external_source_id is unset")
		assert.NotNil(t, config.Automerge)
		assert.Equal(t, false, *config.Automerge.AuthenticatedWebMessaging)
		assert.Equal(t, true, *config.Automerge.WebTracking)

		return &config, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	proxy.getWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, _ string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		return &platformclientv2.Deploymentidentityresolutionconfig{
			ResolveIdentities: platformclientv2.Bool(false),
			Automerge: &platformclientv2.Identityresolutionautomergeconfig{
				AuthenticatedWebMessaging: platformclientv2.Bool(false),
				WebTracking:               platformclientv2.Bool(true),
			},
		}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceWebDeploymentIdentityResolution().Schema,
		buildIdentityResolutionResourceMap(tDeploymentId, false, "", "", true, true))
	d.SetId(tDeploymentId)

	diag := updateWebDeploymentIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tDeploymentId, d.Id())
	assert.Equal(t, tDeploymentId, d.Get("deployment_id").(string))
}

// TestUnitResourceWebDeploymentIdentityResolutionUpdateWithRefs asserts a real division id
// and external source id are sent through.
func TestUnitResourceWebDeploymentIdentityResolutionUpdateWithRefs(t *testing.T) {
	tDeploymentId := uuid.NewString()
	tDivisionId := uuid.NewString()
	tExternalSourceId := uuid.NewString()

	proxy := &webDeploymentIdentityResolutionProxy{}
	proxy.putWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, _ string, config platformclientv2.Deploymentidentityresolutionconfig) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.NotNil(t, config.Division)
		assert.Equal(t, tDivisionId, *config.Division.Id)
		assert.NotNil(t, config.ExternalSource)
		assert.Equal(t, tExternalSourceId, *config.ExternalSource.Id)

		return &config, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	proxy.getWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, _ string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		return &platformclientv2.Deploymentidentityresolutionconfig{
			ResolveIdentities: platformclientv2.Bool(true),
			Division:          &platformclientv2.Writablestarrabledivision{Id: &tDivisionId},
			ExternalSource:    &platformclientv2.Identityresolutionexternalsource{Id: &tExternalSourceId},
		}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceWebDeploymentIdentityResolution().Schema,
		buildIdentityResolutionResourceMap(tDeploymentId, true, tDivisionId, tExternalSourceId, false, false))
	d.SetId(tDeploymentId)

	diag := updateWebDeploymentIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
}

// TestUnitResourceWebDeploymentIdentityResolutionUpdateRejectsSilentDowngrade covers the
// live-API behaviour found on inindca: contacts-service answers 200 but AND-s the
// automerge flags with the channel's capabilities, so a deployment without auth
// configured silently stores authenticated_web_messaging=false. Left undetected that is
// a plan that never converges, so update must fail instead.
func TestUnitResourceWebDeploymentIdentityResolutionUpdateRejectsSilentDowngrade(t *testing.T) {
	tDeploymentId := uuid.NewString()

	proxy := &webDeploymentIdentityResolutionProxy{}
	proxy.putWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, _ string, config platformclientv2.Deploymentidentityresolutionconfig) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		// Echo the request back with authenticated web messaging downgraded, exactly as
		// the platform does.
		stored := config
		stored.Automerge = &platformclientv2.Identityresolutionautomergeconfig{
			AuthenticatedWebMessaging: platformclientv2.Bool(false),
			WebTracking:               config.Automerge.WebTracking,
		}
		return &stored, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceWebDeploymentIdentityResolution().Schema, map[string]interface{}{
		"deployment_id":      tDeploymentId,
		"resolve_identities": true,
		"automerge_config": []interface{}{map[string]interface{}{
			"authenticated_web_messaging": true,
			"web_tracking":                false,
		}},
	})
	d.SetId(tDeploymentId)

	diag := updateWebDeploymentIdentityResolution(ctx, d, gcloud)
	assert.True(t, diag.HasError(), "a silently downgraded field must fail the apply")
	assert.Contains(t, diag[0].Summary, "authenticated_web_messaging")
	assert.Contains(t, diag[0].Detail, "authentication_settings")
}

func TestUnitResourceWebDeploymentIdentityResolutionRead(t *testing.T) {
	tDeploymentId := uuid.NewString()
	tDivisionId := uuid.NewString()
	tExternalSourceId := uuid.NewString()

	proxy := &webDeploymentIdentityResolutionProxy{}
	proxy.getWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, deploymentId string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tDeploymentId, deploymentId)
		return &platformclientv2.Deploymentidentityresolutionconfig{
			ResolveIdentities: platformclientv2.Bool(false),
			Division:          &platformclientv2.Writablestarrabledivision{Id: &tDivisionId},
			ExternalSource:    &platformclientv2.Identityresolutionexternalsource{Id: &tExternalSourceId},
			Automerge: &platformclientv2.Identityresolutionautomergeconfig{
				AuthenticatedWebMessaging: platformclientv2.Bool(true),
				WebTracking:               platformclientv2.Bool(false),
			},
		}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceWebDeploymentIdentityResolution().Schema,
		buildIdentityResolutionResourceMap(tDeploymentId, false, tDivisionId, tExternalSourceId, true, false))
	d.SetId(tDeploymentId)

	diag := readWebDeploymentIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tDeploymentId, d.Get("deployment_id").(string))
	assert.Equal(t, false, d.Get("resolve_identities").(bool))
	assert.Equal(t, tDivisionId, d.Get("division_id").(string))
	assert.Equal(t, tExternalSourceId, d.Get("external_source_id").(string))

	blocks := d.Get("automerge_config").([]interface{})
	assert.Len(t, blocks, 1)
	block := blocks[0].(map[string]interface{})
	assert.Equal(t, true, block["authenticated_web_messaging"])
	assert.Equal(t, false, block["web_tracking"])
}

// TestUnitResourceWebDeploymentIdentityResolutionReadClearsUnsetFields covers the
// stale-state case: a division / external source that disappears from the API must be
// cleared from state rather than left behind, or the next plan drifts forever.
func TestUnitResourceWebDeploymentIdentityResolutionReadClearsUnsetFields(t *testing.T) {
	tDeploymentId := uuid.NewString()
	star := "*"

	tests := []struct {
		name string
		api  *platformclientv2.Deploymentidentityresolutionconfig
	}{
		{
			name: "division and external source absent",
			api: &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
			},
		},
		{
			name: "division returned as the unassigned star sentinel",
			api: &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
				Division:          &platformclientv2.Writablestarrabledivision{Id: &star},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bypassConsistencyChecker(t)

			proxy := &webDeploymentIdentityResolutionProxy{}
			proxy.getWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, _ string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
				return test.api, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
			}

			internalProxy = proxy
			defer func() { internalProxy = nil }()

			ctx := context.Background()
			gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

			// Prior state carries a division, an external source and an automerge block.
			d := schema.TestResourceDataRaw(t, ResourceWebDeploymentIdentityResolution().Schema,
				buildIdentityResolutionResourceMap(tDeploymentId, true, uuid.NewString(), uuid.NewString(), true, true))
			d.SetId(tDeploymentId)

			diag := readWebDeploymentIdentityResolution(ctx, d, gcloud)
			assert.False(t, diag.HasError(), diag)
			assert.Equal(t, "", d.Get("division_id").(string), "division_id must be cleared from state")
			assert.Equal(t, "", d.Get("external_source_id").(string), "external_source_id must be cleared from state")

			// The block was already declared, so it stays in state with automerging off
			// rather than disappearing (that is what keeps the plan convergent).
			blocks := d.Get("automerge_config").([]interface{})
			assert.Len(t, blocks, 1)
			block := blocks[0].(map[string]interface{})
			assert.Equal(t, false, block["authenticated_web_messaging"])
			assert.Equal(t, false, block["web_tracking"])
		})
	}
}

// TestUnitResourceWebDeploymentIdentityResolutionReadOmitsUndeclaredAutomerge is the
// other half of the sticky flatten: a resource with no automerge block must not grow
// one just because the API always returns an all-false object.
func TestUnitResourceWebDeploymentIdentityResolutionReadOmitsUndeclaredAutomerge(t *testing.T) {
	bypassConsistencyChecker(t)

	tDeploymentId := uuid.NewString()

	proxy := &webDeploymentIdentityResolutionProxy{}
	proxy.getWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, _ string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		return &platformclientv2.Deploymentidentityresolutionconfig{
			ResolveIdentities: platformclientv2.Bool(false),
			Automerge: &platformclientv2.Identityresolutionautomergeconfig{
				AuthenticatedWebMessaging: platformclientv2.Bool(false),
				WebTracking:               platformclientv2.Bool(false),
			},
		}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceWebDeploymentIdentityResolution().Schema, map[string]interface{}{
		"deployment_id":      tDeploymentId,
		"resolve_identities": false,
	})
	d.SetId(tDeploymentId)

	diag := readWebDeploymentIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Empty(t, d.Get("automerge_config").([]interface{}), "an undeclared all-false automerge block must stay out of state")
}

func TestUnitResourceWebDeploymentIdentityResolutionReadParentGone(t *testing.T) {
	tDeploymentId := uuid.NewString()

	proxy := &webDeploymentIdentityResolutionProxy{}
	proxy.getWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, _ string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("not found")
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceWebDeploymentIdentityResolution().Schema,
		buildIdentityResolutionResourceMap(tDeploymentId, true, "", "", false, false))
	d.SetId(tDeploymentId)

	diag := readWebDeploymentIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, "", d.Id(), "a 404 on the parent deployment must remove the resource from state")
}

// TestUnitResourceWebDeploymentIdentityResolutionDelete asserts destroy PUTs the platform
// default: resolve_identities back on, division and external source unset, and automerge
// explicitly switched off for every channel rather than left to PUT-merge semantics.
func TestUnitResourceWebDeploymentIdentityResolutionDelete(t *testing.T) {
	tDeploymentId := uuid.NewString()
	putCalled := false

	proxy := &webDeploymentIdentityResolutionProxy{}
	proxy.getWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, _ string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		return &platformclientv2.Deploymentidentityresolutionconfig{
			ResolveIdentities: platformclientv2.Bool(false),
		}, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	proxy.putWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, deploymentId string, config platformclientv2.Deploymentidentityresolutionconfig) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		putCalled = true
		assert.Equal(t, tDeploymentId, deploymentId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, true, *config.ResolveIdentities)
		assert.Nil(t, config.Division, "destroy must not send a division")
		assert.Nil(t, config.ExternalSource, "destroy must not send an external source")
		assert.NotNil(t, config.Automerge, "destroy must explicitly reset automerge")
		assert.Equal(t, false, *config.Automerge.AuthenticatedWebMessaging)
		assert.Equal(t, false, *config.Automerge.WebTracking)

		return &config, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceWebDeploymentIdentityResolution().Schema,
		buildIdentityResolutionResourceMap(tDeploymentId, false, uuid.NewString(), uuid.NewString(), true, true))
	d.SetId(tDeploymentId)

	diag := deleteWebDeploymentIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.True(t, putCalled, "destroy must PUT the default config")
}

func TestUnitResourceWebDeploymentIdentityResolutionDeleteParentGone(t *testing.T) {
	tDeploymentId := uuid.NewString()

	proxy := &webDeploymentIdentityResolutionProxy{}
	proxy.getWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, _ string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("not found")
	}
	proxy.putWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, _ string, config platformclientv2.Deploymentidentityresolutionconfig) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		t.Fatal("destroy must not PUT when the parent deployment is already gone")
		return &config, nil, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceWebDeploymentIdentityResolution().Schema,
		buildIdentityResolutionResourceMap(tDeploymentId, true, "", "", false, false))
	d.SetId(tDeploymentId)

	diag := deleteWebDeploymentIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
}

func TestUnitGetAllWebDeploymentIdentityResolutions(t *testing.T) {
	defaultId, defaultName := uuid.NewString(), "Default Deployment"
	divisionId, divisionName := uuid.NewString(), "Division Deployment"
	automergeId, automergeName := uuid.NewString(), "Automerge Deployment"
	notFoundId, notFoundName := uuid.NewString(), "Not Found Deployment"

	tDivisionId := uuid.NewString()

	deploymentsProxy := &webDeploymentsDeployment.WebDeploymentsProxy{}
	deploymentsProxy.GetAllWebDeploymentsAttr = func(_ context.Context, _ *webDeploymentsDeployment.WebDeploymentsProxy) (*platformclientv2.Expandablewebdeploymententitylisting, *platformclientv2.APIResponse, error) {
		return &platformclientv2.Expandablewebdeploymententitylisting{
			Entities: &[]platformclientv2.Expandablewebdeployment{
				{Id: &defaultId, Name: &defaultName},
				{Id: &divisionId, Name: &divisionName},
				{Id: &automergeId, Name: &automergeName},
				{Id: &notFoundId, Name: &notFoundName},
			},
		}, nil, nil
	}

	proxy := &webDeploymentIdentityResolutionProxy{webDeploymentsProxy: deploymentsProxy}
	proxy.getWebDeploymentIdentityResolutionAttr = func(_ context.Context, _ *webDeploymentIdentityResolutionProxy, deploymentId string) (*platformclientv2.Deploymentidentityresolutionconfig, *platformclientv2.APIResponse, error) {
		okResp := &platformclientv2.APIResponse{StatusCode: http.StatusOK}
		switch deploymentId {
		case defaultId:
			// The platform default every never-configured deployment returns.
			return &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
				Automerge: &platformclientv2.Identityresolutionautomergeconfig{
					AuthenticatedWebMessaging: platformclientv2.Bool(false),
					WebTracking:               platformclientv2.Bool(false),
				},
			}, okResp, nil
		case divisionId:
			return &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
				Division:          &platformclientv2.Writablestarrabledivision{Id: &tDivisionId},
			}, okResp, nil
		case automergeId:
			return &platformclientv2.Deploymentidentityresolutionconfig{
				ResolveIdentities: platformclientv2.Bool(true),
				Automerge: &platformclientv2.Identityresolutionautomergeconfig{
					AuthenticatedWebMessaging: platformclientv2.Bool(false),
					WebTracking:               platformclientv2.Bool(true),
				},
			}, okResp, nil
		case notFoundId:
			return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("not found")
		default:
			t.Fatalf("unexpected deployment id %s", deploymentId)
			return nil, nil, nil
		}
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	resources, diag := getAllWebDeploymentIdentityResolutions(context.Background(), &platformclientv2.Configuration{})

	assert.False(t, diag.HasError())
	assert.Len(t, resources, 2, "only non-default configs are exported")
	assert.Contains(t, resources, divisionId)
	assert.Contains(t, resources, automergeId)
	assert.NotContains(t, resources, defaultId, "default configs must be skipped")
	assert.NotContains(t, resources, notFoundId, "per-deployment 404s must be skipped")
	assert.Equal(t, divisionName+"-identity-resolution", resources[divisionId].BlockLabel)
}

func TestUnitGetAllWebDeploymentIdentityResolutionsListError(t *testing.T) {
	deploymentsProxy := &webDeploymentsDeployment.WebDeploymentsProxy{}
	deploymentsProxy.GetAllWebDeploymentsAttr = func(_ context.Context, _ *webDeploymentsDeployment.WebDeploymentsProxy) (*platformclientv2.Expandablewebdeploymententitylisting, *platformclientv2.APIResponse, error) {
		return nil, nil, fmt.Errorf("mock list error")
	}

	internalProxy = &webDeploymentIdentityResolutionProxy{webDeploymentsProxy: deploymentsProxy}
	defer func() { internalProxy = nil }()

	resources, diag := getAllWebDeploymentIdentityResolutions(context.Background(), &platformclientv2.Configuration{})

	assert.True(t, diag.HasError())
	assert.Nil(t, resources)
}

// buildIdentityResolutionResourceMap builds resource data. An empty divisionId or
// externalSourceId is omitted; declareAutomerge controls whether the block is present.
func buildIdentityResolutionResourceMap(deploymentId string, resolveIdentities bool, divisionId, externalSourceId string, declareAutomerge, webTracking bool) map[string]interface{} {
	resourceMap := map[string]interface{}{
		"deployment_id":      deploymentId,
		"resolve_identities": resolveIdentities,
	}
	if divisionId != "" {
		resourceMap["division_id"] = divisionId
	}
	if externalSourceId != "" {
		resourceMap["external_source_id"] = externalSourceId
	}
	if declareAutomerge {
		resourceMap["automerge_config"] = []interface{}{
			map[string]interface{}{
				"authenticated_web_messaging": !webTracking,
				"web_tracking":                webTracking,
			},
		}
	}

	return resourceMap
}
