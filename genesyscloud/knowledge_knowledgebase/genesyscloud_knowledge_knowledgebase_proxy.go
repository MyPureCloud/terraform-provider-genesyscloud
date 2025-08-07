package knowledge_knowledgebase

import (
	"context"
	"fmt"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

var internalProxy *knowledgebaseProxy

type getAllKnowledgebaseEntitiesFunc func(ctx context.Context, p *knowledgebaseProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)
type getKnowledgebaseByIdFunc func(ctx context.Context, p *knowledgebaseProxy, knowledgebaseId string) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)
type createKnowledgebaseFunc func(ctx context.Context, p *knowledgebaseProxy, knowledgebaseRequest *platformclientv2.Knowledgebasecreaterequest) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)
type updateKnowledgebaseFunc func(ctx context.Context, p *knowledgebaseProxy, knowledgebaseId string, updateBody *platformclientv2.Knowledgebaseupdaterequest) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)
type deleteKnowledgebaseFunc func(ctx context.Context, p *knowledgebaseProxy, knowledgebaseId string) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)

type knowledgebaseProxy struct {
	clientConfig                    *platformclientv2.Configuration
	KnowledgeApi                    *platformclientv2.KnowledgeApi
	getAllKnowledgebaseEntitiesAttr getAllKnowledgebaseEntitiesFunc
	getKnowledgebaseByIdAttr        getKnowledgebaseByIdFunc
	createKnowledgebaseAttr         createKnowledgebaseFunc
	updateKnowledgebaseAttr         updateKnowledgebaseFunc
	deleteKnowledgebaseAttr         deleteKnowledgebaseFunc
	knowledgebaseCache              rc.CacheInterface[platformclientv2.Knowledgebase]
}

func newKnowledgebaseProxy(clientConfig *platformclientv2.Configuration) *knowledgebaseProxy {
	api := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)
	knowledgebaseCache := rc.NewResourceCache[platformclientv2.Knowledgebase]()
	return &knowledgebaseProxy{
		clientConfig:                    clientConfig,
		KnowledgeApi:                    api,
		getAllKnowledgebaseEntitiesAttr: getAllKnowledgebaseEntitiesFn,
		getKnowledgebaseByIdAttr:        getKnowledgebaseByIdFn,
		createKnowledgebaseAttr:         createKnowledgebaseFn,
		updateKnowledgebaseAttr:         updateKnowledgebaseFn,
		deleteKnowledgebaseAttr:         deleteKnowledgebaseFn,
		knowledgebaseCache:              knowledgebaseCache,
	}
}

func GetKnowledgebaseProxy(clientConfig *platformclientv2.Configuration) *knowledgebaseProxy {
	if internalProxy == nil {
		internalProxy = newKnowledgebaseProxy(clientConfig)
	}

	return internalProxy
}

// getAllKnowledgebaseEntities retrieves all Genesys Cloud knowledgebases
func (p *knowledgebaseProxy) getAllKnowledgebaseEntities(ctx context.Context, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.getAllKnowledgebaseEntitiesAttr(ctx, p, published)
}

// getKnowledgebaseById retrieves knowledgebase by its id
func (p *knowledgebaseProxy) getKnowledgebaseById(ctx context.Context, knowledgebaseId string) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.getKnowledgebaseByIdAttr(ctx, p, knowledgebaseId)
}

// createKnowledgebase creates a knowledgebase in Genesys Cloud
func (p *knowledgebaseProxy) createKnowledgebase(ctx context.Context, knowledgebaseRequest *platformclientv2.Knowledgebasecreaterequest) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.createKnowledgebaseAttr(ctx, p, knowledgebaseRequest)
}

// updateKnowledgebase updates a knowledgebase in Genesys Cloud
func (p *knowledgebaseProxy) updateKnowledgebase(ctx context.Context, knowledgebaseId string, updateBody *platformclientv2.Knowledgebaseupdaterequest) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.updateKnowledgebaseAttr(ctx, p, knowledgebaseId, updateBody)
}

// deleteKnowledgebase deletes a knowledgebase in Genesys Cloud
func (p *knowledgebaseProxy) deleteKnowledgebase(ctx context.Context, knowledgebaseId string) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.deleteKnowledgebaseAttr(ctx, p, knowledgebaseId)
}

func getAllKnowledgebaseEntitiesFn(ctx context.Context, p *knowledgebaseProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	var (
		after    string
		resp     *platformclientv2.APIResponse
		entities []platformclientv2.Knowledgebase
	)

	const pageSize = 100
	for {
		knowledgeBases, resp, err := p.KnowledgeApi.GetKnowledgeKnowledgebases("", after, "", fmt.Sprintf("%v", pageSize), "", "", published, "", "")
		if err != nil {
			return nil, resp, err
		}

		if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeBases.Entities...)

		if knowledgeBases.NextUri == nil || *knowledgeBases.NextUri == "" {
			break
		}

		previousAfter := after // Store current token before getting next one
		after, err = util.GetQueryParamValueFromUri(*knowledgeBases.NextUri, "after")
		if err != nil {
			return nil, resp, err
		}
		if after == "" || after == previousAfter {
			break
		}
	}

	return &entities, resp, nil
}

func getKnowledgebaseByIdFn(ctx context.Context, p *knowledgebaseProxy, knowledgebaseId string) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.GetKnowledgeKnowledgebase(knowledgebaseId)
}

func createKnowledgebaseFn(ctx context.Context, p *knowledgebaseProxy, knowledgebaseRequest *platformclientv2.Knowledgebasecreaterequest) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.PostKnowledgeKnowledgebases(*knowledgebaseRequest)
}

func updateKnowledgebaseFn(ctx context.Context, p *knowledgebaseProxy, knowledgebaseId string, updateBody *platformclientv2.Knowledgebaseupdaterequest) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.PatchKnowledgeKnowledgebase(knowledgebaseId, *updateBody)
}

func deleteKnowledgebaseFn(ctx context.Context, p *knowledgebaseProxy, knowledgebaseId string) (*platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	delete, resp, err := p.KnowledgeApi.DeleteKnowledgeKnowledgebase(knowledgebaseId)
	if err != nil {
		return delete, resp, err
	}
	rc.DeleteCacheItem(p.knowledgebaseCache, knowledgebaseId)
	return delete, resp, nil
}
