package quality_miner

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func generateMinerData(id string) platformclientv2.Miner {
	return platformclientv2.Miner{
		Id:        &id,
		Name:      platformclientv2.String("Test Miner"),
		MinerType: platformclientv2.String("Topic"),
		Language:  platformclientv2.String("en-US"),
	}
}

func buildMinerDataMap(miner platformclientv2.Miner) map[string]interface{} {
	return map[string]interface{}{
		"name":       *miner.Name,
		"miner_type": *miner.MinerType,
		"language":   *miner.Language,
	}
}

func TestUnitResourceQualityMinerCreate(t *testing.T) {
	tId := uuid.NewString()
	testMiner := generateMinerData(tId)

	var capturedCreateBody *platformclientv2.Miner
	proxy := &qualityMinerProxy{}
	proxy.createQualityMinerAttr = func(ctx context.Context, p *qualityMinerProxy, miner *platformclientv2.Miner) (*platformclientv2.Miner, *platformclientv2.APIResponse, error) {
		capturedCreateBody = miner
		created := testMiner
		return &created, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getQualityMinerByIdAttr = func(ctx context.Context, p *qualityMinerProxy, id string) (*platformclientv2.Miner, *platformclientv2.APIResponse, error) {
		m := testMiner
		return &m, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceQualityMiner().Schema, buildMinerDataMap(testMiner))

	diagErr := createQualityMiner(ctx, d, gcloud)
	assert.False(t, diagErr.HasError())
	assert.Equal(t, tId, d.Id())

	assert.NotNil(t, capturedCreateBody)
	assert.Equal(t, "Test Miner", *capturedCreateBody.Name)
	assert.Equal(t, "Topic", *capturedCreateBody.MinerType)
	assert.Equal(t, "en-US", *capturedCreateBody.Language)
	assert.Nil(t, capturedCreateBody.Id)
}

func TestUnitResourceQualityMinerCreateError(t *testing.T) {
	testMiner := generateMinerData(uuid.NewString())

	proxy := &qualityMinerProxy{}
	proxy.createQualityMinerAttr = func(ctx context.Context, p *qualityMinerProxy, miner *platformclientv2.Miner) (*platformclientv2.Miner, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusBadRequest}, fmt.Errorf("bad request")
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceQualityMiner().Schema, buildMinerDataMap(testMiner))

	diagErr := createQualityMiner(ctx, d, gcloud)
	assert.True(t, diagErr.HasError())
	assert.Empty(t, d.Id())
}

func TestUnitResourceQualityMinerRead(t *testing.T) {
	tId := uuid.NewString()
	testMiner := generateMinerData(tId)

	proxy := &qualityMinerProxy{}
	proxy.getQualityMinerByIdAttr = func(ctx context.Context, p *qualityMinerProxy, id string) (*platformclientv2.Miner, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tId, id)
		m := testMiner
		return &m, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceQualityMiner().Schema, map[string]interface{}{})
	d.SetId(tId)

	diagErr := readQualityMiner(ctx, d, gcloud)
	assert.False(t, diagErr.HasError())
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, "Test Miner", d.Get("name").(string))
	assert.Equal(t, "Topic", d.Get("miner_type").(string))
	assert.Equal(t, "en-US", d.Get("language").(string))
}

func TestUnitResourceQualityMinerRead404RemovesFromState(t *testing.T) {
	tId := uuid.NewString()

	proxy := &qualityMinerProxy{}
	proxy.getQualityMinerByIdAttr = func(ctx context.Context, p *qualityMinerProxy, id string) (*platformclientv2.Miner, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusNotFound}, fmt.Errorf("not found")
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceQualityMiner().Schema, map[string]interface{}{})
	d.SetId(tId)

	diagErr := readQualityMiner(ctx, d, gcloud)
	assert.False(t, diagErr.HasError())
	assert.Empty(t, d.Id())
}

func TestUnitResourceQualityMinerReadError(t *testing.T) {
	t.Setenv("GENESYSCLOUD_CUSTOM_RETRY_TIMEOUT", "0s")
	tId := uuid.NewString()

	proxy := &qualityMinerProxy{}
	proxy.getQualityMinerByIdAttr = func(ctx context.Context, p *qualityMinerProxy, id string) (*platformclientv2.Miner, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("server error")
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceQualityMiner().Schema, map[string]interface{}{})
	d.SetId(tId)

	diagErr := readQualityMiner(ctx, d, gcloud)
	assert.True(t, diagErr.HasError())
}

