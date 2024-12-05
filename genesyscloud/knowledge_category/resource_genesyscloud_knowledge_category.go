package knowledge_category

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllKnowledgeCategories(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	categoryEntities := make([]platformclientv2.Categoryresponse, 0)
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := GetKnowledgeCategoryProxy(clientConfig)

	// get published knowledge bases
	publishedEntities, resp, err := proxy.getAllKnowledgebaseEntities(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_categories", fmt.Sprintf("failed to get published knowledgebase entities: %s", err), resp)
	}
	knowledgeBaseList = append(knowledgeBaseList, *publishedEntities...)

	// get unpublished knowledge bases
	unpublishedEntities, resp, err := proxy.getAllKnowledgebaseEntities(ctx, false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_categories", fmt.Sprintf("failed to get unpublished knowledgebase entities: %s", err), resp)

	}
	knowledgeBaseList = append(knowledgeBaseList, *unpublishedEntities...)

	for _, knowledgeBase := range knowledgeBaseList {
		partialEntities, resp, err := proxy.getAllKnowledgeCategoryEntities(ctx, &knowledgeBase, "")
		if err != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_categories", fmt.Sprintf("failed to get all knowledgebase categories: %s", err), resp)
		}
		categoryEntities = append(categoryEntities, *partialEntities...)
	}

	for _, knowledgeCategory := range categoryEntities {
		id := fmt.Sprintf("%s,%s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id)
		resources[id] = &resourceExporter.ResourceMeta{BlockLabel: *knowledgeCategory.Name}
	}

	return resources, nil
}

func createKnowledgeCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	knowledgeCategory := d.Get("knowledge_category").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeCategoryProxy(sdkConfig)

	knowledgeCategoryRequest := buildKnowledgeCategoryCreate(knowledgeCategory)

	log.Printf("Creating knowledge category %s", knowledgeCategory["name"].(string))
	knowledgeCategoryResponse, resp, err := proxy.createKnowledgeCategory(ctx, knowledgeBaseId, *knowledgeCategoryRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to create knowledge category %s error: %s", d.Id(), err), resp)
	}

	id := fmt.Sprintf("%s,%s", *knowledgeCategoryResponse.Id, *knowledgeCategoryResponse.KnowledgeBase.Id)
	d.SetId(id)

	log.Printf("Created knowledge category %s", *knowledgeCategoryResponse.Id)
	return readKnowledgeCategory(ctx, d, meta)
}

func readKnowledgeCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeCategoryProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeCategory(), constants.ConsistencyChecks(), "genesyscloud_knowledge_category")

	log.Printf("Reading knowledge category %s", knowledgeCategoryId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeCategory, resp, getErr := proxy.getKnowledgeKnowledgebaseCategory(ctx, knowledgeBaseId, knowledgeCategoryId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to read knowledge category %s | error: %s", knowledgeCategoryId, getErr), resp))
			}
			log.Printf("%s", getErr)
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to read knowledge category %s | error: %s", knowledgeCategoryId, getErr), resp))
		}

		newId := fmt.Sprintf("%s,%s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id)
		d.SetId(newId)
		d.Set("knowledge_base_id", *knowledgeCategory.KnowledgeBase.Id)
		d.Set("knowledge_category", flattenKnowledgeCategory(*knowledgeCategory))
		log.Printf("Read knowledge category %s", knowledgeCategoryId)
		return cc.CheckState(d)
	})
}

func updateKnowledgeCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]
	knowledgeCategory := d.Get("knowledge_category").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeCategoryProxy(sdkConfig)

	log.Printf("Updating knowledge category %s", knowledgeCategory["name"].(string))
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge category version
		_, resp, getErr := proxy.getKnowledgeKnowledgebaseCategory(ctx, knowledgeBaseId, knowledgeCategoryId)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to read knowledge category %s error: %s", knowledgeCategoryId, getErr), resp)
		}

		knowledgeCategoryUpdate := buildKnowledgeCategoryUpdate(knowledgeCategory)

		log.Printf("Updating knowledge category %s", knowledgeCategory["name"].(string))
		_, resp, putErr := proxy.updateKnowledgeCategory(ctx, knowledgeBaseId, knowledgeCategoryId, *knowledgeCategoryUpdate)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to update knowledge category %s error: %s", knowledgeCategoryId, putErr), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated knowledge category %s %s", knowledgeCategory["name"].(string), knowledgeCategoryId)
	return readKnowledgeCategory(ctx, d, meta)
}

func deleteKnowledgeCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeCategoryProxy(sdkConfig)

	log.Printf("Deleting knowledge category %s", id)
	_, resp, err := proxy.deleteKnowledgeCategory(ctx, knowledgeBaseId, knowledgeCategoryId)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to delete knowledge category %s error: %s", id, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getKnowledgeKnowledgebaseCategory(ctx, knowledgeBaseId, knowledgeCategoryId)
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge category deleted
				log.Printf("Deleted knowledge category %s", knowledgeCategoryId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Error deleting knowledge category %s | error: %s", knowledgeCategoryId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Knowledge category %s still exists", knowledgeCategoryId), resp))
	})
}
