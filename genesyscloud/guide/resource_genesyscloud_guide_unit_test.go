package guide

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceGuideCreate(t *testing.T) {
	tId := uuid.NewString()
	tName := "Test Guide"
	tSource := "Manual"

	testGuide := &Guide{
		Id:     &tId,
		Name:   &tName,
		Source: &tSource,
		Status: platformclientv2.String("Draft"),
	}

	var guideProxyObj = &guideProxy{}

	guideProxyObj.getGuideByIdAttr = func(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		return testGuide, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	guideProxyObj.createGuideAttr = func(ctx context.Context, p *guideProxy, guide *CreateGuide) (*Guide, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tName, *guide.Name)
		assert.Equal(t, tSource, *guide.Source)
		return testGuide, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = guideProxyObj
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceGuide().Schema
	resourceDataMap := map[string]interface{}{
		"name":   tName,
		"source": tSource,
	}

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	diag := createGuide(ctx, d, gcloud)

	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitResourceGuideRead(t *testing.T) {
	tId := uuid.NewString()
	tName := "Test Guide"
	tSource := "Manual"
	tStatus := "Draft"
	version := "1.0"
	testGuide := &Guide{
		Id:     &tId,
		Name:   &tName,
		Source: &tSource,
		Status: &tStatus,
		LatestSavedVersion: &GuideVersionRef{
			Version: &version,
		},
		LatestProductionReadyVersion: &GuideVersionRef{
			Version: &version,
		},
	}

	guideProxyObj := &guideProxy{}
	guideProxyObj.getGuideByIdAttr = func(ctx context.Context, p *guideProxy, id string) (*Guide, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		return testGuide, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}

	internalProxy = guideProxyObj
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	resourceSchema := ResourceGuide().Schema
	resourceDataMap := map[string]interface{}{
		"name":   tName,
		"source": tSource,
	}

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readGuide(ctx, d, gcloud)

	assert.Equal(t, false, diag.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tName, d.Get("name").(string))
	assert.Equal(t, tSource, d.Get("source").(string))
	assert.Equal(t, tStatus, d.Get("status").(string))
	assert.Equal(t, version, d.Get("latest_saved_version").(string))
	assert.Equal(t, version, d.Get("latest_production_ready_version").(string))
}
