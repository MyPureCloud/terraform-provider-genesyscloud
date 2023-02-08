package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

func getAllKnowledgeKnowledgebases(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		knowledgeBases, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebases("", "", "", fmt.Sprintf("%v", pageSize), "", "", false, "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of knowledge bases: %v", getErr)
		}

		if knowledgeBases.Entities == nil || len(*knowledgeBases.Entities) == 0 {
			break
		}

		for _, knowledgeBase := range *knowledgeBases.Entities {
			resources[*knowledgeBase.Id] = &ResourceMeta{Name: *knowledgeBase.Name}
		}
	}

	return resources, nil
}

func knowledgeKnowledgebaseExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllKnowledgeKnowledgebases),
		RefAttrs:         map[string]*RefAttrSettings{}, // No references
	}
}

func resourceKnowledgeKnowledgebase() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Base",

		CreateContext: createWithPooledClient(createKnowledgeKnowledgebase),
		ReadContext:   readWithPooledClient(readKnowledgeKnowledgebase),
		UpdateContext: updateWithPooledClient(updateKnowledgeKnowledgebase),
		DeleteContext: deleteWithPooledClient(deleteKnowledgeKnowledgebase),
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Creating knowledge base %s", name)
	knowledgeBase, _, err := knowledgeAPI.PostKnowledgeKnowledgebases(platformclientv2.Knowledgebase{
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
	sdkConfig := meta.(*providerMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Reading knowledge base %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		knowledgeBase, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebase(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read knowledge base %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read knowledge base %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceKnowledgeKnowledgebase())

		d.Set("name", *knowledgeBase.Name)
		d.Set("description", *knowledgeBase.Description)
		d.Set("core_language", *knowledgeBase.CoreLanguage)
		d.Set("published", *knowledgeBase.Published)
		log.Printf("Read knowledge base %s %s", d.Id(), *knowledgeBase.Name)
		return cc.CheckState()
	})
}

func updateKnowledgeKnowledgebase(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	coreLanguage := d.Get("core_language").(string)

	sdkConfig := meta.(*providerMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating knowledge base %s", name)
	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge base version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebase(d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read knowledge base %s: %s", d.Id(), getErr)
		}

		update := platformclientv2.Knowledgebase{
			Name:         &name,
			Description:  &description,
			CoreLanguage: &coreLanguage,
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

	sdkConfig := meta.(*providerMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting knowledge base %s", name)
	_, _, err := knowledgeAPI.DeleteKnowledgeKnowledgebase(d.Id())
	if err != nil {
		return diag.Errorf("Failed to delete knowledge base %s: %s", name, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebase(d.Id())
		if err != nil {
			if isStatus404(resp) {
				// Knowledge base deleted
				log.Printf("Deleted Knowledge base %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting knowledge base %s: %s", d.Id(), err))
		}

		return resource.RetryableError(fmt.Errorf("Knowledge base %s still exists", d.Id()))
	})
}
