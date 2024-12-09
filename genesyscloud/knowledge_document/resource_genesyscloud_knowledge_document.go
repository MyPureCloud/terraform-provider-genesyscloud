package knowledge_document

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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllKnowledgeDocuments(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	documentEntities := make([]platformclientv2.Knowledgedocumentresponse, 0)
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := GetKnowledgeDocumentProxy(clientConfig)
	//knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)

	// get published knowledge bases
	publishedEntities, response, err := proxy.GetAllKnowledgebaseEntities(ctx, true) //getAllKnowledgebaseEntities(*knowledgeAPI, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("%s", err), response)

	}
	knowledgeBaseList = append(knowledgeBaseList, *publishedEntities...)

	// get unpublished knowledge bases
	unpublishedEntities, response, err := proxy.GetAllKnowledgebaseEntities(ctx, false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("%s", err), response)
	}
	knowledgeBaseList = append(knowledgeBaseList, *unpublishedEntities...)

	for _, knowledgeBase := range knowledgeBaseList {
		partialEntities, response, err := proxy.GetAllKnowledgeDocumentEntities(ctx, &knowledgeBase)
		if err != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("%s", err), response)
		}
		documentEntities = append(documentEntities, *partialEntities...)
	}

	for _, knowledgeDocument := range documentEntities {
		id := fmt.Sprintf("%s,%s", *knowledgeDocument.Id, *knowledgeDocument.KnowledgeBase.Id)
		resources[id] = &resourceExporter.ResourceMeta{BlockLabel: *knowledgeDocument.Title}
	}

	return resources, nil
}

func createKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	published := d.Get("published").(bool)
	proxy := GetKnowledgeDocumentProxy(sdkConfig)

	body, buildErr := buildKnowledgeDocumentCreateRequest(ctx, d, proxy, knowledgeBaseId)
	if buildErr != nil {
		return buildErr
	}

	log.Printf("Creating knowledge document")
	knowledgeDocument, resp, err := proxy.createKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, body)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to create knowledge document %s error: %s", d.Id(), err), resp)
	}

	if published {
		_, resp, versionErr := proxy.createKnowledgebaseDocumentVersions(ctx, knowledgeBaseId, *knowledgeDocument.Id, &platformclientv2.Knowledgedocumentversion{})
		if versionErr != nil {
			_, deleteError := proxy.deleteKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, *knowledgeDocument.Id)
			if deleteError != nil {
				log.Printf("failed to delete draft knowledge document %s error: %s", *knowledgeDocument.Id, deleteError)
			}
			return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to publish knowledge document error: %s", versionErr), resp)
		}
	}

	id := fmt.Sprintf("%s,%s", *knowledgeDocument.Id, knowledgeBaseId)
	d.SetId(id)

	log.Printf("Created knowledge document %s", *knowledgeDocument.Id)
	return readKnowledgeDocument(ctx, d, meta)
}

func readKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := id[1]
	state := "Draft"
	if d.Get("published").(bool) == true {
		state = "Published"
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeDocumentProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeDocument(), constants.ConsistencyChecks(), "genesyscloud_knowledge_document")

	log.Printf("Reading knowledge document %s", knowledgeDocumentId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeDocument, resp, getErr := proxy.getKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, knowledgeDocumentId, nil, state)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to read knowledge document %s: %s", knowledgeDocumentId, getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to read knowledge document %s: %s", knowledgeDocumentId, getErr), resp))
		}

		// required
		id := fmt.Sprintf("%s,%s", *knowledgeDocument.Id, knowledgeBaseId)
		d.SetId(id)
		d.Set("knowledge_base_id", *knowledgeDocument.KnowledgeBase.Id)

		flattenedDocument, err := flattenKnowledgeDocument(ctx, knowledgeDocument, proxy, knowledgeBaseId)
		if err != nil {
			return retry.NonRetryableError(err)
		}
		d.Set("knowledge_document", flattenedDocument)

		if *knowledgeDocument.State == "Published" {
			d.Set("published", true)
		} else {
			d.Set("published", false)
		}

		log.Printf("Read Knowledge document %s", *knowledgeDocument.Id)
		checkState := cc.CheckState(d)
		return checkState
	})
}

func updateKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	state := "Draft"
	if d.Get("published").(bool) == true {
		state = "Published"
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeDocumentProxy(sdkConfig)

	log.Printf("Updating Knowledge document %s", knowledgeDocumentId)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current Knowledge document version
		_, resp, getErr := proxy.getKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, knowledgeDocumentId, nil, state)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to read knowledge document %s error: %s", knowledgeDocumentId, getErr), resp)
		}

		update, err := buildKnowledgeDocumentRequest(ctx, d, proxy, knowledgeBaseId)
		if err != nil {
			return nil, err
		}

		log.Printf("Updating knowledge document %s", knowledgeDocumentId)
		_, resp, putErr := proxy.updateKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, knowledgeDocumentId, update)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to update knowledge document %s error: %s", knowledgeDocumentId, putErr), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated Knowledge document %s", knowledgeDocumentId)
	return readKnowledgeDocument(ctx, d, meta)
}

func deleteKnowledgeDocument(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeDocumentId := id[0]
	knowledgeBaseId := id[1]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := GetKnowledgeDocumentProxy(sdkConfig)

	log.Printf("Deleting Knowledge document %s", knowledgeDocumentId)
	resp, err := proxy.deleteKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, knowledgeDocumentId)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Failed to delete knowledge document %s error: %s", knowledgeDocumentId, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		state := "Draft"
		if d.Get("published").(bool) == true {
			state = "Published"
		}

		_, resp, err := proxy.getKnowledgeKnowledgebaseDocument(ctx, knowledgeBaseId, knowledgeDocumentId, nil, state)
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge document deleted
				log.Printf("Deleted Knowledge document %s", knowledgeDocumentId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Error deleting Knowledge document %s | error: %s", knowledgeDocumentId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_document", fmt.Sprintf("Knowledge document %s still exists", knowledgeDocumentId), resp))
	})
}
