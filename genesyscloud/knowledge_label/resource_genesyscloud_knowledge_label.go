package knowledge_label

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
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

func getAllKnowledgeLabels(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	resources := make(resourceExporter.ResourceIDMetaMap)

	proxy := GetKnowledgeLabelProxy(clientConfig)

	// get published knowledge bases
	publishedEntities, response, err := proxy.GetAllKnowledgebaseEntities(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), response)

	}
	knowledgeBaseList = append(knowledgeBaseList, *publishedEntities...)

	// get unpublished knowledge bases
	unpublishedEntities, response, err := proxy.GetAllKnowledgebaseEntities(ctx, false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), response)

	}
	knowledgeBaseList = append(knowledgeBaseList, *unpublishedEntities...)

	for _, knowledgeBase := range knowledgeBaseList {
		partialEntities, response, err := proxy.GetAllKnowledgeLabelEntities(ctx, &knowledgeBase)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), response)

		}

		for _, knowledgeLabel := range *partialEntities {
			id := fmt.Sprintf("%s,%s", *knowledgeLabel.Id, *knowledgeBase.Id)
			resources[id] = &resourceExporter.ResourceMeta{BlockLabel: *knowledgeLabel.Name}
		}
	}

	return resources, nil
}

func createKnowledgeLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	knowledgeLabel := d.Get("knowledge_label").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeLabelProxy(sdkConfig)

	knowledgeLabelRequest := buildKnowledgeLabel(knowledgeLabel)

	log.Printf("Creating knowledge label %s", knowledgeLabel["name"].(string))
	knowledgeLabelResponse, resp, err := proxy.createKnowledgeLabel(ctx, knowledgeBaseId, &knowledgeLabelRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create knowledge label %s error: %s", knowledgeBaseId, err), resp)
	}

	id := fmt.Sprintf("%s,%s", *knowledgeLabelResponse.Id, knowledgeBaseId)
	d.SetId(id)

	log.Printf("Created knowledge label %s", *knowledgeLabelResponse.Id)
	return readKnowledgeLabel(ctx, d, meta)
}

func readKnowledgeLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeLabelId := id[0]
	knowledgeBaseId := id[1]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeLabelProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeLabel(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading knowledge label %s", knowledgeLabelId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeLabel, resp, getErr := proxy.getKnowledgeLabel(ctx, knowledgeBaseId, knowledgeLabelId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge label %s | error: %s", knowledgeLabelId, getErr), resp))
			}
			log.Printf("%s", getErr)
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge label %s | error: %s", knowledgeLabelId, getErr), resp))
		}

		newId := fmt.Sprintf("%s,%s", *knowledgeLabel.Id, knowledgeBaseId)
		d.SetId(newId)
		d.Set("knowledge_base_id", knowledgeBaseId)
		d.Set("knowledge_label", flattenKnowledgeLabel(knowledgeLabel))
		log.Printf("Read knowledge label %s", knowledgeLabelId)
		return cc.CheckState(d)
	})
}

func updateKnowledgeLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeLabelId := id[0]
	knowledgeBaseId := id[1]
	knowledgeLabel := d.Get("knowledge_label").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeLabelProxy(sdkConfig)

	log.Printf("Updating knowledge label %s", knowledgeLabel["name"].(string))
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge label version
		_, resp, getErr := proxy.getKnowledgeLabel(ctx, knowledgeBaseId, knowledgeLabelId)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge label %s error: %s", knowledgeLabelId, getErr), resp)
		}

		knowledgeLabelUpdate := buildKnowledgeLabelUpdate(knowledgeLabel)

		log.Printf("Updating knowledge label %s", knowledgeLabel["name"].(string))
		_, resp, putErr := proxy.updateKnowledgeLabel(ctx, knowledgeBaseId, knowledgeLabelId, &knowledgeLabelUpdate)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update knowledge label %s error: %s", knowledgeLabelId, putErr), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated knowledge label %s %s", knowledgeLabel["name"].(string), knowledgeLabelId)
	return readKnowledgeLabel(ctx, d, meta)
}

func deleteKnowledgeLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeLabelId := id[0]
	knowledgeBaseId := id[1]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeLabelProxy(sdkConfig)

	log.Printf("Deleting knowledge label %s", id)
	_, resp, err := proxy.deleteKnowledgeLabel(ctx, knowledgeBaseId, knowledgeLabelId)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete knowledge label %s error: %s", id, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getKnowledgeLabel(ctx, knowledgeBaseId, knowledgeLabelId)
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge label deleted
				log.Printf("Deleted knowledge label %s", knowledgeLabelId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting knowledge label %s | error: %s", knowledgeLabelId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Knowledge label %s still exists", knowledgeLabelId), resp))
	})
}
