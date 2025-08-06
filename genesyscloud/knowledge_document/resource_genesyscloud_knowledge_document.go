package knowledge_document

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	consistencyChecker "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func getAllKnowledgeDocuments(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := GetKnowledgeDocumentProxy(clientConfig)

	// get published knowledge bases
	publishedEntities, response, err := proxy.GetAllKnowledgebaseEntities(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), response)
	}
	if publishedEntities != nil && len(*publishedEntities) > 0 {
		knowledgeBaseList = append(knowledgeBaseList, *publishedEntities...)
	}

	// get unpublished knowledge bases
	unpublishedEntities, response, err := proxy.GetAllKnowledgebaseEntities(ctx, false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), response)
	}
	if unpublishedEntities != nil && len(*unpublishedEntities) > 0 {
		knowledgeBaseList = append(knowledgeBaseList, *unpublishedEntities...)
	}

	for _, knowledgeBase := range knowledgeBaseList {
		partialEntities, response, err := proxy.GetAllKnowledgeDocumentEntities(ctx, &knowledgeBase)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), response)
		}
		for _, knowledgeDocument := range *partialEntities {
			blockHash := ""
			if knowledgeDocument.Category != nil && knowledgeDocument.Category.Id != nil {
				category, _, err := proxy.getKnowledgeKnowledgebaseCategory(ctx, *knowledgeBase.Id, *knowledgeDocument.Category.Id)
				if err != nil {
					return nil, diag.Errorf("error reading knowledge document %s category %s: %s", *knowledgeDocument.Id, *knowledgeDocument.Category.Id, err)

				}
				blockHash, err = util.QuickHashFields(*category.Name)
				if err != nil {
					return nil, diag.Errorf("error hashing knowledge document %s: %s", *knowledgeDocument.Id, err)
				}
			}
			id := BuildDocumentResourceDataID(*knowledgeDocument.Id, *knowledgeBase.Id)
			resources[id] = &resourceExporter.ResourceMeta{BlockLabel: *knowledgeBase.Name + "_" + *knowledgeDocument.Title, BlockHash: blockHash}
		}

	}

	return resources, nil
}

func createKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	proxy := GetKnowledgeDocumentProxy(sdkConfig)

	body, buildErr := buildKnowledgeDocumentCreateRequest(ctx, d, proxy, knowledgeBaseId)
	if buildErr != nil {
		diags = append(diags, buildErr...)
		if diags.HasError() {
			return diags
		}
	}

	log.Printf("Creating knowledge document for knowledge base '%s'. Title: '%s'", knowledgeBaseId, *body.Title)
	knowledgeDocument, resp, err := proxy.createKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, body)
	if err != nil {
		createDiagErr := util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create knowledge document for knowledge base '%s', title '%s'. Error: %s", knowledgeBaseId, *body.Title, err.Error()), resp)
		log.Println(createDiagErr)
		diags = append(diags, createDiagErr...)
		return diags
	}

	id := BuildDocumentResourceDataID(*knowledgeDocument.Id, knowledgeBaseId)
	d.SetId(id)

	log.Printf("Created knowledge document %s, title '%s'. Resource schema ID: '%s'", *knowledgeDocument.Id, *body.Title, id)
	return append(diags, readKnowledgeDocument(ctx, d, meta)...)
}

func readKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	knowledgeDocumentId, knowledgeBaseId := parseDocumentResourceDataID(d.Id())
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeDocumentProxy(sdkConfig)
	cc := consistencyChecker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeDocument(), constants.ConsistencyChecks(), ResourceType)

	state := ""
	if !d.Get("published").(bool) {
		state = "Draft"
	}

	log.Printf("Reading knowledge document '%s'", knowledgeDocumentId)
	retryErr := util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		log.Printf("Reading knowledge document '%s'. Knowledge base '%s'", knowledgeDocumentId, knowledgeBaseId)
		knowledgeDocument, resp, getErr := proxy.getKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, knowledgeDocumentId, nil, state)
		if getErr != nil {
			err := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge document %s: %s", knowledgeDocumentId, getErr), resp)
			log.Println(err)
			if util.IsStatus404(resp) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		// required
		id := BuildDocumentResourceDataID(*knowledgeDocument.Id, knowledgeBaseId)
		if id != d.Id() {
			log.Printf("[WARN] Concatenated ID after read is not the same as what was in the resource data schema. Original: %s. New: %s", d.Id(), id)
		}
		d.SetId(id)

		if knowledgeDocument.KnowledgeBase != nil && knowledgeDocument.KnowledgeBase.Id != nil {
			_ = d.Set("knowledge_base_id", *knowledgeDocument.KnowledgeBase.Id)
		}

		log.Printf("Flattening knowledge document schema for document '%s'", d.Id())
		flattenedDocument, err := flattenKnowledgeDocument(ctx, knowledgeDocument, proxy, knowledgeBaseId)
		if err != nil {
			log.Printf("Failed to flatten knowledge document schema for document '%s': %s", d.Id(), err.Error())
			return retry.NonRetryableError(err)
		}

		_ = d.Set("knowledge_document", flattenedDocument)

		log.Printf("Read Knowledge document %s", *knowledgeDocument.Id)
		return cc.CheckState(d)
	})

	if retryErr != nil {
		diags = append(diags, retryErr...)
	}

	return diags
}

func updateKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	knowledgeDocumentId, _ := parseDocumentResourceDataID(d.Id())
	knowledgeBaseId := d.Get("knowledge_base_id").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeDocumentProxy(sdkConfig)

	updateErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		update, err := buildKnowledgeDocumentRequest(ctx, d, proxy, knowledgeBaseId)
		if err != nil {
			return nil, err
		}

		log.Printf("Updating knowledge document '%s'. Knowledge Base ID: '%s'", knowledgeDocumentId, knowledgeBaseId)
		_, resp, putErr := proxy.updateKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, knowledgeDocumentId, update)
		if putErr != nil {
			updateDiagErr := util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update knowledge document '%s'. Knowledge Base ID: '%s'. Error: %s", knowledgeDocumentId, knowledgeBaseId, putErr), resp)
			log.Println(updateDiagErr)
			return resp, updateDiagErr
		}

		return resp, nil
	})

	if updateErr != nil {
		diags = append(diags, updateErr...)
		if diags.HasError() {
			return diags
		}
	}

	_ = d.Set("published", false)
	log.Printf("Succesfully updated knowledge document '%s'. Knowledge base: '%s'", knowledgeDocumentId, knowledgeBaseId)
	return append(diags, readKnowledgeDocument(ctx, d, meta)...)
}

func deleteKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := id[1]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeDocumentProxy(sdkConfig)

	log.Printf("Deleting Knowledge document '%s'. Knowledge base ID: '%s'", knowledgeDocumentId, knowledgeBaseId)
	resp, err := proxy.deleteKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, knowledgeDocumentId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete knowledge document %s error: %s", knowledgeDocumentId, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err = proxy.getKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, knowledgeDocumentId, nil, "")
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge document deleted
				log.Printf("Deleted Knowledge document %s", knowledgeDocumentId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting Knowledge document '%s' | error: %s", knowledgeDocumentId, err.Error()), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Knowledge document '%s' still exists", knowledgeDocumentId), resp))
	})
}
