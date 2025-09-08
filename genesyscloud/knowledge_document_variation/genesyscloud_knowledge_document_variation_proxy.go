package knowledge_document_variation

import (
	"context"
	"fmt"
	"log"
	"net/http"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"time"
)

var internalProxy *variationRequestProxy

type createVariationFunc func(ctx context.Context, p *variationRequestProxy, documentVariationRequest *platformclientv2.Documentvariationrequest, knowledgeDocumentId, knowledgeBaseId string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error)
type getAllVariationsFunc func(ctx context.Context, p *variationRequestProxy, knowledgeBaseId, documentId, documentState string, expand []string) (*[]platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error)
type getVariationRequestByIdFunc func(ctx context.Context, p *variationRequestProxy, documentVariationId string, documentId string, knowledgeBaseId string, documentState string, expand []string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error)
type updateVariationRequestFunc func(ctx context.Context, p *variationRequestProxy, documentVariationId string, documentId string, knowledgeBaseId string, body platformclientv2.Documentvariationrequest) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error)
type deleteVariationRequestFunc func(ctx context.Context, p *variationRequestProxy, variationId, documentId, baseId string) (*platformclientv2.APIResponse, error)
type getVariationRequestIdByNameFunc func(ctx context.Context, p *variationRequestProxy, name, knowledgeBaseID, knowledgeDocumentID string) (string, *platformclientv2.APIResponse, bool, error)
type createKnowledgeKnowledgebaseDocumentVersionsFunc func(ctx context.Context, p *variationRequestProxy, knowledgeDocumentId, knowledgeBaseId string, version *platformclientv2.Knowledgedocumentversion) (*platformclientv2.Knowledgedocumentversion, *platformclientv2.APIResponse, error)
type GetAllKnowledgebaseEntitiesFunc func(ctx context.Context, p *variationRequestProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error)
type getLatestPublishedOrDraftVariationFunc func(ctx context.Context, p *variationRequestProxy, ids *resourceIDs) (variation *platformclientv2.Documentvariationresponse, response *platformclientv2.APIResponse, err error)

// variationRequestProxy contains all of the methods that call genesys cloud APIs.
type variationRequestProxy struct {
	clientConfig                                     *platformclientv2.Configuration
	knowledgeApi                                     *platformclientv2.KnowledgeApi
	createVariationAttr                              createVariationFunc
	getAllVariationsAttr                             getAllVariationsFunc
	getVariationRequestByIdAttr                      getVariationRequestByIdFunc
	getVariationRequestIdByNameAttr                  getVariationRequestIdByNameFunc
	updateVariationRequestAttr                       updateVariationRequestFunc
	deleteVariationRequestAttr                       deleteVariationRequestFunc
	createKnowledgeKnowledgebaseDocumentVersionsAttr createKnowledgeKnowledgebaseDocumentVersionsFunc
	GetAllKnowledgebaseEntitiesAttr                  GetAllKnowledgebaseEntitiesFunc
	getLatestPublishedOrDraftVariationAttr           getLatestPublishedOrDraftVariationFunc
	variationCache                                   rc.CacheInterface[platformclientv2.Documentvariationresponse]
}

// newVariationRequestProxy initializes the variation request proxy with all of the data needed to communicate with Genesys Cloud
func newVariationRequestProxy(clientConfig *platformclientv2.Configuration) *variationRequestProxy {
	api := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)
	variationCache := rc.NewResourceCache[platformclientv2.Documentvariationresponse]()
	return &variationRequestProxy{
		clientConfig:                                     clientConfig,
		knowledgeApi:                                     api,
		createVariationAttr:                              createVariationFn,
		getAllVariationsAttr:                             getAllVariationsFn,
		getVariationRequestByIdAttr:                      getVariationRequestByIdFn,
		updateVariationRequestAttr:                       updateVariationRequestFn,
		deleteVariationRequestAttr:                       deleteVariationRequestFn,
		getVariationRequestIdByNameAttr:                  getVariationRequestIdByNameFn,
		createKnowledgeKnowledgebaseDocumentVersionsAttr: createKnowledgeKnowledgebaseDocumentVersionsFn,
		GetAllKnowledgebaseEntitiesAttr:                  getAllKnowledgebaseEntitiesFn,
		getLatestPublishedOrDraftVariationAttr:           getLatestPublishedOrDraftVariationFn,
		variationCache:                                   variationCache,
	}
}

func getVariationRequestProxy(clientConfig *platformclientv2.Configuration) *variationRequestProxy {
	if internalProxy == nil {
		internalProxy = newVariationRequestProxy(clientConfig)
	}
	return internalProxy
}

// CreateVariation creates a Genesys Cloud variation request
func (p *variationRequestProxy) CreateVariation(ctx context.Context, variationRequest *platformclientv2.Documentvariationrequest, knowledgeDocumentId, knowledgeBaseId string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	return p.createVariationAttr(ctx, p, variationRequest, knowledgeDocumentId, knowledgeBaseId)
}

