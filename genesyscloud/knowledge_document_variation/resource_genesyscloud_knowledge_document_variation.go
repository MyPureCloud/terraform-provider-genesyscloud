package knowledgedocumentvariation

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	knowledgeDocument "terraform-provider-genesyscloud/genesyscloud/knowledge_document"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

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
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("%v", err), response)
		}

		// Retrieve the documents for each knowledge base
		for _, knowledgeDoc := range *variationEntities {
			// parse document state
			var documentState string
			isValidState := strings.EqualFold(*knowledgeDoc.State, "Published") || strings.EqualFold(*knowledgeDoc.State, "Draft")
			if isValidState {
				documentState = *knowledgeDoc.State
			}

			// get the variations for each document
			knowledgeDocumentVariations, resp, getErr := variationProxy.getAllVariations(ctx, *knowledgeBase.Id, *knowledgeDoc.Id, documentState, nil)
			if getErr != nil {
				return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of knowledge document variations error: %v", err), resp)
			}

			if knowledgeDocumentVariations == nil || len(*knowledgeDocumentVariations) == 0 {
				break
			}

			for _, knowledgeDocumentVariation := range *knowledgeDocumentVariations {
				id := fmt.Sprintf("%s %s %s", *knowledgeDocumentVariation.Id, *knowledgeDoc.KnowledgeBase.Id, *knowledgeDoc.Id)
				blockLabel := *knowledgeBase.Name + "_" + *knowledgeDoc.Title
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
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("%v", err), response)
	}
	knowledgeBaseList = append(knowledgeBaseList, *publishedEntities...)

	// get unpublished knowledge bases
	unpublishedEntities, response, err := proxy.GetAllKnowledgebaseEntities(ctx, false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("%v", err), response)
	}
	knowledgeBaseList = append(knowledgeBaseList, *unpublishedEntities...)

	return knowledgeBaseList, nil
}

func createKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	variationProxy := getVariationRequestProxy(sdkConfig)

	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	documentResourceId := d.Get("knowledge_document_id").(string)
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]
	knowledgeDocumentVariation := d.Get("knowledge_document_variation").([]interface{})[0].(map[string]interface{})

	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	knowledgeDocumentVariationRequest := buildKnowledgeDocumentVariation(knowledgeDocumentVariation)

	log.Printf("Creating knowledge document variation for document %s", knowledgeDocumentId)

	knowledgeDocumentVariationResponse, resp, err := variationProxy.CreateVariation(ctx, knowledgeDocumentVariationRequest, knowledgeDocumentId, knowledgeBaseId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create variation for knowledge document %s error: %s", d.Id(), err), resp)
	}

	if published == true {
		_, resp, versionErr := variationProxy.createKnowledgeKnowledgebaseDocumentVersions(ctx, knowledgeDocumentId, knowledgeBaseId, &platformclientv2.Knowledgedocumentversion{})
		if versionErr != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish knowledge document error: %s", err), resp)
		}
	}

	id := fmt.Sprintf("%s %s %s", *knowledgeDocumentVariationResponse.Id, knowledgeBaseId, documentResourceId)
	d.SetId(id)

	log.Printf("Created knowledge document variation %s", *knowledgeDocumentVariationResponse.Id)
	return readKnowledgeDocumentVariation(ctx, d, meta)
}

func readKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	variationProxy := getVariationRequestProxy(sdkConfig)

	id := strings.Split(d.Id(), " ")
	documentVariationId := id[0]
	knowledgeBaseId := id[1]
	documentResourceId := id[2]
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]

	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeDocumentVariation(), constants.ConsistencyChecks(), ResourceType)

	documentState := ""

	// If the published flag is set, use it to set documentState param
	if published, ok := d.GetOk("published"); ok {
		if published == true {
			documentState = "Published"
		} else {
			documentState = "Draft"
		}
	}

	log.Printf("Reading knowledge document variation %s", documentVariationId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		var knowledgeDocumentVariation *platformclientv2.Documentvariationresponse
		/*
		 * If the published flag is not set, get both published and draft variation and choose the most recent
		 * If it is set, base the document state param off the flag.
		 * The published flag has to be optional for the import case, where only the resource ID is available.
		 * If the published flag were required, it would cause consistency issues for the import state.
		 */
		if documentState == "" {
			publishedVariation, resp, publishedErr := variationProxy.getVariationRequestById(ctx, documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Published", nil)

			if publishedErr != nil {
				// Published version may or may not exist, so if status is 404, sleep and retry once and then move on to retrieve draft variation.
				if util.IsStatus404(resp) {
					time.Sleep(2 * time.Second)
					retryVariation, retryResp, retryErr := variationProxy.getVariationRequestById(ctx, documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Published", nil)

					if retryErr != nil {
						if !util.IsStatus404(retryResp) {
							log.Printf("%s", retryErr)
							return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, retryErr), retryResp))
						}
					} else {
						publishedVariation = retryVariation
					}

				} else {
					log.Printf("%s", publishedErr)
					return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, publishedErr), resp))
				}
			}

			draftVariation, resp, draftErr := variationProxy.getVariationRequestById(ctx, documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft", nil)
			if draftErr != nil {
				if util.IsStatus404(resp) {
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, draftErr), resp))
				}
				log.Printf("%s", draftErr)
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s: %s", documentVariationId, draftErr), resp))
			}

			if publishedVariation != nil && publishedVariation.DateModified != nil && publishedVariation.DateModified.After(*draftVariation.DateModified) {
				knowledgeDocumentVariation = publishedVariation
			} else {
				knowledgeDocumentVariation = draftVariation
			}
		} else {
			variation, resp, getErr := variationProxy.getVariationRequestById(ctx, documentVariationId, knowledgeDocumentId, knowledgeBaseId, documentState, nil)
			if getErr != nil {
				if util.IsStatus404(resp) {
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s | error: %s", documentVariationId, getErr), resp))
				}
				log.Printf("%s", getErr)
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s | error: %s", documentVariationId, getErr), resp))
			}

			knowledgeDocumentVariation = variation
		}

		newId := fmt.Sprintf("%s %s %s", *knowledgeDocumentVariation.Id, *knowledgeDocumentVariation.Document.KnowledgeBase.Id, documentResourceId)
		d.SetId(newId)
		_ = d.Set("knowledge_base_id", *knowledgeDocumentVariation.Document.KnowledgeBase.Id)
		_ = d.Set("knowledge_document_id", documentResourceId)
		_ = d.Set("knowledge_document_variation", flattenKnowledgeDocumentVariation(*knowledgeDocumentVariation))

		if knowledgeDocumentVariation.DocumentVersion != nil && knowledgeDocumentVariation.DocumentVersion.Id != nil && len(*knowledgeDocumentVariation.DocumentVersion.Id) > 0 {
			_ = d.Set("published", true)
		} else {
			_ = d.Set("published", false)
		}

		log.Printf("Read knowledge document variation %s", documentVariationId)
		return cc.CheckState(d)
	})
}

func updateKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	variationProxy := getVariationRequestProxy(sdkConfig)

	id := strings.Split(d.Id(), " ")
	documentVariationId := id[0]
	knowledgeBaseId := id[1]
	documentResourceId := id[2]
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]
	knowledgeDocumentVariation := d.Get("knowledge_document_variation").([]interface{})[0].(map[string]interface{})

	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	log.Printf("Updating knowledge document variation %s", documentVariationId)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge document variation version
		_, resp, getErr := variationProxy.getVariationRequestById(ctx, documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft", nil)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document variation %s error: %s", id, getErr), resp)
		}

		knowledgeDocumentVariationUpdate := buildKnowledgeDocumentVariationUpdate(knowledgeDocumentVariation)

		_, resp, putErr := variationProxy.updateVariationRequest(ctx, documentVariationId, knowledgeDocumentId, knowledgeBaseId, *knowledgeDocumentVariationUpdate)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update knowledge document variation %s error: %s", documentVariationId, putErr), resp)
		}
		if published == true {
			_, resp, versionErr := variationProxy.createKnowledgeKnowledgebaseDocumentVersions(ctx, knowledgeDocumentId, knowledgeBaseId, &platformclientv2.Knowledgedocumentversion{})

			if versionErr != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish knowledge document %s error: %s", id, versionErr), resp)
			}
		}

		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated knowledge document variation %s", documentVariationId)
	return readKnowledgeDocumentVariation(ctx, d, meta)
}

func deleteKnowledgeDocumentVariation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	variationProxy := getVariationRequestProxy(sdkConfig)

	id := strings.Split(d.Id(), " ")
	documentVariationId := id[0]
	knowledgeBaseId := id[1]
	documentResourceId := id[2]
	knowledgeDocumentId := strings.Split(documentResourceId, ",")[0]

	published := false
	if publishedIn, ok := d.GetOk("published"); ok {
		published = publishedIn.(bool)
	}

	log.Printf("Deleting knowledge document variation %s", id)

	resp, err := variationProxy.deleteVariationRequest(ctx, documentVariationId, knowledgeDocumentId, knowledgeBaseId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete knowledge document variation %s error: %s", id, err), resp)
	}

	if published == true {
		/*
		 * If the published flag is set, attempt to publish a new document version without the variation.
		 * However a document cannot be published if it has no variations, so first check that the document has other variations
		 * A new document version can only be published if there are other variations than the one being removed
		 */

		variations, resp, variationErr := variationProxy.getAllVariations(ctx, knowledgeBaseId, knowledgeDocumentId, "Draft", nil)
		if variationErr != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to retrieve knowledge document variations error: %s", err), resp)
		}

		if len(*variations) > 0 {
			_, resp, versionErr := variationProxy.createKnowledgeKnowledgebaseDocumentVersions(ctx, knowledgeBaseId, knowledgeDocumentId, &platformclientv2.Knowledgedocumentversion{})

			if versionErr != nil {
				return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish knowledge document error: %s", err), resp)
			}
		}
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		// The DELETE resource for knowledge document variations only removes draft variations. So set the documentState param to "Draft" for the check
		_, resp, err := variationProxy.getVariationRequestById(ctx, documentVariationId, knowledgeDocumentId, knowledgeBaseId, "Draft", nil)
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted knowledge document variation %s", documentVariationId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting knowledge document variation %s | error: %s", documentVariationId, err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Knowledge document variation %s still exists", documentVariationId), resp))
	})
}
