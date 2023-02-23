package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v92/platformclientv2"
)

var (
	knowledgeCategory = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Knowledge base name. Changing the name attribute will cause the knowledge_category resource to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Knowledge base description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"parent_id": {
				Description: "Knowledge category parent id",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
)

func getAllKnowledgeCategories(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
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
			knowledgeBaseList = append(knowledgeBaseList, knowledgeBase)
		}
	}
	for _, knowledgeBase := range knowledgeBaseList {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			knowledgeCategories, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategories(*knowledgeBase.Id, *knowledgeBase.CoreLanguage, "", "", "", fmt.Sprintf("%v", pageSize), "")
			if getErr != nil {
				return nil, diag.Errorf("Failed to get page of knowledge categories: %v", getErr)
			}

			if knowledgeCategories.Entities == nil || len(*knowledgeCategories.Entities) == 0 {
				break
			}

			for _, knowledgeCategory := range *knowledgeCategories.Entities {
				id := fmt.Sprintf("%s %s %s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id, *knowledgeCategory.LanguageCode)
				resources[id] = &ResourceMeta{Name: *knowledgeCategory.Name}
			}
		}
	}

	return resources, nil
}

func knowledgeCategoryExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllKnowledgeCategories),
		RefAttrs: map[string]*RefAttrSettings{
			"knowledge_base_id": {RefType: "genesyscloud_knowledge_knowledgebase"},
		},
	}
}

func resourceKnowledgeCategory() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Category",

		CreateContext: createWithPooledClient(createKnowledgeCategory),
		ReadContext:   readWithPooledClient(readKnowledgeCategory),
		UpdateContext: updateWithPooledClient(updateKnowledgeCategory),
		DeleteContext: deleteWithPooledClient(deleteKnowledgeCategory),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"knowledge_base_id": {
				Description: "Knowledge base id of the category",
				Type:        schema.TypeString,
				Required:    true,
			},
			"language_code": {
				Description:  "language code of the category",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"en-US", "en-UK", "en-AU", "de-DE", "es-US", "es-ES", "fr-FR", "pt-BR", "nl-NL", "it-IT", "fr-CA"}, false),
			},
			"knowledge_category": {
				Description: "Knowledge category parent id",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        knowledgeCategory,
			},
		},
	}
}

func createKnowledgeCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	languageCode := d.Get("language_code").(string)
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	knowledgeCategory := d.Get("knowledge_category").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	knowledgeCategoryRequest := buildKnowledgeCategory(knowledgeCategory)

	log.Printf("Creating knowledge category %s", knowledgeCategory["name"].(string))
	knowledgeCategoryResponse, _, err := knowledgeAPI.PostKnowledgeKnowledgebaseLanguageCategories(knowledgeBaseId, languageCode, knowledgeCategoryRequest)
	if err != nil {
		return diag.Errorf("Failed to create knowledge category %s: %s", knowledgeBaseId, err)
	}

	id := fmt.Sprintf("%s %s %s", *knowledgeCategoryResponse.Id, *knowledgeCategoryResponse.KnowledgeBase.Id, *knowledgeCategoryResponse.LanguageCode)
	d.SetId(id)

	log.Printf("Created knowledge category %s", *knowledgeCategoryResponse.Id)
	return readKnowledgeCategory(ctx, d, meta)
}

func readKnowledgeCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Reading knowledge category %s", knowledgeCategoryId)
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		knowledgeCategory, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, languageCode)
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read knowledge category %s: %s", knowledgeCategoryId, getErr))
			}
			log.Printf("%s", getErr)
			return resource.NonRetryableError(fmt.Errorf("Failed to read knowledge category %s: %s", knowledgeCategoryId, getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceKnowledgeCategory())

		newId := fmt.Sprintf("%s %s %s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id, *knowledgeCategory.LanguageCode)
		d.SetId(newId)
		d.Set("knowledge_base_id", *knowledgeCategory.KnowledgeBase.Id)
		d.Set("language_code", *knowledgeCategory.LanguageCode)
		d.Set("knowledge_category", flattenKnowledgeCategory(knowledgeCategory))
		log.Printf("Read knowledge category %s", knowledgeCategoryId)
		return cc.CheckState()
	})
}

func updateKnowledgeCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]
	knowledgeCategory := d.Get("knowledge_category").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating knowledge category %s", knowledgeCategory["name"].(string))
	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge category version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, languageCode)
		if getErr != nil {
			return resp, diag.Errorf("Failed to read knowledge category %s: %s", knowledgeCategoryId, getErr)
		}

		knowledgeCategoryUpdate := buildKnowledgeCategory(knowledgeCategory)

		log.Printf("Updating knowledge category %s", knowledgeCategory["name"].(string))
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, languageCode, knowledgeCategoryUpdate)
		if putErr != nil {
			return resp, diag.Errorf("Failed to update knowledge category %s: %s", knowledgeCategoryId, putErr)
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
	id := strings.Split(d.Id(), " ")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting knowledge category %s", id)
	_, _, err := knowledgeAPI.DeleteKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, languageCode)
	if err != nil {
		return diag.Errorf("Failed to delete knowledge category %s: %s", id, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, languageCode)
		if err != nil {
			if isStatus404(resp) {
				// Knowledge base deleted
				log.Printf("Deleted knowledge category %s", knowledgeCategoryId)
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting knowledge category %s: %s", knowledgeCategoryId, err))
		}

		return resource.RetryableError(fmt.Errorf("Knowledge category %s still exists", knowledgeCategoryId))
	})
}

func buildKnowledgeCategory(categoryIn map[string]interface{}) platformclientv2.Knowledgecategoryrequest {
	categoryOut := platformclientv2.Knowledgecategoryrequest{}

	if name, ok := categoryIn["name"].(string); ok && name != "" {
		categoryOut.Name = &name
	}
	if description, ok := categoryIn["description"].(string); ok && description != "" {
		categoryOut.Description = &description
	}
	if parentId, ok := categoryIn["parent_id"].(string); ok && parentId != "" {
		categoryOut.Parent = &platformclientv2.Documentcategoryinput{
			Id: &parentId,
		}
	}

	return categoryOut
}

func flattenKnowledgeCategory(categoryIn *platformclientv2.Knowledgeextendedcategory) []interface{} {
	categoryOut := make(map[string]interface{})

	if categoryIn.Name != nil {
		categoryOut["name"] = *categoryIn.Name
	}
	if categoryIn.Description != nil {
		categoryOut["description"] = *categoryIn.Description
	}
	if categoryIn.Parent != nil && categoryIn.Parent.Id != nil {
		categoryOut["parent_id"] = *categoryIn.Parent.Id
	}

	return []interface{}{categoryOut}
}
