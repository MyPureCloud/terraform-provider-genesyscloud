package knowledge_document_variation

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	knowledgeDocument "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/knowledge_document"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

const variationIdSeparator = " "

type resourceIDs struct {
	knowledgeDocumentVariationID    string
	knowledgeBaseID                 string
	knowledgeDocumentResourceDataID string
	knowledgeDocumentID             string
}

func getAllKnowledgeDocumentVariations(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	variationProxy := getVariationRequestProxy(clientConfig)
	knowledgeDocumentProxy := knowledgeDocument.GetKnowledgeDocumentProxy(clientConfig)

	// Get all the Knowledge Base Entities
	knowledgeBaseList, err := getAllKnowledgeBases(ctx, variationProxy)
	if err != nil {
		return nil, err
	}

	for _, knowledgeBase := range knowledgeBaseList {
		variationEntities, response, err := knowledgeDocumentProxy.GetAllKnowledgeDocumentEntities(ctx, &knowledgeBase)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), response)
		}

		// Retrieve the documents for each knowledge base
		for _, knowledgeDoc := range *variationEntities {
			// parse document state
			var documentState string

			if knowledgeDoc.State != nil && *knowledgeDoc.State != "" {
				isValidState := strings.EqualFold(*knowledgeDoc.State, "Published") || strings.EqualFold(*knowledgeDoc.State, "Draft")
				if isValidState {
					documentState = *knowledgeDoc.State
				}
			}

			// get the variations for each document
			knowledgeDocumentVariations, resp, getErr := variationProxy.getAllVariations(ctx, *knowledgeBase.Id, *knowledgeDoc.Id, documentState, nil)
			if getErr != nil {
				return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of knowledge document variations error: %v", getErr), resp)
			}

			if knowledgeDocumentVariations == nil || len(*knowledgeDocumentVariations) == 0 {
				break
			}

			for _, knowledgeDocumentVariation := range *knowledgeDocumentVariations {
				id := buildVariationId(*knowledgeBase.Id, knowledgeDocument.BuildDocumentResourceDataID(*knowledgeDoc.Id, *knowledgeBase.Id), *knowledgeDocumentVariation.Id)

				blockLabel := util.StringOrNil(knowledgeBase.Name) + "_" + util.StringOrNil(knowledgeDoc.Title)

				if knowledgeDocumentVariation.Name != nil && *knowledgeDocumentVariation.Name != "" {
					blockLabel = blockLabel + "_" + *knowledgeDocumentVariation.Name
				} else {
					blockLabel = blockLabel + "_" + *knowledgeDocumentVariation.Id
				}
				resources[id] = &resourceExporter.ResourceMeta{BlockLabel: blockLabel}
			}
		}
	}

	return resources, nil
}

func getAllKnowledgeBases(ctx context.Context, proxy *variationRequestProxy) ([]platformclientv2.Knowledgebase, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)

	// get published knowledge bases
	publishedEntities, response, err := proxy.GetAllKnowledgebaseEntities(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), response)
	}
	if publishedEntities != nil {
		knowledgeBaseList = append(knowledgeBaseList, *publishedEntities...)
	}

	// get unpublished knowledge bases
	unpublishedEntities, response, err := proxy.GetAllKnowledgebaseEntities(ctx, false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), response)
	}

	if unpublishedEntities != nil {
		knowledgeBaseList = append(knowledgeBaseList, *unpublishedEntities...)
	}

	return knowledgeBaseList, nil
}

func createKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	variationProxy := getVariationRequestProxy(sdkConfig)

	ids := getKnowledgeIdsFromResourceData(d)
	knowledgeDocumentVariation, _ := d.Get("knowledge_document_variation").([]interface{})[0].(map[string]interface{})

	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	knowledgeDocumentVariationRequest := buildKnowledgeDocumentVariation(knowledgeDocumentVariation)

	log.Printf("Creating knowledge document variation for document %s", ids.knowledgeDocumentID)

	knowledgeDocumentVariationResponse, resp, err := variationProxy.CreateVariation(ctx, knowledgeDocumentVariationRequest, ids.knowledgeDocumentID, ids.knowledgeBaseID)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create variation for knowledge document (%s) error: %s", ids.knowledgeDocumentID, err), resp)
	}

	if published {
		_, resp, versionErr := variationProxy.createKnowledgeKnowledgebaseDocumentVersions(ctx, ids.knowledgeDocumentID, ids.knowledgeBaseID, &platformclientv2.Knowledgedocumentversion{})
		if versionErr != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish knowledge document error: %s", err), resp)
		}
	}

	id := buildVariationId(ids.knowledgeBaseID, ids.knowledgeDocumentResourceDataID, *knowledgeDocumentVariationResponse.Id)
	d.SetId(id)

	log.Printf("Created knowledge document variation %s", *knowledgeDocumentVariationResponse.Id)
	return readKnowledgeDocumentVariation(ctx, d, meta)
}

func readKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		sdkConfig      = meta.(*provider.ProviderMeta).ClientConfig
		variationProxy = getVariationRequestProxy(sdkConfig)

		knowledgeDocVariation *platformclientv2.Documentvariationresponse
		apiResponse           *platformclientv2.APIResponse
	)

	ids, err := parseResourceIDs(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeDocumentVariation(), constants.ConsistencyChecks(), ResourceType)

	documentState := ""
	if published, ok := d.GetOk("published"); ok {
		if published == true {
			documentState = "Published"
		} else {
			documentState = "Draft"
		}
	}

	log.Printf("Reading knowledge document variation %s", ids.knowledgeDocumentVariationID)
	retryErr := util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeDocVariation, apiResponse, err = variationProxy.getVariationRequestByIdAndState(ctx, ids, documentState)
		if err != nil {
			if util.IsStatus404(apiResponse) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		if knowledgeDocVariation.Document == nil || knowledgeDocVariation.Document.KnowledgeBase == nil {
			return retry.NonRetryableError(fmt.Errorf("returned knowledge document variation '%s' did not include a Document object", ids.knowledgeDocumentVariationID))
		}

		newId := buildVariationId(*knowledgeDocVariation.Document.KnowledgeBase.Id, ids.knowledgeDocumentResourceDataID, *knowledgeDocVariation.Id)
		d.SetId(newId)

		_ = d.Set("knowledge_base_id", *knowledgeDocVariation.Document.KnowledgeBase.Id)
		_ = d.Set("knowledge_document_id", ids.knowledgeDocumentResourceDataID)
		_ = d.Set("knowledge_document_variation", flattenKnowledgeDocumentVariation(*knowledgeDocVariation))

		if knowledgeDocVariation.DocumentVersion != nil && knowledgeDocVariation.DocumentVersion.Id != nil && len(*knowledgeDocVariation.DocumentVersion.Id) > 0 {
			_ = d.Set("published", true)
		} else {
			_ = d.Set("published", false)
		}

		log.Printf("Read knowledge document variation %s", ids.knowledgeDocumentVariationID)
		return cc.CheckState(d)
	})
	if retryErr != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s: %v", ids.knowledgeDocumentVariationID, retryErr), apiResponse)
	}
	return nil
}

func updateKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	variationProxy := getVariationRequestProxy(sdkConfig)

	ids, err := parseResourceIDs(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	knowledgeDocumentVariation, _ := d.Get("knowledge_document_variation").([]interface{})[0].(map[string]interface{})

	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	log.Printf("Updating knowledge document variation %s", ids.knowledgeDocumentVariationID)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge document variation version
		_, resp, getErr := variationProxy.getVariationRequestById(ctx, ids.knowledgeDocumentVariationID, ids.knowledgeDocumentID, ids.knowledgeBaseID, "Draft", nil)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s error: %s", ids.knowledgeDocumentVariationID, getErr), resp)
		}

		knowledgeDocumentVariationUpdate := buildKnowledgeDocumentVariationUpdate(knowledgeDocumentVariation)

		_, resp, putErr := variationProxy.updateVariationRequest(ctx, ids.knowledgeDocumentVariationID, ids.knowledgeDocumentID, ids.knowledgeBaseID, *knowledgeDocumentVariationUpdate)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update knowledge document variation %s error: %s", ids.knowledgeDocumentVariationID, putErr), resp)
		}
		if published {
			_, resp, versionErr := variationProxy.createKnowledgeKnowledgebaseDocumentVersions(ctx, ids.knowledgeDocumentID, ids.knowledgeBaseID, &platformclientv2.Knowledgedocumentversion{})

			if versionErr != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish knowledge document %s error: %s", ids.knowledgeDocumentID, versionErr), resp)
			}
		}

		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated knowledge document variation %s", ids.knowledgeDocumentVariationID)
	return readKnowledgeDocumentVariation(ctx, d, meta)
}

func deleteKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	variationProxy := getVariationRequestProxy(sdkConfig)

	ids, err := parseResourceIDs(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	log.Printf("Deleting knowledge document variation %s", ids.knowledgeDocumentVariationID)

	resp, err := variationProxy.deleteVariationRequest(ctx, ids.knowledgeDocumentVariationID, ids.knowledgeDocumentID, ids.knowledgeBaseID)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete knowledge document variation %s error: %s", ids.knowledgeDocumentVariationID, err), resp)
	}

	if published {
		/*
		 * If the published flag is set, attempt to publish a new document version without the variation.
		 * However a document cannot be published if it has no variations, so first check that the document has other variations
		 * A new document version can only be published if there are other variations than the one being removed
		 */

		variations, resp, variationErr := variationProxy.getAllVariations(ctx, ids.knowledgeBaseID, ids.knowledgeDocumentID, "Draft", nil)
		if variationErr != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to retrieve knowledge document variations error: %s", err), resp)
		}

		if variations != nil && len(*variations) > 0 {
			_, resp, versionErr := variationProxy.createKnowledgeKnowledgebaseDocumentVersions(ctx, ids.knowledgeBaseID, ids.knowledgeDocumentID, &platformclientv2.Knowledgedocumentversion{})

			if versionErr != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish knowledge document error: %s", err), resp)
			}
		}
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		// The DELETE resource for knowledge document variations only removes draft variations. So set the documentState param to "Draft" for the check
		_, resp, err := variationProxy.getVariationRequestById(ctx, ids.knowledgeDocumentVariationID, ids.knowledgeDocumentID, ids.knowledgeBaseID, "Draft", nil)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted knowledge document variation %s", ids.knowledgeDocumentVariationID)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting knowledge document variation %s | error: %s", ids.knowledgeDocumentVariationID, err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Knowledge document variation %s still exists", ids.knowledgeDocumentVariationID), resp))
	})
}
