package knowledge_label

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var internalProxy *knowledgeLabelProxy

type GetAllKnowledgebaseEntitiesFunc func(ctx context.Context, p *knowledgeLabelProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)
type GetAllKnowledgeLabelEntitiesFunc func(ctx context.Context, p *knowledgeLabelProxy, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Labelresponse, *platformclientv2.APIResponse, error)
type getKnowledgeLabelFunc func(ctx context.Context, p *knowledgeLabelProxy, knowledgeBaseId string, labelId string) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error)
type createKnowledgeLabelFunc func(ctx context.Context, p *knowledgeLabelProxy, knowledgeBaseId string, body *platformclientv2.Labelcreaterequest) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error)
type deleteKnowledgeLabelFunc func(ctx context.Context, p *knowledgeLabelProxy, knowledgeBaseId string, knowledgeLabelId string) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error)
type updateKnowledgeLabelFunc func(ctx context.Context, p *knowledgeLabelProxy, knowledgeBaseId string, knowledgeLabelId string, body *platformclientv2.Labelupdaterequest) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error)

type knowledgeLabelProxy struct {
	clientConfig                     *platformclientv2.Configuration
	KnowledgeApi                     *platformclientv2.KnowledgeApi
	getKnowledgeLabelAttr            getKnowledgeLabelFunc
	GetAllKnowledgebaseEntitiesAttr  GetAllKnowledgebaseEntitiesFunc
	GetAllKnowledgeLabelEntitiesAttr GetAllKnowledgeLabelEntitiesFunc
	createKnowledgeLabelAttr         createKnowledgeLabelFunc
	deleteKnowledgeLabelAttr         deleteKnowledgeLabelFunc
	updateKnowledgeLabelAttr         updateKnowledgeLabelFunc
	knowledgeLabelCache              rc.CacheInterface[platformclientv2.Labelresponse]
}

func newKnowledgeLabelProxy(clientConfig *platformclientv2.Configuration) *knowledgeLabelProxy {
	api := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)
	knowledgeLabelCache := rc.NewResourceCache[platformclientv2.Labelresponse]()
	return &knowledgeLabelProxy{
		clientConfig:                     clientConfig,
		KnowledgeApi:                     api,
		getKnowledgeLabelAttr:            getKnowledgeLabelFn,
		GetAllKnowledgebaseEntitiesAttr:  GetAllKnowledgebaseEntitiesFn,
		GetAllKnowledgeLabelEntitiesAttr: GetAllKnowledgeLabelEntitiesFn,
		createKnowledgeLabelAttr:         createKnowledgeLabelFn,
		deleteKnowledgeLabelAttr:         deleteKnowledgeLabelFn,
		updateKnowledgeLabelAttr:         updateKnowledgeLabelFn,
		knowledgeLabelCache:              knowledgeLabelCache,
	}
}

func GetKnowledgeLabelProxy(clientConfig *platformclientv2.Configuration) *knowledgeLabelProxy {
	if internalProxy == nil {
		internalProxy = newKnowledgeLabelProxy(clientConfig)
	}

	return internalProxy
}

func (p *knowledgeLabelProxy) getKnowledgeLabel(ctx context.Context, knowledgeBaseId string, labelId string) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeLabelAttr(ctx, p, knowledgeBaseId, labelId)
}

func (p *knowledgeLabelProxy) createKnowledgeLabel(ctx context.Context, knowledgeBaseId string, body *platformclientv2.Labelcreaterequest) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	return p.createKnowledgeLabelAttr(ctx, p, knowledgeBaseId, body)
}

func (p *knowledgeLabelProxy) updateKnowledgeLabel(ctx context.Context, knowledgeBaseId string, knowledgeLabelId string, body *platformclientv2.Labelupdaterequest) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	return p.updateKnowledgeLabelAttr(ctx, p, knowledgeBaseId, knowledgeLabelId, body)
}

func (p *knowledgeLabelProxy) deleteKnowledgeLabel(ctx context.Context, knowledgeBaseId string, knowledgeLabelId string) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	return p.deleteKnowledgeLabelAttr(ctx, p, knowledgeBaseId, knowledgeLabelId)
}

func (p *knowledgeLabelProxy) GetAllKnowledgebaseEntities(ctx context.Context, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.GetAllKnowledgebaseEntitiesAttr(ctx, p, published)
}