// getVariationRequest retrieves all Genesys Cloud variation request
func (p *variationRequestProxy) getAllVariations(ctx context.Context, knowledgeBaseId, documentId, documentState string, expand []string) (*[]platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	return p.getAllVariationsAttr(ctx, p, knowledgeBaseId, documentId, documentState, expand)
}

// getVariationRequestById returns a single Genesys Cloud variation request by Id
func (p *variationRequestProxy) getVariationRequestById(ctx context.Context, documentVariationId string, documentId string, knowledgeBaseId string, documentState string, expand []string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	return p.getVariationRequestByIdAttr(ctx, p, documentVariationId, documentId, knowledgeBaseId, documentState, expand)
}

// getVariationRequestIdByName returns a single Genesys Cloud variation request by a name
func (p *variationRequestProxy) getVariationRequestIdByName(ctx context.Context, name, knowledgeBaseID, knowledgeDocumentID string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getVariationRequestIdByNameAttr(ctx, p, name, knowledgeBaseID, knowledgeDocumentID)
}

// updateVariationRequest updates a Genesys Cloud variation request
func (p *variationRequestProxy) updateVariationRequest(ctx context.Context, documentVariationId string, documentId string, knowledgeBaseId string, body platformclientv2.Documentvariationrequest) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	return p.updateVariationRequestAttr(ctx, p, documentVariationId, documentId, knowledgeBaseId, body)
}

// deleteVariationRequest deletes a Genesys Cloud variation request by Id
func (p *variationRequestProxy) deleteVariationRequest(ctx context.Context, variationId, documentId, baseId string) (*platformclientv2.APIResponse, error) {
	return p.deleteVariationRequestAttr(ctx, p, variationId, documentId, baseId)
}

func (p *variationRequestProxy) createKnowledgeKnowledgebaseDocumentVersions(ctx context.Context, knowledgeDocumentId, knowledgeBaseId string, version *platformclientv2.Knowledgedocumentversion) (*platformclientv2.Knowledgedocumentversion, *platformclientv2.APIResponse, error) {
	return p.createKnowledgeKnowledgebaseDocumentVersionsAttr(ctx, p, knowledgeDocumentId, knowledgeBaseId, version)
}

func (p *variationRequestProxy) GetAllKnowledgebaseEntities(ctx context.Context, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	return p.GetAllKnowledgebaseEntitiesAttr(ctx, p, published)
}

// getVariationRequestByIdAndState reads the variation by ID and state
// If no state is specified, get both published and draft variation and choose the most recent
func (p *variationRequestProxy) getVariationRequestByIdAndState(ctx context.Context, ids *resourceIDs, state string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	if state == "" {
		return p.getLatestPublishedOrDraftVariation(ctx, ids)
	}
	return p.getVariationRequestById(ctx, ids.knowledgeDocumentVariationID, ids.knowledgeDocumentID, ids.knowledgeBaseID, state, nil)
}

func (p *variationRequestProxy) getLatestPublishedOrDraftVariation(ctx context.Context, ids *resourceIDs) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	return p.getLatestPublishedOrDraftVariationAttr(ctx, p, ids)
}

func getLatestPublishedOrDraftVariationFn(ctx context.Context, p *variationRequestProxy, ids *resourceIDs) (_ *platformclientv2.Documentvariationresponse, resp *platformclientv2.APIResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("getLatestPublishedOrDraftVariationFn: %w", err)
		}
	}()

	var publishedVariation *platformclientv2.Documentvariationresponse

	const maxRetries = 3
	for i := 1; i <= maxRetries; i++ {
		publishedVariation, resp, err = p.getVariationRequestByIdAndState(ctx, ids, "Published")
		if err == nil {
			break
		}
		if util.IsStatus404(resp) {
			publishedVariation = nil
			time.Sleep(1 * time.Second)
			continue
		}
		return nil, resp, err
	}

	draftVariation, resp, err := p.getVariationRequestByIdAndState(ctx, ids, "Draft")
	if err != nil {
		return nil, resp, err
	}

	if publishedVariation == nil {
		return draftVariation, resp, nil
	}

	if publishedVariation.DateModified == nil || draftVariation.DateModified == nil {
		log.Println("getLatestPublishedOrDraftVariation: cannot determine which variation was modified more recently. Returning published.")
		return publishedVariation, resp, nil
	}
	if publishedVariation.DateModified.After(*draftVariation.DateModified) {
		return publishedVariation, resp, nil
	}
	return draftVariation, resp, nil
}

// createVariationFn is an implementation function for creating a Genesys Cloud variation request
func createVariationFn(ctx context.Context, p *variationRequestProxy, variationRequest *platformclientv2.Documentvariationrequest, knowledgeDocumentId, knowledgeBaseId string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	return p.knowledgeApi.PostKnowledgeKnowledgebaseDocumentVariations(knowledgeBaseId, knowledgeDocumentId, *variationRequest)
}