func TestUnitResourceQualityMinerDelete(t *testing.T) {
	tId := uuid.NewString()
	testMiner := generateMinerData(tId)

	deleteCalled := false
	proxy := &qualityMinerProxy{}
	proxy.deleteQualityMinerAttr = func(ctx context.Context, p *qualityMinerProxy, id string) (*platformclientv2.APIResponse, error) {
		deleteCalled = true
		assert.Equal(t, tId, id)
		return &platformclientv2.APIResponse{StatusCode: http.StatusNoContent}, nil
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceQualityMiner().Schema, buildMinerDataMap(testMiner))
	d.SetId(tId)

	diagErr := deleteQualityMiner(ctx, d, gcloud)
	assert.False(t, diagErr.HasError())
	assert.True(t, deleteCalled)
}

func TestUnitResourceQualityMinerDeleteError(t *testing.T) {
	tId := uuid.NewString()
	testMiner := generateMinerData(tId)

	proxy := &qualityMinerProxy{}
	proxy.deleteQualityMinerAttr = func(ctx context.Context, p *qualityMinerProxy, id string) (*platformclientv2.APIResponse, error) {
		return &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("server error")
	}
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, ResourceQualityMiner().Schema, buildMinerDataMap(testMiner))
	d.SetId(tId)

	diagErr := deleteQualityMiner(ctx, d, gcloud)
	assert.True(t, diagErr.HasError())
}

func TestUnitDataSourceQualityMinerRead(t *testing.T) {
	tId := uuid.NewString()
	testMiner := generateMinerData(tId)

	proxy := &qualityMinerProxy{}
	proxy.getAllQualityMinersAttr = func(ctx context.Context, p *qualityMinerProxy) (*[]platformclientv2.Miner, *platformclientv2.APIResponse, error) {
		miners := []platformclientv2.Miner{testMiner}
		return &miners, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getQualityMinerIdByNameAttr = getQualityMinerIdByNameFn
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, DataSourceQualityMiner().Schema, map[string]interface{}{
		"name":       "Test Miner",
		"miner_type": "Topic",
	})

	diagErr := dataSourceQualityMinerRead(ctx, d, gcloud)
	assert.False(t, diagErr.HasError())
	assert.Equal(t, tId, d.Id())
}

func TestUnitDataSourceQualityMinerReadFiltersByType(t *testing.T) {
	topicId := uuid.NewString()
	intentId := uuid.NewString()
	topicMiner := platformclientv2.Miner{Id: &topicId, Name: platformclientv2.String("Shared"), MinerType: platformclientv2.String("Topic")}
	intentMiner := platformclientv2.Miner{Id: &intentId, Name: platformclientv2.String("OnlyIntent"), MinerType: platformclientv2.String("Intent")}

	proxy := &qualityMinerProxy{}
	proxy.getAllQualityMinersAttr = func(ctx context.Context, p *qualityMinerProxy) (*[]platformclientv2.Miner, *platformclientv2.APIResponse, error) {
		miners := []platformclientv2.Miner{topicMiner, intentMiner}
		return &miners, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getQualityMinerIdByNameAttr = getQualityMinerIdByNameFn
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, DataSourceQualityMiner().Schema, map[string]interface{}{
		"name":       "OnlyIntent",
		"miner_type": "Intent",
	})

	diagErr := dataSourceQualityMinerRead(ctx, d, gcloud)
	assert.False(t, diagErr.HasError())
	assert.Equal(t, intentId, d.Id())
}

func TestUnitDataSourceQualityMinerReadNotFound(t *testing.T) {
	t.Setenv("GENESYSCLOUD_CUSTOM_RETRY_TIMEOUT", "0s")

	proxy := &qualityMinerProxy{}
	proxy.getAllQualityMinersAttr = func(ctx context.Context, p *qualityMinerProxy) (*[]platformclientv2.Miner, *platformclientv2.APIResponse, error) {
		miners := []platformclientv2.Miner{}
		return &miners, &platformclientv2.APIResponse{StatusCode: http.StatusOK}, nil
	}
	proxy.getQualityMinerIdByNameAttr = getQualityMinerIdByNameFn
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, DataSourceQualityMiner().Schema, map[string]interface{}{
		"name": "Nope",
	})

	diagErr := dataSourceQualityMinerRead(ctx, d, gcloud)
	assert.True(t, diagErr.HasError())
}

func TestUnitDataSourceQualityMinerReadGetAllError(t *testing.T) {
	t.Setenv("GENESYSCLOUD_CUSTOM_RETRY_TIMEOUT", "0s")

	proxy := &qualityMinerProxy{}
	proxy.getAllQualityMinersAttr = func(ctx context.Context, p *qualityMinerProxy) (*[]platformclientv2.Miner, *platformclientv2.APIResponse, error) {
		return nil, &platformclientv2.APIResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("server error")
	}
	proxy.getQualityMinerIdByNameAttr = getQualityMinerIdByNameFn
	internalProxy = proxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}
	d := schema.TestResourceDataRaw(t, DataSourceQualityMiner().Schema, map[string]interface{}{
		"name": "Anything",
	})

	diagErr := dataSourceQualityMinerRead(ctx, d, gcloud)
	assert.True(t, diagErr.HasError())
}
