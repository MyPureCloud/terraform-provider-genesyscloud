package architect_ivr_identity_resolution

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	architectIvr "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceArchitectIvrIdentityResolutionUpdate(t *testing.T) {
	tIvrId := uuid.NewString()

	proxy := &architectIvrIdentityResolutionProxy{}
	proxy.putArchitectIvrIdentityResolutionAttr = func(ctx context.Context, p *architectIvrIdentityResolutionProxy, ivrId string, config platformclientv2.Ivridentityresolutionconfig) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tIvrId, ivrId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, false, *config.ResolveIdentities)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &config, &apiResponse, nil
	}

	proxy.getArchitectIvrIdentityResolutionAttr = func(ctx context.Context, p *architectIvrIdentityResolutionProxy, ivrId string) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Ivridentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceArchitectIvrIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tIvrId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tIvrId)

	diag := updateArchitectIvrIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tIvrId, d.Id())
	assert.Equal(t, tIvrId, d.Get("ivr_id").(string))
}

func TestUnitResourceArchitectIvrIdentityResolutionRead(t *testing.T) {
	tIvrId := uuid.NewString()
	tDivisionId := uuid.NewString()

	proxy := &architectIvrIdentityResolutionProxy{}
	proxy.getArchitectIvrIdentityResolutionAttr = func(ctx context.Context, p *architectIvrIdentityResolutionProxy, ivrId string) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error) {
		resolveIdentities := false
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Ivridentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
			Division: &platformclientv2.Writablestarrabledivision{
				Id: &tDivisionId,
			},
		}, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceArchitectIvrIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tIvrId, false, tDivisionId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tIvrId)

	diag := readArchitectIvrIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
	assert.Equal(t, tIvrId, d.Id())
	assert.Equal(t, tIvrId, d.Get("ivr_id").(string))

	assert.Equal(t, false, d.Get("resolve_identities"))
	assert.Equal(t, tDivisionId, d.Get("division_id"))
}

func TestUnitResourceArchitectIvrIdentityResolutionDelete(t *testing.T) {
	tIvrId := uuid.NewString()

	ivrProxy := &architectIvr.ArchitectIvrProxy{}
	ivrProxy.GetArchitectIvrAttr = func(ctx context.Context, p *architectIvr.ArchitectIvrProxy, ivrId string) (*platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &platformclientv2.Ivr{Id: &tIvrId}, &apiResponse, nil
	}

	proxy := &architectIvrIdentityResolutionProxy{
		architectIvrProxy: ivrProxy,
	}
	proxy.putArchitectIvrIdentityResolutionAttr = func(ctx context.Context, p *architectIvrIdentityResolutionProxy, ivrId string, config platformclientv2.Ivridentityresolutionconfig) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tIvrId, ivrId)
		assert.NotNil(t, config.ResolveIdentities)
		assert.Equal(t, true, *config.ResolveIdentities)
		assert.Nil(t, config.Division)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return &config, &apiResponse, nil
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	resourceSchema := ResourceArchitectIvrIdentityResolution().Schema
	resourceDataMap := buildIdentityResolutionResourceMap(tIvrId, false, "")

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tIvrId)

	diag := deleteArchitectIvrIdentityResolution(ctx, d, gcloud)
	assert.False(t, diag.HasError(), diag)
}

func TestUnitGetAllArchitectIvrIdentityResolution(t *testing.T) {
	defaultIvrId := uuid.NewString()
	defaultIvrName := "Default IVR"
	customIvrId := uuid.NewString()
	customIvrName := "Custom IVR"
	notFoundIvrId := uuid.NewString()
	notFoundIvrName := "Not Found IVR"

	resolveTrue := true
	resolveFalse := false
	divisionId := uuid.NewString()

	ivrProxy := &architectIvr.ArchitectIvrProxy{}
	ivrProxy.GetAllArchitectIvrsAttr = func(_ context.Context, _ *architectIvr.ArchitectIvrProxy, _ string) (*[]platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		return &[]platformclientv2.Ivr{
			{Id: &defaultIvrId, Name: &defaultIvrName},
			{Id: &customIvrId, Name: &customIvrName},
			{Id: &notFoundIvrId, Name: &notFoundIvrName},
		}, nil, nil
	}

	proxy := &architectIvrIdentityResolutionProxy{
		architectIvrProxy: ivrProxy,
	}
	proxy.getArchitectIvrIdentityResolutionAttr = func(ctx context.Context, p *architectIvrIdentityResolutionProxy, ivrId string) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error) {
		switch ivrId {
		case defaultIvrId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Ivridentityresolutionconfig{
				ResolveIdentities: &resolveTrue,
			}, &apiResponse, nil
		case customIvrId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
			return &platformclientv2.Ivridentityresolutionconfig{
				ResolveIdentities: &resolveFalse,
				Division: &platformclientv2.Writablestarrabledivision{
					Id: &divisionId,
				},
			}, &apiResponse, nil
		case notFoundIvrId:
			apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusNotFound}
			return nil, &apiResponse, fmt.Errorf("not found")
		default:
			t.Fatalf("unexpected IVR ID %s", ivrId)
			return nil, nil, nil
		}
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllArchitectIvrIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.False(t, diag.HasError())
	assert.Len(t, resources, 1)
	assert.Contains(t, resources, customIvrId)
	assert.Equal(t, customIvrName+"-identity-resolution", resources[customIvrId].BlockLabel)
}

func TestUnitGetAllArchitectIvrIdentityResolutionListError(t *testing.T) {
	ivrProxy := &architectIvr.ArchitectIvrProxy{}
	ivrProxy.GetAllArchitectIvrsAttr = func(_ context.Context, _ *architectIvr.ArchitectIvrProxy, _ string) (*[]platformclientv2.Ivr, *platformclientv2.APIResponse, error) {
		return nil, nil, fmt.Errorf("mock list error")
	}

	proxy := &architectIvrIdentityResolutionProxy{
		architectIvrProxy: ivrProxy,
	}

	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	resources, diag := getAllArchitectIvrIdentityResolution(ctx, &platformclientv2.Configuration{})

	assert.True(t, diag.HasError())
	assert.Nil(t, resources)
}

func buildIdentityResolutionResourceMap(ivrId string, resolveIdentities bool, divisionId string) map[string]interface{} {
	return map[string]interface{}{
		"ivr_id":             ivrId,
		"resolve_identities": resolveIdentities,
		"division_id":        divisionId,
	}
}
