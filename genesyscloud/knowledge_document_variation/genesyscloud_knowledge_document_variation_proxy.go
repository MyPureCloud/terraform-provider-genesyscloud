package knowledgedocumentvariation

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

var internalProxy *variationRequestProxy

type createVariationFunc func(ctx context.Context, p *variationRequestProxy, documentVariationRequest *platformclientv2.Documentvariationrequest, knowledgeDocumentId, knowledgeBaseId string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error)
type getAllVariationsFunc func(ctx context.Context, p *variationRequestProxy) (*[]platformclientv2.Documentvariationrequest, *platformclientv2.APIResponse, error)
type getVariationRequestIdByNameFunc func(ctx context.Context, p *variationRequestProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getVariationRequestByIdFunc func(ctx context.Context, p *variationRequestProxy, documentVariationId string, documentId string, knowledgeBaseId string, documentState string, expand []string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error)
type updateVariationRequestFunc func(ctx context.Context, p *variationRequestProxy, id string, documentVariationRequest *platformclientv2.Documentvariationrequest) (*platformclientv2.Documentvariationrequest, *platformclientv2.APIResponse, error)
type deleteVariationRequestFunc func(ctx context.Context, p *variationRequestProxy, variationId, documentId, baseId string) (*platformclientv2.APIResponse, error)

type createKnowledgeKnowledgebaseDocumentVersionsFunc func(ctx context.Context, p *variationRequestProxy, knowledgeDocumentId, knowledgeBaseId string, version *platformclientv2.Knowledgedocumentversion) (*platformclientv2.Knowledgedocumentversion, *platformclientv2.APIResponse, error)

// variationRequestProxy contains all of the methods that call genesys cloud APIs.
type variationRequestProxy struct {
	clientConfig                    *platformclientv2.Configuration
	knowledgeApi                    *platformclientv2.KnowledgeApi
	createVariationAttr             createVariationFunc
	getAllVariationsAttr            getAllVariationsFunc
	getVariationRequestIdByNameAttr getVariationRequestIdByNameFunc
	getVariationRequestByIdAttr     getVariationRequestByIdFunc
	updateVariationRequestAttr      updateVariationRequestFunc
	deleteVariationRequestAttr      deleteVariationRequestFunc

	createKnowledgeKnowledgebaseDocumentVersionsAttr createKnowledgeKnowledgebaseDocumentVersionsFunc
}

// newVariationRequestProxy initializes the variation request proxy with all of the data needed to communicate with Genesys Cloud
func newVariationRequestProxy(clientConfig *platformclientv2.Configuration) *variationRequestProxy {
	api := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)
	return &variationRequestProxy{
		clientConfig:                    clientConfig,
		knowledgeApi:                    api,
		createVariationAttr:             createVariationFn,
		getAllVariationsAttr:            getAllVariationsFn,
		getVariationRequestIdByNameAttr: getVariationRequestIdByNameFn,
		getVariationRequestByIdAttr:     getVariationRequestByIdFn,
		updateVariationRequestAttr:      updateVariationRequestFn,
		deleteVariationRequestAttr:      deleteVariationRequestFn,

		createKnowledgeKnowledgebaseDocumentVersionsAttr: createKnowledgeKnowledgebaseDocumentVersionsFn,
	}
}

// getVariationRequestProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
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
func (p *variationRequestProxy) getAllVariations(ctx context.Context) (*[]platformclientv2.Documentvariationrequest, *platformclientv2.APIResponse, error) {
	return p.getAllVariationsAttr(ctx, p)
}

// getVariationRequestIdByName returns a single Genesys Cloud variation request by a name
func (p *variationRequestProxy) getVariationRequestIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getVariationRequestIdByNameAttr(ctx, p, name)
}

// getVariationRequestById returns a single Genesys Cloud variation request by Id
func (p *variationRequestProxy) getVariationRequestById(ctx context.Context, documentVariationId string, documentId string, knowledgeBaseId string, documentState string, expand []string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	return p.getVariationRequestByIdAttr(ctx, p, documentVariationId, documentId, knowledgeBaseId, documentState, expand)
}

// updateVariationRequest updates a Genesys Cloud variation request
func (p *variationRequestProxy) updateVariationRequest(ctx context.Context, id string, variationRequest *platformclientv2.Documentvariationrequest) (*platformclientv2.Documentvariationrequest, *platformclientv2.APIResponse, error) {
	return p.updateVariationRequestAttr(ctx, p, id, variationRequest)
}

// deleteVariationRequest deletes a Genesys Cloud variation request by Id
func (p *variationRequestProxy) deleteVariationRequest(ctx context.Context, variationId, documentId, baseId string) (*platformclientv2.APIResponse, error) {
	return p.deleteVariationRequestAttr(ctx, p, variationId, documentId, baseId)
}

