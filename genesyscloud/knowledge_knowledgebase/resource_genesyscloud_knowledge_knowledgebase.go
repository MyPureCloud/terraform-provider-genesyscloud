package knowledge_knowledgebase

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

func getAllKnowledgeKnowledgebases(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	knowledgebaseProxy := GetKnowledgebaseProxy(clientConfig)

	publishedEntities, resp, err := knowledgebaseProxy.getAllKnowledgebaseEntities(ctx, true)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), resp)
	}

	for _, knowledgeBase := range *publishedEntities {
		resources[*knowledgeBase.Id] = &resourceExporter.ResourceMeta{BlockLabel: *knowledgeBase.Name}
	}

	unpublishedEntities, resp, err := knowledgebaseProxy.getAllKnowledgebaseEntities(ctx, false)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, err.Error(), resp)
	}

	for _, knowledgeBase := range *unpublishedEntities {
		resources[*knowledgeBase.Id] = &resourceExporter.ResourceMeta{BlockLabel: *knowledgeBase.Name}
	}

	return resources, nil
}

func createKnowledgeKnowledgebase(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	coreLanguage := d.Get("core_language").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgebaseProxy := GetKnowledgebaseProxy(sdkConfig)

	knowledgebaseRequest := &platformclientv2.Knowledgebasecreaterequest{
		Name:         &name,
		Description:  &description,
		CoreLanguage: &coreLanguage,
	}

	log.Printf("Creating knowledge base %s", name)
	knowledgeBase, resp, err := knowledgebaseProxy.createKnowledgebase(ctx, knowledgebaseRequest)

	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create knowledge base %s error: %s", name, err), resp)
	}

	d.SetId(*knowledgeBase.Id)

	log.Printf("Created knowledge base %s %s", name, *knowledgeBase.Id)
	return readKnowledgeKnowledgebase(ctx, d, meta)
}

func readKnowledgeKnowledgebase(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgebaseProxy := GetKnowledgebaseProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeKnowledgebase(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading knowledge base %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeBase, resp, getErr := knowledgebaseProxy.getKnowledgebaseById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge base %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge base %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", knowledgeBase.Name)
		resourcedata.SetNillableValue(d, "description", knowledgeBase.Description)
		resourcedata.SetNillableValue(d, "core_language", knowledgeBase.CoreLanguage)
		resourcedata.SetNillableValue(d, "published", knowledgeBase.Published)
		log.Printf("Read knowledge base %s %s", d.Id(), *knowledgeBase.Name)
		return cc.CheckState(d)
	})
}

func updateKnowledgeKnowledgebase(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgebaseProxy := GetKnowledgebaseProxy(sdkConfig)

	log.Printf("Updating knowledge base %s", name)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge base version
		_, resp, getErr := knowledgebaseProxy.getKnowledgebaseById(ctx, d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read knowledge base %s error: %s", name, getErr), resp)
		}

		update := &platformclientv2.Knowledgebaseupdaterequest{
			Name:        &name,
			Description: &description,
		}

		log.Printf("Updating knowledge base %s", name)
		_, resp, putErr := knowledgebaseProxy.updateKnowledgebase(ctx, d.Id(), update)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update knowledge base %s error: %s", name, putErr), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated knowledge base %s %s", name, d.Id())
	return readKnowledgeKnowledgebase(ctx, d, meta)
}

func deleteKnowledgeKnowledgebase(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgebaseProxy := GetKnowledgebaseProxy(sdkConfig)

	log.Printf("Deleting knowledge base %s", name)
	_, resp, err := knowledgebaseProxy.deleteKnowledgebase(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete knowledge base %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := knowledgebaseProxy.getKnowledgebaseById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge base deleted
				log.Printf("Deleted Knowledge base %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting knowledge base %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Knowledge base %s still exists", d.Id()), resp))
	})
}
