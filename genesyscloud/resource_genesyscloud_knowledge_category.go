package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

var (
	knowledgeCategory = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Knowledge base name. Changing the name attribute will cause the knowledge_category resource to be dropped and recreated with a new ID.",
				Type:        schema.TypeString,
				Required:    true,
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

func getAllKnowledgeCategories(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	categoryEntities := make([]platformclientv2.Categoryresponse, 0)
	resources := make(resourceExporter.ResourceIDMetaMap)
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(clientConfig)

	// get published knowledge bases
	publishedEntities, err := getAllKnowledgebaseEntities(*knowledgeAPI, true)
	if err != nil {
		return nil, err
	}
	knowledgeBaseList = append(knowledgeBaseList, *publishedEntities...)

	// get unpublished knowledge bases
	unpublishedEntities, err := getAllKnowledgebaseEntities(*knowledgeAPI, false)
	if err != nil {
		return nil, err
	}
	knowledgeBaseList = append(knowledgeBaseList, *unpublishedEntities...)

	for _, knowledgeBase := range knowledgeBaseList {
		partialEntities, err := getAllKnowledgeCategoryEntities(*knowledgeAPI, &knowledgeBase)
		if err != nil {
			return nil, err
		}
		categoryEntities = append(categoryEntities, *partialEntities...)
	}

	for _, knowledgeCategory := range categoryEntities {
		id := fmt.Sprintf("%s,%s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id)
		resources[id] = &resourceExporter.ResourceMeta{Name: *knowledgeCategory.Name}
	}

	return resources, nil
}

func getAllKnowledgeCategoryEntities(knowledgeAPI platformclientv2.KnowledgeApi, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Categoryresponse, diag.Diagnostics) {
	var (
		after    string
		entities []platformclientv2.Categoryresponse
	)

	const pageSize = 100
	for i := 0; ; i++ {
		knowledgeCategories, _, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategories(*knowledgeBase.Id, "", after, fmt.Sprintf("%v", pageSize), "", false, "", "", "", false)
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of knowledge categories: %v", getErr)
		}

		if knowledgeCategories.Entities == nil || len(*knowledgeCategories.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeCategories.Entities...)

		if knowledgeCategories.NextUri == nil || *knowledgeCategories.NextUri == "" {
			break
		}

		u, err := url.Parse(*knowledgeCategories.NextUri)
		if err != nil {
			return nil, diag.Errorf("Failed to parse after cursor from knowledge category nextUri: %v", err)
		}
		m, _ := url.ParseQuery(u.RawQuery)
		if afterSlice, ok := m["after"]; ok && len(afterSlice) > 0 {
			after = afterSlice[0]
			if after == "" {
				break
			}
		}
	}

	return &entities, nil
}

func KnowledgeCategoryExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllKnowledgeCategories),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"knowledge_base_id": {RefType: "genesyscloud_knowledge_knowledgebase"},
		},
	}
}

func ResourceKnowledgeCategory() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Category",

		CreateContext: CreateWithPooledClient(createKnowledgeCategory),
		ReadContext:   ReadWithPooledClient(readKnowledgeCategory),
		UpdateContext: UpdateWithPooledClient(updateKnowledgeCategory),
		DeleteContext: DeleteWithPooledClient(deleteKnowledgeCategory),
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
			"knowledge_category": {
				Description: "Knowledge category id",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        knowledgeCategory,
			},
		},
	}
}

func createKnowledgeCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	knowledgeCategory := d.Get("knowledge_category").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	knowledgeCategoryRequest := buildKnowledgeCategoryCreate(knowledgeCategory)

	log.Printf("Creating knowledge category %s", knowledgeCategory["name"].(string))
	knowledgeCategoryResponse, _, err := knowledgeAPI.PostKnowledgeKnowledgebaseCategories(knowledgeBaseId, *knowledgeCategoryRequest)
	if err != nil {
		return diag.Errorf("Failed to create knowledge category %s: %s", knowledgeBaseId, err)
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Reading knowledge category %s", knowledgeCategoryId)
	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeCategory, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId)
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read knowledge category %s: %s", knowledgeCategoryId, getErr))
			}
			log.Printf("%s", getErr)
			return retry.NonRetryableError(fmt.Errorf("Failed to read knowledge category %s: %s", knowledgeCategoryId, getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeCategory())

		newId := fmt.Sprintf("%s,%s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id)
		d.SetId(newId)
		d.Set("knowledge_base_id", *knowledgeCategory.KnowledgeBase.Id)
		d.Set("knowledge_category", flattenKnowledgeCategory(*knowledgeCategory))
		log.Printf("Read knowledge category %s", knowledgeCategoryId)
		return cc.CheckState()
	})
}

func updateKnowledgeCategory(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), ",")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]
	knowledgeCategory := d.Get("knowledge_category").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating knowledge category %s", knowledgeCategory["name"].(string))
	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge category version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId)
		if getErr != nil {
			return resp, diag.Errorf("Failed to read knowledge category %s: %s", knowledgeCategoryId, getErr)
		}

		knowledgeCategoryUpdate := buildKnowledgeCategoryUpdate(knowledgeCategory)

		log.Printf("Updating knowledge category %s", knowledgeCategory["name"].(string))
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId, *knowledgeCategoryUpdate)
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
	id := strings.Split(d.Id(), ",")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting knowledge category %s", id)
	_, _, err := knowledgeAPI.DeleteKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId)
	if err != nil {
		return diag.Errorf("Failed to delete knowledge category %s: %s", id, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId)
		if err != nil {
			if IsStatus404(resp) {
				// Knowledge category deleted
				log.Printf("Deleted knowledge category %s", knowledgeCategoryId)
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting knowledge category %s: %s", knowledgeCategoryId, err))
		}

		return retry.RetryableError(fmt.Errorf("Knowledge category %s still exists", knowledgeCategoryId))
	})
}

func buildKnowledgeCategoryUpdate(categoryIn map[string]interface{}) *platformclientv2.Categoryupdaterequest {
	name := categoryIn["name"].(string)

	categoryOut := platformclientv2.Categoryupdaterequest{
		Name: &name,
	}

	if description, ok := categoryIn["description"].(string); ok && description != "" {
		categoryOut.Description = &description
	}
	if parentId, ok := categoryIn["parent_id"].(string); ok && parentId != "" {
		categoryOut.ParentCategoryId = &parentId
	}

	return &categoryOut
}

func buildKnowledgeCategoryCreate(categoryIn map[string]interface{}) *platformclientv2.Categorycreaterequest {
	name := categoryIn["name"].(string)

	categoryOut := platformclientv2.Categorycreaterequest{
		Name: &name,
	}

	if description, ok := categoryIn["description"].(string); ok && description != "" {
		categoryOut.Description = &description
	}
	if parentId, ok := categoryIn["parent_id"].(string); ok && parentId != "" {
		categoryOut.ParentCategoryId = &parentId
	}

	return &categoryOut
}

func flattenKnowledgeCategory(categoryIn platformclientv2.Categoryresponse) []interface{} {
	categoryOut := make(map[string]interface{})

	if categoryIn.Name != nil {
		categoryOut["name"] = *categoryIn.Name
	}
	if categoryIn.Description != nil {
		categoryOut["description"] = *categoryIn.Description
	}
	if categoryIn.ParentCategory != nil && (*categoryIn.ParentCategory).Id != nil {
		categoryOut["parent_id"] = (*categoryIn.ParentCategory).Id
	}

	return []interface{}{categoryOut}
}
