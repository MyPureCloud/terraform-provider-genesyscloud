package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func getAllKnowledgeKnowledgebases(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)

	publishedEntities, err := getAllKnowledgebaseEntities(*knowledgeAPI, true)
	if err != nil {
		return nil, err
	}

	for _, knowledgeBase := range *publishedEntities {
		resources[*knowledgeBase.Id] = &resourceExporter.ResourceMeta{Name: *knowledgeBase.Name}
	}

	unpublishedEntities, err := getAllKnowledgebaseEntities(*knowledgeAPI, false)
	if err != nil {
		return nil, err
	}

	for _, knowledgeBase := range *unpublishedEntities {
		resources[*knowledgeBase.Id] = &resourceExporter.ResourceMeta{Name: *knowledgeBase.Name}
	}

	return resources, nil
}

func getAllKnowledgebaseEntities(knowledgeApi platformclientv2.KnowledgeApi, published bool) (*[]platformclientv2.Knowledgebase, diag.Diagnostics) {
	var (
		after    string
		entities []platformclientv2.Knowledgebase
	)

	const pageSize = 100
	for i := 0; ; i++ {
		knowledgeBases, _, getErr := knowledgeApi.GetKnowledgeKnowledgebases("", after, "", fmt.Sprintf("%v", pageSize), "", "", published, "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of knowledge bases: %v", getErr)
		}

		if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeBases.Entities...)

		if knowledgeBases.NextUri == nil || *knowledgeBases.NextUri == "" {
			break
		}

		u, err := url.Parse(*knowledgeBases.NextUri)
		if err != nil {
			return nil, diag.Errorf("Failed to parse after cursor from knowledge base nextUri: %v", err)
		}
		m, _ := url.ParseQuery(u.RawQuery)
		if afterSlice, ok := m["after"]; ok && len(afterSlice) > 0 {
			after = afterSlice[0]
		}
		if after == "" {
			break
		}
	}

	return &entities, nil
}

func KnowledgeKnowledgebaseExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllKnowledgeKnowledgebases),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
	}
}

func ResourceKnowledgeKnowledgebase() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Base",

		CreateContext: CreateWithPooledClient(createKnowledgeKnowledgebase),
		ReadContext:   ReadWithPooledClient(readKnowledgeKnowledgebase),
		UpdateContext: UpdateWithPooledClient(updateKnowledgeKnowledgebase),
		DeleteContext: DeleteWithPooledClient(deleteKnowledgeKnowledgebase),
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
				Description:  "Core language for knowledge base in which initial content must be created, language codes [en-US, en-UK, en-AU, de-DE] are supported currently, however the new DX knowledge will support all these language codes",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"en-US", "en-UK", "en-AU", "de-DE", "es-US", "es-ES", "fr-FR", "pt-BR", "nl-NL", "it-IT", "fr-CA"}, false),
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Creating knowledge base %s", name)
	knowledgeBase, _, err := knowledgeAPI.PostKnowledgeKnowledgebases(platformclientv2.Knowledgebasecreaterequest{
		Name:         &name,
		Description:  &description,
		CoreLanguage: &coreLanguage,
	})

	if err != nil {
		return diag.Errorf("Failed to create knowledge base %s: %s", name, err)
	}

	d.SetId(*knowledgeBase.Id)

	log.Printf("Created knowledge base %s %s", name, *knowledgeBase.Id)
	return readKnowledgeKnowledgebase(ctx, d, meta)
}

func readKnowledgeKnowledgebase(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Reading knowledge base %s", d.Id())
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeBase, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebase(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read knowledge base %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read knowledge base %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeKnowledgebase())

		resourcedata.SetNillableValue(d, "name", knowledgeBase.Name)
		resourcedata.SetNillableValue(d, "description", knowledgeBase.Description)
		resourcedata.SetNillableValue(d, "core_language", knowledgeBase.CoreLanguage)
		resourcedata.SetNillableValue(d, "published", knowledgeBase.Published)
		log.Printf("Read knowledge base %s %s", d.Id(), *knowledgeBase.Name)
		return cc.CheckState()
	})
}

func updateKnowledgeKnowledgebase(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating knowledge base %s", name)
	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge base version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebase(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read knowledge base %s: %s", d.Id(), getErr)
		}

		update := platformclientv2.Knowledgebaseupdaterequest{
			Name:        &name,
			Description: &description,
		}

		log.Printf("Updating knowledge base %s", name)
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebase(d.Id(), update)
		if putErr != nil {
			return resp, diag.Errorf("Failed to update knowledge base %s: %s", d.Id(), putErr)
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting knowledge base %s", name)
	_, _, err := knowledgeAPI.DeleteKnowledgeKnowledgebase(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete knowledge base %s: %s", name, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebase(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// Knowledge base deleted
				log.Printf("Deleted Knowledge base %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting knowledge base %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Knowledge base %s still exists", d.Id()))
	})
}
