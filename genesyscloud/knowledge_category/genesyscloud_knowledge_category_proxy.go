package knowledge_category

import (
	"context"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var internalProxy *knowledgeCategoryProxy

type getAllKnowledgebaseEntitiesFunc func(ctx context.Context, p *knowledgeCategoryProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)
type getAllKnowledgeCategoryEntitiesFunc func(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)
type getKnowledgeKnowledgebaseCategoryFunc func(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)
type createKnowledgeCategoryFunc func(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, body platformclientv2.Categorycreaterequest) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)
type updateKnowledgeCategoryFunc func(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, categoryId string, body platformclientv2.Categoryupdaterequest) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)
type deleteKnowledgeCategoryFunc func(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)

type knowledgeCategoryProxy struct {
	clientConfig                          *platformclientv2.Configuration
	KnowledgeApi                          *platformclientv2.KnowledgeApi
	getAllKnowledgebaseEntitiesAttr       getAllKnowledgebaseEntitiesFunc
	getAllKnowledgeCategoryEntitiesAttr   getAllKnowledgeCategoryEntitiesFunc
	getKnowledgeKnowledgebaseCategoryAttr getKnowledgeKnowledgebaseCategoryFunc
	createKnowledgeCategoryAttr           createKnowledgeCategoryFunc
	updateKnowledgeCategoryAttr           updateKnowledgeCategoryFunc
	deleteKnowledgeCategoryAttr           deleteKnowledgeCategoryFunc
	knowledgeCategoryCache                rc.CacheInterface[platformclientv2.Categoryresponse]
}

func newKnowledgeCategoryProxy(clientConfig *platformclientv2.Configuration) *knowledgeCategoryProxy {
	api := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)
	knowledgeCategoryCache := rc.NewResourceCache[platformclientv2.Categoryresponse]()
	return &knowledgeCategoryProxy{
		clientConfig:                          clientConfig,
		KnowledgeApi:                          api,
		getAllKnowledgebaseEntitiesAttr:       getAllKnowledgebaseEntitiesFn,
		getAllKnowledgeCategoryEntitiesAttr:   getAllKnowledgeCategoryEntitiesFn,
		getKnowledgeKnowledgebaseCategoryAttr: getKnowledgeKnowledgebaseCategoryFn,
		createKnowledgeCategoryAttr:           createKnowledgeCategoryFn,
		updateKnowledgeCategoryAttr:           updateKnowledgeCategoryFn,
		deleteKnowledgeCategoryAttr:           deleteKnowledgeCategoryFn,
		knowledgeCategoryCache:                knowledgeCategoryCache,
	}
}

func GetKnowledgeCategoryProxy(clientConfig *platformclientv2.Configuration) *knowledgeCategoryProxy {
	if internalProxy == nil {
		internalProxy = newKnowledgeCategoryProxy(clientConfig)
	}
	return internalProxy
}

func (p *knowledgeCategoryProxy) getAllKnowledgebaseEntities(ctx context.Context, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.getAllKnowledgebaseEntitiesAttr(ctx, p, published)
}

func (p *knowledgeCategoryProxy) getAllKnowledgeCategoryEntities(ctx context.Context, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.getAllKnowledgeCategoryEntitiesAttr(ctx, p, knowledgeBase)
}

func (p *knowledgeCategoryProxy) getKnowledgeKnowledgebaseCategory(ctx context.Context, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeKnowledgebaseCategoryAttr(ctx, p, knowledgeBaseId, categoryId)
}

func (p *knowledgeCategoryProxy) createKnowledgeCategory(ctx context.Context, knowledgeBaseId string, body platformclientv2.Categorycreaterequest) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.createKnowledgeCategoryAttr(ctx, p, knowledgeBaseId, body)
}

func (p *knowledgeCategoryProxy) updateKnowledgeCategory(ctx context.Context, knowledgeBaseId string, categoryId string, body platformclientv2.Categoryupdaterequest) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.updateKnowledgeCategoryAttr(ctx, p, knowledgeBaseId, categoryId, body)
}

func (p *knowledgeCategoryProxy) deleteKnowledgeCategory(ctx context.Context, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.deleteKnowledgeCategoryAttr(ctx, p, knowledgeBaseId, categoryId)
}
