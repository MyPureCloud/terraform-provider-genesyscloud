package knowledge_document

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *knowledgeDocumentProxy

type getKnowledgeKnowledgebaseCategoriesFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, pageSize int, categoryName string) (*[]platformclientv2.Knowledgecategory, *platformclientv2.APIResponse, error)
type getKnowledgeKnowledgebaseLabelsFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, pageSize int, labelName string) (*[]platformclientv2.Knowledgelabel, *platformclientv2.APIResponse, error)
type getKnowledgeKnowledgebaseDocumentFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string) (*platformclientv2.KnowledgeDocument, *platformclientv2.APIResponse, error)
type getAllKnowledgebaseEntitiesFunc func(ctx context.Context, p *knowledgeDocumentProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)
type createKnowledgeKnowledgebaseDocumentFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, body *platformclientv2.Knowledgecategory) (*platformclientv2.Knowledgecategory, *platformclientv2.APIResponse, error)
type deleteKnowledgeKnowledgebaseDocumentFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string) (*platformclientv2.APIResponse, error)
type updateKnowledgeKnowledgebaseDocumentFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string, body *platformclientv2.KnowledgeDocument) (*platformclientv2.KnowledgeDocument, *platformclientv2.APIResponse, error)

type knowledgeDocumentProxy struct {
	clientConfig                             *platformclientv2.Configuration
	knowledgeApi                             *platformclientv2.KnowledgeApi
	getKnowledgeKnowledgebaseCategoriesAttr  getKnowledgeKnowledgebaseCategoriesFunc
	getKnowledgeKnowledgebaseLabelsAttr      getKnowledgeKnowledgebaseLabelsFunc
	getKnowledgeKnowledgebaseDocumentAttr    getKnowledgeKnowledgebaseDocumentFunc
	getAllKnowledgebaseEntitiesAttr          getAllKnowledgebaseEntitiesFunc
	createKnowledgeKnowledgebaseDocumentAttr createKnowledgeKnowledgebaseDocumentFunc
	deleteKnowledgeKnowledgebaseDocumentAttr deleteKnowledgeKnowledgebaseDocumentFunc
	updateKnowledgeKnowledgebaseDocumentAttr updateKnowledgeKnowledgebaseDocumentFunc
	knowledgeDocumentCache                   rc.CacheInterface[platformclientv2.KnowledgeDocument]
}

func newKnowledgeDocumentProxy(clientConfig *platformclientv2.Configuration) *knowledgeDocumentProxy {
	api := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)
	knowledgeDocumentCache := rc.NewResourceCache[platformclientv2.KnowledgeDocument]()
	return &knowledgeDocumentProxy{
		clientConfig:                             clientConfig,
		knowledgeApi:                             api,
		getKnowledgeKnowledgebaseCategoriesAttr:  getKnowledgeKnowledgebaseCategoriesFn,
		getKnowledgeKnowledgebaseLabelsAttr:      getKnowledgeKnowledgebaseLabelsFn,
		getKnowledgeKnowledgebaseDocumentAttr:    getKnowledgeKnowledgebaseDocumentFn,
		getAllKnowledgebaseEntitiesAttr:          getAllKnowledgebaseEntitiesFn,
		createKnowledgeKnowledgebaseDocumentAttr: createKnowledgeKnowledgebaseDocumentFn,
		deleteKnowledgeKnowledgebaseDocumentAttr: deleteKnowledgeKnowledgebaseDocumentFn,
		updateKnowledgeKnowledgebaseDocumentAttr: updateKnowledgeKnowledgebaseDocumentFn,
		knowledgeDocumentCache:                   knowledgeDocumentCache,
	}
}

func getKnowledgeDocumentProxy(clientConfig *platformclientv2.Configuration) *knowledgeDocumentProxy {
	if internalProxy == nil {
		internalProxy = newKnowledgeDocumentProxy(clientConfig)
	}

	return internalProxy
}

func (p *knowledgeDocumentProxy) getKnowledgeKnowledgebaseCategories(ctx context.Context, knowledgeBaseId string, pageSize int, categoryName string) (*[]platformclientv2.Knowledgecategory, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeKnowledgebaseCategoriesAttr(ctx, p, knowledgeBaseId, pageSize, categoryName)
}

func (p *knowledgeDocumentProxy) getKnowledgeKnowledgebaseLabels(ctx context.Context, knowledgeBaseId string, pageSize int, labelName string) (*[]platformclientv2.Knowledgelabel, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeKnowledgebaseLabelsAttr(ctx, p, knowledgeBaseId, pageSize, labelName)
}

func (p *knowledgeDocumentProxy) getKnowledgeKnowledgebaseDocument(ctx context.Context, knowledgeBaseId string, documentId string) (*platformclientv2.KnowledgeDocument, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeKnowledgebaseDocumentAttr(ctx, p, knowledgeBaseId, documentId)
}

func (p *knowledgeDocumentProxy) getAllKnowledgebaseEntities(ctx context.Context, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.getAllKnowledgebaseEntitiesAttr(ctx, p, published)
}

func (p *knowledgeDocumentProxy) createKnowledgeKnowledgebaseDocument(ctx context.Context, knowledgeBaseId string, body *platformclientv2.Knowledgecategory) (*platformclientv2.Knowledgecategory, *platformclientv2.APIResponse, error) {
	return p.createKnowledgeKnowledgebaseDocumentAttr(ctx, p, knowledgeBaseId, body)
}

func (p *knowledgeDocumentProxy) deleteKnowledgeKnowledgebaseDocument(ctx context.Context, knowledgeBaseId string, documentId string) (*platformclientv2.APIResponse, error) {
	return p.deleteKnowledgeKnowledgebaseDocumentAttr(ctx, p, knowledgeBaseId, documentId)
}

func (p *knowledgeDocumentProxy) updateKnowledgeKnowledgebaseDocument(ctx context.Context, knowledgeBaseId string, documentId string, body *platformclientv2.KnowledgeDocument) (*platformclientv2.KnowledgeDocument, *platformclientv2.APIResponse, error) {
	return p.updateKnowledgeKnowledgebaseDocumentAttr(ctx, p, knowledgeBaseId, documentId, body)
}

func getKnowledgeKnowledgebaseCategoriesFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, pageSize int, categoryName string) (*[]platformclientv2.Knowledgecategory, *platformclientv2.APIResponse, error) {

}