func (p *variationRequestProxy) createKnowledgeKnowledgebaseDocumentVersions(ctx context.Context, knowledgeDocumentId, knowledgeBaseId string, version *platformclientv2.Knowledgedocumentversion) (*platformclientv2.Knowledgedocumentversion, *platformclientv2.APIResponse, error) {
	return p.createKnowledgeKnowledgebaseDocumentVersionsAttr(ctx, p, knowledgeDocumentId, knowledgeBaseId, version)
}

// createVariationFn is an implementation function for creating a Genesys Cloud variation request
func createVariationFn(ctx context.Context, p *variationRequestProxy, variationRequest *platformclientv2.Documentvariationrequest, knowledgeDocumentId, knowledgeBaseId string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	return p.knowledgeApi.PostKnowledgeKnowledgebaseDocumentVariations(knowledgeBaseId, knowledgeDocumentId, *variationRequest)
}

// getAllVariationsFn is the implementation for retrieving all variation request in Genesys Cloud
func getAllVariationsFn(ctx context.Context, p *variationRequestProxy) (*[]platformclientv2.Documentvariationrequest, *platformclientv2.APIResponse, error) {
	var allDocumentVariationRequests []platformclientv2.Documentvariationrequest
	const pageSize = 100

	documentVariationRequests, resp, err := p.knowledgeApi.GetKnowledgeKnowledgebaseDocumentVariations()
	if err != nil {
		return nil, resp, err
	}
	if documentVariationRequests.Entities == nil || len(*documentVariationRequests.Entities) == 0 {
		return &allDocumentVariationRequests, resp, nil
	}
	for _, documentVariationRequest := range *documentVariationRequests.Entities {
		allDocumentVariationRequests = append(allDocumentVariationRequests, documentVariationRequest)
	}

	for pageNum := 2; pageNum <= *documentVariationRequests.PageCount; pageNum++ {
		documentVariationRequests, _, err := p.knowledgeApi.GetKnowledgeKnowledgebaseDocumentVariations()
		if err != nil {
			return nil, resp, err
		}

		if documentVariationRequests.Entities == nil || len(*documentVariationRequests.Entities) == 0 {
			break
		}

		for _, documentVariationRequest := range *documentVariationRequests.Entities {
			allDocumentVariationRequests = append(allDocumentVariationRequests, documentVariationRequest)
		}
	}

	return &allDocumentVariationRequests, resp, nil
}

// getVariationRequestIdByNameFn is an implementation of the function to get a Genesys Cloud variation request by name
func getVariationRequestIdByNameFn(ctx context.Context, p *variationRequestProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	documentVariationRequests, resp, err := p.knowledgeApi.GetKnowledgeKnowledgebaseDocumentVariations()
	if err != nil {
		return "", resp, false, err
	}

	if documentVariationRequests.Entities == nil || len(*documentVariationRequests.Entities) == 0 {
		return "", resp, true, err
	}

	for _, documentVariationRequest := range *documentVariationRequests.Entities {
		if *documentVariationRequest.Name == name {
			log.Printf("Retrieved the variation request id %s by name %s", *documentVariationRequest.Id, name)
			return *documentVariationRequest.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("Unable to find variation request with name %s", name)
}

// getVariationRequestByIdFn is an implementation of the function to get a Genesys Cloud variation request by Id
func getVariationRequestByIdFn(ctx context.Context, p *variationRequestProxy, documentVariationId string, documentId string, knowledgeBaseId string, documentState string, expand []string) (*platformclientv2.Documentvariationresponse, *platformclientv2.APIResponse, error) {
	return p.knowledgeApi.GetKnowledgeKnowledgebaseDocumentVariation(documentVariationId, documentId, knowledgeBaseId, documentState, expand)
}

// updateVariationRequestFn is an implementation of the function to update a Genesys Cloud variation request
func updateVariationRequestFn(ctx context.Context, p *variationRequestProxy, id string, variationRequest *platformclientv2.Documentvariationrequest) (*platformclientv2.Documentvariationrequest, *platformclientv2.APIResponse, error) {
	return p.knowledgeApi.PatchKnowledgeKnowledgebaseDocumentVariation(id, *variationRequest)
}

// deleteVariationRequestFn is an implementation function for deleting a Genesys Cloud variation request
func deleteVariationRequestFn(ctx context.Context, p *variationRequestProxy, variationId, documentId, baseId string) (*platformclientv2.APIResponse, error) {
	return p.knowledgeApi.DeleteKnowledgeKnowledgebaseDocumentVariation(variationId, documentId, baseId)
}

func createKnowledgeKnowledgebaseDocumentVersionsFn(ctx context.Context, p *variationRequestProxy, knowledgeDocumentId, knowledgeBaseId string, version *platformclientv2.Knowledgedocumentversion) (*platformclientv2.Knowledgedocumentversion, *platformclientv2.APIResponse, error) {
	return p.knowledgeApi.PostKnowledgeKnowledgebaseDocumentVersions(knowledgeBaseId, knowledgeDocumentId, *version)
}
