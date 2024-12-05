package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"terraform-provider-genesyscloud/genesyscloud/validators"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllKnowledgeKnowledgebases(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)

	publishedEntities, err := getAllKnowledgebaseEntities(*knowledgeAPI, true)
	if err != nil {
		return nil, err
	}

	for _, knowledgeBase := range *publishedEntities {
		resources[*knowledgeBase.Id] = &resourceExporter.ResourceMeta{BlockLabel: *knowledgeBase.Name}
	}

	unpublishedEntities, err := getAllKnowledgebaseEntities(*knowledgeAPI, false)
	if err != nil {
		return nil, err
	}

	for _, knowledgeBase := range *unpublishedEntities {
		resources[*knowledgeBase.Id] = &resourceExporter.ResourceMeta{BlockLabel: *knowledgeBase.Name}
	}

	return resources, nil
}

func getAllKnowledgebaseEntities(knowledgeApi platformclientv2.KnowledgeApi, published bool) (*[]platformclientv2.Knowledgebase, diag.Diagnostics) {
	var (
		after    string
		err      error
		entities []platformclientv2.Knowledgebase
	)

	const pageSize = 100
	for {
		knowledgeBases, resp, getErr := knowledgeApi.GetKnowledgeKnowledgebases("", after, "", fmt.Sprintf("%v", pageSize), "", "", published, "", "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Failed to get page of knowledge bases error: %s", getErr), resp)
		}

		if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeBases.Entities...)

		if knowledgeBases.NextUri == nil || *knowledgeBases.NextUri == "" {
			break
		}

		after, err = util.GetQueryParamValueFromUri(*knowledgeBases.NextUri, "after")
		if err != nil {
			return nil, util.BuildDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Failed to parse after cursor from knowledge base nextUri"), err)
		}
		if after == "" {
			break
		}
	}

	return &entities, nil
}

func KnowledgeKnowledgebaseExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllKnowledgeKnowledgebases),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceKnowledgeKnowledgebase() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Base",

		CreateContext: provider.CreateWithPooledClient(createKnowledgeKnowledgebase),
		ReadContext:   provider.ReadWithPooledClient(readKnowledgeKnowledgebase),
		UpdateContext: provider.UpdateWithPooledClient(updateKnowledgeKnowledgebase),
		DeleteContext: provider.DeleteWithPooledClient(deleteKnowledgeKnowledgebase),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Knowledge base name",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "Knowledge base description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"core_language": {
				Description:      "Core language for knowledge base in which initial content must be created, language codes [en-US, en-UK, en-AU, de-DE] are supported currently, however the new DX knowledge will support all these language codes",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateLanguageCode,
			},
			"published": {
				Description: "Flag that indicates the knowledge base is published",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func createKnowledgeKnowledgebase(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	coreLanguage := d.Get("core_language").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Creating knowledge base %s", name)
	knowledgeBase, resp, err := knowledgeAPI.PostKnowledgeKnowledgebases(platformclientv2.Knowledgebasecreaterequest{
		Name:         &name,
		Description:  &description,
		CoreLanguage: &coreLanguage,
	})

	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Failed to create knowledge base %s error: %s", name, err), resp)
	}

	d.SetId(*knowledgeBase.Id)

	log.Printf("Created knowledge base %s %s", name, *knowledgeBase.Id)
	return readKnowledgeKnowledgebase(ctx, d, meta)
}

func readKnowledgeKnowledgebase(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeKnowledgebase(), constants.ConsistencyChecks(), "genesyscloud_knowledge_knowledgebase")

	log.Printf("Reading knowledge base %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeBase, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebase(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Failed to read knowledge base %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Failed to read knowledge base %s | error: %s", d.Id(), getErr), resp))
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
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating knowledge base %s", name)
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge base version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebase(d.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Failed to read knowledge base %s error: %s", name, getErr), resp)
		}

		update := platformclientv2.Knowledgebaseupdaterequest{
			Name:        &name,
			Description: &description,
		}

		log.Printf("Updating knowledge base %s", name)
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebase(d.Id(), update)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Failed to update knowledge base %s error: %s", name, putErr), resp)
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
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting knowledge base %s", name)
	_, resp, err := knowledgeAPI.DeleteKnowledgeKnowledgebase(d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Failed to delete knowledge base %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebase(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge base deleted
				log.Printf("Deleted Knowledge base %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Error deleting knowledge base %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_knowledgebase", fmt.Sprintf("Knowledge base %s still exists", d.Id()), resp))
	})
}

func GenerateKnowledgeKnowledgebaseResource(
	resourceLabel string,
	name string,
	description string,
	coreLanguage string) string {
	return fmt.Sprintf(`resource "genesyscloud_knowledge_knowledgebase" "%s" {
		name = "%s"
        description = "%s"
        core_language = "%s"
	}
	`, resourceLabel, name, description, coreLanguage)
}