func (p *knowledgeLabelProxy) GetAllKnowledgeLabelEntities(ctx context.Context, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	return p.GetAllKnowledgeLabelEntitiesAttr(ctx, p, knowledgeBase)
}

func GetAllKnowledgebaseEntitiesFn(ctx context.Context, p *knowledgeLabelProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	var (
		after                 string
		err                   error
		knowledgeBaseEntities []platformclientv2.Knowledgebase
	)

	const pageSize = 100
	for {
		knowledgeBases, resp, getErr := p.KnowledgeApi.GetKnowledgeKnowledgebases("", after, "", fmt.Sprintf("%v", pageSize), "", "", published, "", "")
		if getErr != nil {
			return nil, resp, fmt.Errorf("failed to get page of knowledge bases error: %s", getErr)
		}

		if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
			break
		}

		knowledgeBaseEntities = append(knowledgeBaseEntities, *knowledgeBases.Entities...)

		if knowledgeBases.NextUri == nil || *knowledgeBases.NextUri == "" {
			break
		}

		after, err = util.GetQueryParamValueFromUri(*knowledgeBases.NextUri, "after")
		if err != nil {
			return nil, resp, fmt.Errorf("failed to parse after cursor from knowledge base nextUri: %s", err)
		}
		if after == "" {
			break
		}
	}

	return &knowledgeBaseEntities, nil, nil

}

func GetAllKnowledgeLabelEntitiesFn(ctx context.Context, p *knowledgeLabelProxy, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {

	var (
		after    string
		err      error
		entities []platformclientv2.Labelresponse
	)

	const pageSize = 100
	for {
		knowledgeLabels, resp, getErr := p.KnowledgeApi.GetKnowledgeKnowledgebaseLabels(*knowledgeBase.Id, "", after, fmt.Sprintf("%v", pageSize), "", false)
		if getErr != nil {
			return nil, resp, fmt.Errorf("failed to get page of knowledge bases error: %s", getErr)
		}

		if knowledgeLabels.Entities == nil || len(*knowledgeLabels.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeLabels.Entities...)

		if knowledgeLabels.NextUri == nil || *knowledgeLabels.NextUri == "" {
			break
		}

		after, err = util.GetQueryParamValueFromUri(*knowledgeLabels.NextUri, "after")
		if err != nil {
			return nil, resp, fmt.Errorf("failed to parse after cursor from knowledge base nextUri: %s", err)
		}
		if after == "" {
			break
		}
	}

	//Cache the knowledgeLabel resource into the p.authRoleCache for later use
	for _, knowledgeLabel := range entities {
		id := fmt.Sprintf("%s,%s", *knowledgeBase.Id, *knowledgeLabel.Id)
		rc.SetCache(p.knowledgeLabelCache, id, knowledgeLabel)
	}

	return &entities, nil, nil
}

func getKnowledgeLabelFn(ctx context.Context, p *knowledgeLabelProxy, knowledgeBaseId string, labelId string) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	id := fmt.Sprintf("%s,%s", knowledgeBaseId, labelId)
	if knowledgeLabel := rc.GetCacheItem(p.knowledgeLabelCache, id); knowledgeLabel != nil {
		return knowledgeLabel, nil, nil
	}
	return p.KnowledgeApi.GetKnowledgeKnowledgebaseLabel(knowledgeBaseId, labelId)
}

func createKnowledgeLabelFn(ctx context.Context, p *knowledgeLabelProxy, knowledgeBaseId string, body *platformclientv2.Labelcreaterequest) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.PostKnowledgeKnowledgebaseLabels(knowledgeBaseId, *body)
}

func deleteKnowledgeLabelFn(ctx context.Context, p *knowledgeLabelProxy, knowledgeBaseId string, knowledgeLabelId string) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	data, resp, err := p.KnowledgeApi.DeleteKnowledgeKnowledgebaseLabel(knowledgeBaseId, knowledgeLabelId)
	if err != nil {
		return nil, resp, err
	}
	id := fmt.Sprintf("%s,%s", knowledgeBaseId, knowledgeLabelId)
	rc.DeleteCacheItem(p.knowledgeLabelCache, id)
	return data, nil, nil
}

func updateKnowledgeLabelFn(ctx context.Context, p *knowledgeLabelProxy, knowledgeBaseId string, knowledgeLabelId string, body *platformclientv2.Labelupdaterequest) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.PatchKnowledgeKnowledgebaseLabel(knowledgeBaseId, knowledgeLabelId, *body)
}
