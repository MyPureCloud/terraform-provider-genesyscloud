package knowledge_category

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

var internalProxy *knowledgeCategoryProxy

type getAllKnowledgebaseEntitiesFunc func(ctx context.Context, p *knowledgeCategoryProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)
type getAllKnowledgeCategoryEntitiesFunc func(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBase *platformclientv2.Knowledgebase, categoryName string) (*[]platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)
type getKnowledgeKnowledgebaseCategoryFunc func(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)
type createKnowledgeCategoryFunc func(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, body platformclientv2.Categorycreaterequest) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)
type updateKnowledgeCategoryFunc func(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, categoryId string, body platformclientv2.Categoryupdaterequest) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)
type deleteKnowledgeCategoryFunc func(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)
type getKnowledgeCategoryByNameFunc func(ctx context.Context, p *knowledgeCategoryProxy, categoryName string, knowledgeBaseName string) (string, bool, *platformclientv2.APIResponse, error)
type knowledgeCategoryProxy struct {
	clientConfig                          *platformclientv2.Configuration
	KnowledgeApi                          *platformclientv2.KnowledgeApi
	getAllKnowledgebaseEntitiesAttr       getAllKnowledgebaseEntitiesFunc
	getAllKnowledgeCategoryEntitiesAttr   getAllKnowledgeCategoryEntitiesFunc
	getKnowledgeKnowledgebaseCategoryAttr getKnowledgeKnowledgebaseCategoryFunc
	createKnowledgeCategoryAttr           createKnowledgeCategoryFunc
	updateKnowledgeCategoryAttr           updateKnowledgeCategoryFunc
	deleteKnowledgeCategoryAttr           deleteKnowledgeCategoryFunc
	getKnowledgeCategoryByNameAttr        getKnowledgeCategoryByNameFunc
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
		getKnowledgeCategoryByNameAttr:        getKnowledgeCategoryByNameFn,
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

func (p *knowledgeCategoryProxy) getAllKnowledgeCategoryEntities(ctx context.Context, knowledgeBase *platformclientv2.Knowledgebase, categoryName string) (*[]platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.getAllKnowledgeCategoryEntitiesAttr(ctx, p, knowledgeBase, categoryName)
}

func (p *knowledgeCategoryProxy) getKnowledgeKnowledgebaseCategory(ctx context.Context, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeKnowledgebaseCategoryAttr(ctx, p, knowledgeBaseId, categoryId)
}

func (p *knowledgeCategoryProxy) getKnowledgeCategoryByName(ctx context.Context, categoryName string, knowledgeBaseName string) (string, bool, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeCategoryByNameAttr(ctx, p, categoryName, knowledgeBaseName)
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

func getAllKnowledgebaseEntitiesFn(ctx context.Context, p *knowledgeCategoryProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	var (
		after    string
		entities []platformclientv2.Knowledgebase
	)

	const pageSize = 100
	for {
		knowledgeBases, resp, getErr := p.KnowledgeApi.GetKnowledgeKnowledgebases("", after, "", fmt.Sprintf("%v", pageSize), "", "", published, "", "")
		if getErr != nil {
			return nil, resp, getErr
		}

		if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeBases.Entities...)

		if knowledgeBases.NextUri == nil || *knowledgeBases.NextUri == "" {
			break
		}

		after, err := util.GetQueryParamValueFromUri(*knowledgeBases.NextUri, "after")
		if err != nil {
			return nil, resp, err
		}
		if after == "" {
			break
		}
	}

	return &entities, nil, nil

}

func getAllKnowledgeCategoryEntitiesFn(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBase *platformclientv2.Knowledgebase, categoryName string) (*[]platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	var (
		after    string
		entities []platformclientv2.Categoryresponse
	)

	const pageSize = 100
	for i := 0; ; i++ {
		knowledgeCategories, resp, getErr := p.KnowledgeApi.GetKnowledgeKnowledgebaseCategories(*knowledgeBase.Id, "", after, fmt.Sprintf("%v", pageSize), "", false, categoryName, "", "", false)
		if getErr != nil {
			return nil, resp, getErr
		}

		if knowledgeCategories.Entities == nil || len(*knowledgeCategories.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeCategories.Entities...)

		if knowledgeCategories.NextUri == nil || *knowledgeCategories.NextUri == "" {
			break
		}

		after, err := util.GetQueryParamValueFromUri(*knowledgeCategories.NextUri, "after")
		if err != nil {
			return nil, resp, err
		}
		if after == "" {
			break
		}
	}

	for _, knowledgeCategory := range entities {
		if knowledgeCategory.Id == nil {
			continue
		}
		rc.SetCache(p.knowledgeCategoryCache, *knowledgeCategory.Id, knowledgeCategory)
	}

	return &entities, nil, nil
}

func getKnowledgeKnowledgebaseCategoryFn(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	if knowledgeCategory := rc.GetCacheItem(p.knowledgeCategoryCache, categoryId); knowledgeCategory != nil {
		return knowledgeCategory, nil, nil
	}
	return p.KnowledgeApi.GetKnowledgeKnowledgebaseCategory(knowledgeBaseId, categoryId)
}

func getKnowledgeCategoryByNameFn(ctx context.Context, p *knowledgeCategoryProxy, categoryName string, knowledgeBaseName string) (string, bool, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	publishedKnowledgeBases, publishedResp, getPublishedErr := p.KnowledgeApi.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), knowledgeBaseName, "", true, "", "")
	unpublishedKnowledgeBases, unpublishedResp, getUnpublishedErr := p.KnowledgeApi.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), knowledgeBaseName, "", false, "", "")

	if getPublishedErr != nil {
		return "", false, publishedResp, getPublishedErr
	}
	if getUnpublishedErr != nil {
		return "", false, unpublishedResp, getUnpublishedErr
	}

	noPublishedEntities := publishedKnowledgeBases.Entities == nil || len(*publishedKnowledgeBases.Entities) == 0
	noUnpublishedEntities := unpublishedKnowledgeBases.Entities == nil || len(*unpublishedKnowledgeBases.Entities) == 0
	if noPublishedEntities && noUnpublishedEntities {

		return "", true, publishedResp, nil
	}

	// prefer published knowledge base
	for _, knowledgeBase := range *publishedKnowledgeBases.Entities {
		if knowledgeBase.Name != nil && *knowledgeBase.Name == knowledgeBaseName {
			knowledgeCategories, resp, getErr := p.getAllKnowledgeCategoryEntities(ctx, &knowledgeBase, categoryName)

			if getErr != nil {
				return "", false, resp, getErr
			}

			for _, knowledgeCategory := range *knowledgeCategories {
				if *knowledgeCategory.Name == categoryName {
					id := fmt.Sprintf("%s,%s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id)
					return id, false, resp, nil
				}
			}
		}
	}

	for _, knowledgeBase := range *unpublishedKnowledgeBases.Entities {
		if knowledgeBase.Name != nil && *knowledgeBase.Name == knowledgeBaseName {
			knowledgeCategories, resp, getErr := p.getAllKnowledgeCategoryEntities(ctx, &knowledgeBase, categoryName)

			if getErr != nil {
				return "", false, resp, getErr
			}

			for _, knowledgeCategory := range *knowledgeCategories {
				if *knowledgeCategory.Name == categoryName {
					id := fmt.Sprintf("%s,%s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id)
					return id, false, resp, nil
				}
			}
		}
	}
	return "", true, publishedResp, nil
}
func createKnowledgeCategoryFn(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, body platformclientv2.Categorycreaterequest) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.PostKnowledgeKnowledgebaseCategories(knowledgeBaseId, body)
}

func updateKnowledgeCategoryFn(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, categoryId string, body platformclientv2.Categoryupdaterequest) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.PatchKnowledgeKnowledgebaseCategory(knowledgeBaseId, categoryId, body)
}

func deleteKnowledgeCategoryFn(ctx context.Context, p *knowledgeCategoryProxy, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.DeleteKnowledgeKnowledgebaseCategory(knowledgeBaseId, categoryId)
}
