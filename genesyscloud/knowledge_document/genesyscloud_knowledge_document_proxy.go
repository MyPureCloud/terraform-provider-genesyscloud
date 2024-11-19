package knowledge_document

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var internalProxy *knowledgeDocumentProxy

type getKnowledgeKnowledgebaseCategoryFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error)
type getKnowledgeKnowledgebaseCategoriesFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, categoryName string) (*platformclientv2.Categoryresponselisting, *platformclientv2.APIResponse, error)
type getKnowledgeKnowledgebaseLabelsFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, labelName string) (*platformclientv2.Labellisting, *platformclientv2.APIResponse, error)
type getKnowledgeKnowledgebaseLabelFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, labelId string) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error)
type getKnowledgeKnowledgebaseDocumentFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string, expand []string, state string) (*platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error)
type GetAllKnowledgebaseEntitiesFunc func(ctx context.Context, p *knowledgeDocumentProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)
type GetAllKnowledgeDocumentEntitiesFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error)
type createKnowledgeKnowledgebaseDocumentFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, body *platformclientv2.Knowledgedocumentcreaterequest) (*platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error)
type createKnowledgebaseDocumentVersionsFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string, body *platformclientv2.Knowledgedocumentversion) (*platformclientv2.Knowledgedocumentversion, *platformclientv2.APIResponse, error)
type deleteKnowledgeKnowledgebaseDocumentFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string) (*platformclientv2.APIResponse, error)
type updateKnowledgeKnowledgebaseDocumentFunc func(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string, body *platformclientv2.Knowledgedocumentreq) (*platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error)

type knowledgeDocumentProxy struct {
	clientConfig                             *platformclientv2.Configuration
	KnowledgeApi                             *platformclientv2.KnowledgeApi
	getKnowledgeKnowledgebaseCategoryAttr    getKnowledgeKnowledgebaseCategoryFunc
	getKnowledgeKnowledgebaseCategoriesAttr  getKnowledgeKnowledgebaseCategoriesFunc
	getKnowledgeKnowledgebaseLabelsAttr      getKnowledgeKnowledgebaseLabelsFunc
	getKnowledgeKnowledgebaseLabelAttr       getKnowledgeKnowledgebaseLabelFunc
	getKnowledgeKnowledgebaseDocumentAttr    getKnowledgeKnowledgebaseDocumentFunc
	GetAllKnowledgebaseEntitiesAttr          GetAllKnowledgebaseEntitiesFunc
	GetAllKnowledgeDocumentEntitiesAttr      GetAllKnowledgeDocumentEntitiesFunc
	createKnowledgeKnowledgebaseDocumentAttr createKnowledgeKnowledgebaseDocumentFunc
	createKnowledgebaseDocumentVersionsAttr  createKnowledgebaseDocumentVersionsFunc
	deleteKnowledgeKnowledgebaseDocumentAttr deleteKnowledgeKnowledgebaseDocumentFunc
	updateKnowledgeKnowledgebaseDocumentAttr updateKnowledgeKnowledgebaseDocumentFunc
	knowledgeDocumentCache                   rc.CacheInterface[platformclientv2.Knowledgedocumentresponse]
}

func newKnowledgeDocumentProxy(clientConfig *platformclientv2.Configuration) *knowledgeDocumentProxy {
	api := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)
	knowledgeDocumentCache := rc.NewResourceCache[platformclientv2.Knowledgedocumentresponse]()
	return &knowledgeDocumentProxy{
		clientConfig:                             clientConfig,
		KnowledgeApi:                             api,
		getKnowledgeKnowledgebaseCategoryAttr:    getKnowledgeKnowledgebaseCategoryFn,
		getKnowledgeKnowledgebaseCategoriesAttr:  getKnowledgeKnowledgebaseCategoriesFn,
		getKnowledgeKnowledgebaseLabelsAttr:      getKnowledgeKnowledgebaseLabelsFn,
		getKnowledgeKnowledgebaseLabelAttr:       getKnowledgeKnowledgebaseLabelFn,
		getKnowledgeKnowledgebaseDocumentAttr:    getKnowledgeKnowledgebaseDocumentFn,
		GetAllKnowledgebaseEntitiesAttr:          GetAllKnowledgebaseEntitiesFn,
		GetAllKnowledgeDocumentEntitiesAttr:      GetAllKnowledgeDocumentEntitiesFn,
		createKnowledgeKnowledgebaseDocumentAttr: createKnowledgeKnowledgebaseDocumentFn,
		createKnowledgebaseDocumentVersionsAttr:  createKnowledgebaseDocumentVersionsFn,
		deleteKnowledgeKnowledgebaseDocumentAttr: deleteKnowledgeKnowledgebaseDocumentFn,
		updateKnowledgeKnowledgebaseDocumentAttr: updateKnowledgeKnowledgebaseDocumentFn,
		knowledgeDocumentCache:                   knowledgeDocumentCache,
	}
}