// getAllVariationsFn is the implementation for retrieving all variation request in Genesys Cloud
func getAllVariationsFn(_ context.Context, p *variationRequestProxy, knowledgeBaseId, documentId, documentState string, expand []string) (*[]platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	var (
		allVariations []platformclientv2.Documentvariationresponse
		after         string
		resp          *platformclientv2.APIResponse
	)

	const pageSize = "100"

	for {
		variations, resp, err := p.knowledgeApi.GetKnowledgeKnowledgebaseDocumentVariations(knowledgeBaseId, documentId, "", after, pageSize, documentState, expand)
		if err != nil {
			return nil, resp, err
		}

		if variations.Entities == nil || len(*variations.Entities) == 0 {
			break
		}

		allVariations = append(allVariations, *variations.Entities...)

		if variations.NextUri == nil || *variations.NextUri == "" {
			break
		}

		after, err := util.GetQueryParamValueFromUri(*variations.NextUri, "after")
		if err != nil {
			return nil, resp, err
		}
		if after == "" {
			break
		}
	}

	for _, variation := range allVariations {
		rc.SetCache(p.variationCache, *variation.Id, variation)
	}

	return &allVariations, resp, nil
}

// getVariationRequestByIdFn is an implementation of the function to get a Genesys Cloud variation request by Id
func getVariationRequestByIdFn(_ context.Context, p *variationRequestProxy, documentVariationId string, documentId string, knowledgeBaseId string, documentState string, expand []string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	variation := rc.GetCacheItem(p.variationCache, documentVariationId)
	if variation != nil {
		return variation, nil, nil
	}
	return p.knowledgeApi.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, documentId, knowledgeBaseId, documentState, expand)
}

// getVariationRequestIdByNameFn is an implementation of the function to get a Genesys Cloud variation request by name
func getVariationRequestIdByNameFn(ctx context.Context, p *variationRequestProxy, name, knowledgeBaseID, knowledgeDocumentID string) (string, *platformclientv2.APIResponse, bool, error) {
	var allVariations []platformclientv2.Documentvariationresponse

	// API throws a 404 if no variations of particular documentState are found
	// Check for published state, ignore 404 and check for draft state
	// Append the two lists together and proceed as normal provided there is at least 1 entity returned
	allPublishedVariations, resp, err := getAllVariationsFn(ctx, p, knowledgeBaseID, knowledgeDocumentID, "Published", []string{})
	if err != nil && resp.StatusCode != http.StatusNotFound {
		return "", resp, false, err
	}
	allDraftVariations, resp, err := getAllVariationsFn(ctx, p, knowledgeBaseID, knowledgeDocumentID, "Draft", []string{})
	if err != nil {
		return "", resp, false, err
	}

	if allPublishedVariations != nil {
		allVariations = append(allVariations, *allPublishedVariations...)
	}
	if allDraftVariations != nil {
		allVariations = append(allVariations, *allDraftVariations...)
	}

	if len(allVariations) == 0 {
		return "", resp, true, err
	}

	for _, variation := range allVariations {
		if *variation.Name == name {
			log.Printf("Retrieved the variation request id %s by name %s", *variation.Id, name)
			return *variation.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("unable to find variation request with name %s", name)
}

// updateVariationRequestFn is an implementation of the function to update a Genesys Cloud variation request
func updateVariationRequestFn(_ context.Context, p *variationRequestProxy, documentVariationId string, documentId string, knowledgeBaseId string, body platformclientv2.Documentvariationrequest) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	return p.knowledgeApi.PatchKnowledgeKnowledgebaseDocumentVariation(documentVariationId, documentId, knowledgeBaseId, body)
}

// deleteVariationRequestFn is an implementation function for deleting a Genesys Cloud variation request
func deleteVariationRequestFn(_ context.Context, p *variationRequestProxy, variationId, documentId, baseId string) (*platformclientv2.APIResponse, error) {
	resp, err := p.knowledgeApi.DeleteKnowledgeKnowledgebaseDocumentVariation(variationId, documentId, baseId)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.variationCache, variationId)
	return nil, nil
}

func createKnowledgeKnowledgebaseDocumentVersionsFn(ctx context.Context, p *variationRequestProxy, knowledgeDocumentId, knowledgeBaseId string, version *platformclientv2.Knowledgedocumentversion) (*platformclientv2.Knowledgedocumentversion, *platformclientv2.APIResponse, error) {
	return p.knowledgeApi.PostKnowledgeKnowledgebaseDocumentVersions(knowledgeBaseId, knowledgeDocumentId, *version)
}

func getAllKnowledgebaseEntitiesFn(ctx context.Context, p *variationRequestProxy, published bool) (*[]platformclientv2.Knowledgebase, *platformclientv2.APIResponse, error) {
	var (
		after                 string
		err                   error
		knowledgeBaseEntities []platformclientv2.Knowledgebase
	)

	const pageSize = 100
	for {
		knowledgeBases, resp, getErr := p.knowledgeApi.GetKnowledgeKnowledgebases("", after, "", fmt.Sprintf("%v", pageSize), "", "", published, "", "")
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