func GetKnowledgeDocumentProxy(clientConfig *platformclientv2.Configuration) *knowledgeDocumentProxy {
	if internalProxy == nil {
		internalProxy = newKnowledgeDocumentProxy(clientConfig)
	}

	return internalProxy
}

func (p *knowledgeDocumentProxy) getKnowledgeKnowledgebaseCategory(ctx context.Context, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeKnowledgebaseCategoryAttr(ctx, p, knowledgeBaseId, categoryId)
}

func (p *knowledgeDocumentProxy) getKnowledgeKnowledgebaseCategories(ctx context.Context, knowledgeBaseId string, categoryName string) (*platformclientv2.Categoryresponselisting, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeKnowledgebaseCategoriesAttr(ctx, p, knowledgeBaseId, categoryName)
}

func (p *knowledgeDocumentProxy) getKnowledgeKnowledgebaseLabels(ctx context.Context, knowledgeBaseId string, labelName string) (*platformclientv2.Labellisting, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeKnowledgebaseLabelsAttr(ctx, p, knowledgeBaseId, labelName)
}

func (p *knowledgeDocumentProxy) getKnowledgeKnowledgebaseLabel(ctx context.Context, knowledgeBaseId string, labelId string) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeKnowledgebaseLabelAttr(ctx, p, knowledgeBaseId, labelId)
}

func (p *knowledgeDocumentProxy) getKnowledgeKnowledgebaseDocument(ctx context.Context, knowledgeBaseId string, documentId string, expand []string, state string) (*platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error) {
	return p.getKnowledgeKnowledgebaseDocumentAttr(ctx, p, knowledgeBaseId, documentId, expand, state)
}

func (p *knowledgeDocumentProxy) GetAllKnowledgebaseEntities(ctx context.Context, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.GetAllKnowledgebaseEntitiesAttr(ctx, p, published)
}

func (p *knowledgeDocumentProxy) GetAllKnowledgeDocumentEntities(ctx context.Context, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error) {
	return p.GetAllKnowledgeDocumentEntitiesAttr(ctx, p, knowledgeBase)
}

func (p *knowledgeDocumentProxy) createKnowledgeKnowledgebaseDocument(ctx context.Context, knowledgeBaseId string, body *platformclientv2.Knowledgedocumentcreaterequest) (*platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error) {
	return p.createKnowledgeKnowledgebaseDocumentAttr(ctx, p, knowledgeBaseId, body)
}

func (p *knowledgeDocumentProxy) createKnowledgebaseDocumentVersions(ctx context.Context, knowledgeBaseId string, documentId string, body *platformclientv2.Knowledgedocumentversion) (*platformclientv2.Knowledgedocumentversion, *platformclientv2.APIResponse, error) {
	return p.createKnowledgebaseDocumentVersionsAttr(ctx, p, knowledgeBaseId, documentId, body)
}

func (p *knowledgeDocumentProxy) deleteKnowledgeKnowledgebaseDocument(ctx context.Context, knowledgeBaseId string, documentId string) (*platformclientv2.APIResponse, error) {
	return p.deleteKnowledgeKnowledgebaseDocumentAttr(ctx, p, knowledgeBaseId, documentId)
}

func (p *knowledgeDocumentProxy) updateKnowledgeKnowledgebaseDocument(ctx context.Context, knowledgeBaseId string, documentId string, body *platformclientv2.Knowledgedocumentreq) (*platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error) {
	return p.updateKnowledgeKnowledgebaseDocumentAttr(ctx, p, knowledgeBaseId, documentId, body)
}

func getKnowledgeKnowledgebaseCategoryFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, categoryId string) (*platformclientv2.Categoryresponse, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.GetKnowledgeKnowledgebaseCategory(knowledgeBaseId, categoryId)
}

func getKnowledgeKnowledgebaseCategoriesFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, categoryName string) (*platformclientv2.Categoryresponselisting, *platformclientv2.APIResponse, error) {
	pageSize := 1
	return p.KnowledgeApi.GetKnowledgeKnowledgebaseCategories(knowledgeBaseId, "", "", fmt.Sprintf("%v", pageSize), "", false, categoryName, "", "", false)
}
func getKnowledgeKnowledgebaseLabelsFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, labelName string) (*platformclientv2.Labellisting, *platformclientv2.APIResponse, error) {
	pageSize := 1
	return p.KnowledgeApi.GetKnowledgeKnowledgebaseLabels(knowledgeBaseId, "", "", fmt.Sprintf("%v", pageSize), labelName, false)
}

func getKnowledgeKnowledgebaseLabelFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, labelId string) (*platformclientv2.Labelresponse, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.GetKnowledgeKnowledgebaseLabel(knowledgeBaseId, labelId)
}

func getKnowledgeKnowledgebaseDocumentFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string, expand []string, state string) (*platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.GetKnowledgeKnowledgebaseDocument(knowledgeBaseId, documentId, expand, state)
}

func GetAllKnowledgebaseEntitiesFn(ctx context.Context, p *knowledgeDocumentProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	var (
		after    string
		entities []platformclientv2.Knowledgebase
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

		entities = append(entities, *knowledgeBases.Entities...)

		if knowledgeBases.NextUri == nil || *knowledgeBases.NextUri == "" {
			break
		}

		after, err := util.GetQueryParamValueFromUri(*knowledgeBases.NextUri, "after")
		if err != nil {
			return nil, resp, fmt.Errorf("failed to parse after cursor from knowledge base nextUri: %s", err)
		}
		if after == "" {
			break
		}
	}

	return &entities, nil, nil

}

func GetAllKnowledgeDocumentEntitiesFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error) {

	var (
		after    string
		entities []platformclientv2.Knowledgedocumentresponse
	)

	resources := make(resourceExporter.ResourceIDMetaMap)

	const pageSize = 100
	// prepare base url
	resourcePath := fmt.Sprintf("/api/v2/knowledge/knowledgebases/%s/documents", url.PathEscape(*knowledgeBase.Id))
	listDocumentsBaseUrl := fmt.Sprintf("%s%s", p.KnowledgeApi.Configuration.BasePath, resourcePath)

	for {
		// prepare query params
		queryParams := make(map[string]string, 0)
		queryParams["after"] = after
		queryParams["pageSize"] = fmt.Sprintf("%v", pageSize)
		queryParams["includeDrafts"] = "true"

		// prepare headers
		headers := make(map[string]string)
		headers["Authorization"] = fmt.Sprintf("Bearer %s", p.clientConfig.AccessToken)
		headers["Content-Type"] = "application/json"
		headers["Accept"] = "application/json"

		// execute request
		response, err := p.clientConfig.APIClient.CallAPI(listDocumentsBaseUrl, "GET", nil, headers, queryParams, nil, "", nil, "")
		if err != nil {
			return nil, response, fmt.Errorf("failed to read knowledge document list response error: %s", err)
		}

		// process response
		var knowledgeDocuments platformclientv2.Knowledgedocumentresponselisting
		unmarshalErr := json.Unmarshal(response.RawBody, &knowledgeDocuments)
		if unmarshalErr != nil {
			return nil, response, fmt.Errorf("failed to unmarshal knowledge document list response: %s", unmarshalErr)
		}

		/**
		 * Todo: restore direct SDK invocation and remove workaround once the SDK supports optional boolean args.
		 */
		// knowledgeDocuments, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseDocuments(*knowledgeBase.Id, "", after, fmt.Sprintf("%v", pageSize), "", nil, nil, true, true, nil, nil)
		// if getErr != nil {
		// 	return nil, diag.Errorf("Failed to get page of knowledge documents: %v", getErr)
		// }

		if knowledgeDocuments.Entities == nil || len(*knowledgeDocuments.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeDocuments.Entities...)

		if knowledgeDocuments.NextUri == nil || *knowledgeDocuments.NextUri == "" {
			break
		}

		after, err = util.GetQueryParamValueFromUri(*knowledgeDocuments.NextUri, "after")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse after cursor from knowledge document nextUri: %s", err)
		}
		if after == "" {
			break
		}
		for _, knowledgeDocument := range *knowledgeDocuments.Entities {
			id := fmt.Sprintf("%s,%s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id)
			resources[id] = &resourceExporter.ResourceMeta{Name: *knowledgeDocument.Title}
		}
	}

	return &entities, nil, nil
}
func createKnowledgeKnowledgebaseDocumentFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, body *platformclientv2.Knowledgedocumentcreaterequest) (*platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.PostKnowledgeKnowledgebaseDocuments(knowledgeBaseId, *body)
}

func createKnowledgebaseDocumentVersionsFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string, body *platformclientv2.Knowledgedocumentversion) (*platformclientv2.Knowledgedocumentversion, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.PostKnowledgeKnowledgebaseDocumentVersions(knowledgeBaseId, documentId, *body)
}

func deleteKnowledgeKnowledgebaseDocumentFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string) (*platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.DeleteKnowledgeKnowledgebaseDocument(knowledgeBaseId, documentId)
}

func updateKnowledgeKnowledgebaseDocumentFn(ctx context.Context, p *knowledgeDocumentProxy, knowledgeBaseId string, documentId string, body *platformclientv2.Knowledgedocumentreq) (*platformclientv2.Knowledgedocumentresponse, *platformclientv2.APIResponse, error) {
	return p.KnowledgeApi.PatchKnowledgeKnowledgebaseDocument(knowledgeBaseId, documentId, *body)
}
