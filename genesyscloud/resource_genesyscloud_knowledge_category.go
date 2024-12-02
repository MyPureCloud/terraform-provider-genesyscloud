package genesyscloud

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
		resources[id] = &resourceExporter.ResourceMeta{BlockLabel: *knowledgeCategory.Name}
	}

	return resources, nil
}

func getAllKnowledgeCategoryEntities(knowledgeAPI platformclientv2.KnowledgeApi, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Categoryresponse, diag.Diagnostics) {
	var (
		after    string
		err      error
		entities []platformclientv2.Categoryresponse
	)

	const pageSize = 100
	for i := 0; ; i++ {
		knowledgeCategories, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategories(*knowledgeBase.Id, "", after, fmt.Sprintf("%v", pageSize), "", false, "", "", "", false)
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to read knowledge document error: %s", getErr), resp)
		}

		if knowledgeCategories.Entities == nil || len(*knowledgeCategories.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeCategories.Entities...)

		if knowledgeCategories.NextUri == nil || *knowledgeCategories.NextUri == "" {
			break
		}

		after, err = util.GetQueryParamValueFromUri(*knowledgeCategories.NextUri, "after")
		if err != nil {
			return nil, util.BuildDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to parse after cursor from knowledge category nextUri"), err)
		}
		if after == "" {
			break
		}
	}

	return &entities, nil
}

func KnowledgeCategoryExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllKnowledgeCategories),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"knowledge_base_id":            {RefType: "genesyscloud_knowledge_knowledgebase"},
			"knowledge_category.parent_id": {RefType: "genesyscloud_knowledge_category"},
		},
	}
}

func ResourceKnowledgeCategory() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Category",

		CreateContext: provider.CreateWithPooledClient(createKnowledgeCategory),
		ReadContext:   provider.ReadWithPooledClient(readKnowledgeCategory),
		UpdateContext: provider.UpdateWithPooledClient(updateKnowledgeCategory),
		DeleteContext: provider.DeleteWithPooledClient(deleteKnowledgeCategory),
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	knowledgeCategoryRequest := buildKnowledgeCategoryCreate(knowledgeCategory)

	log.Printf("Creating knowledge category %s", knowledgeCategory["name"].(string))
	knowledgeCategoryResponse, resp, err := knowledgeAPI.PostKnowledgeKnowledgebaseCategories(knowledgeBaseId, *knowledgeCategoryRequest)
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
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeCategory(), constants.ConsistencyChecks(), "genesyscloud_knowledge_category")

	log.Printf("Reading knowledge category %s", knowledgeCategoryId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeCategory, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId)
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
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating knowledge category %s", knowledgeCategory["name"].(string))
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge category version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to read knowledge category %s error: %s", knowledgeCategoryId, getErr), resp)
		}

		knowledgeCategoryUpdate := buildKnowledgeCategoryUpdate(knowledgeCategory)

		log.Printf("Updating knowledge category %s", knowledgeCategory["name"].(string))
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId, *knowledgeCategoryUpdate)
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
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting knowledge category %s", id)
	_, resp, err := knowledgeAPI.DeleteKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_category", fmt.Sprintf("Failed to delete knowledge category %s error: %s", id, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseCategory(knowledgeBaseId, knowledgeCategoryId)
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

func buildKnowledgeCategoryUpdate(categoryIn map[string]interface{}) *platformclientv2.Categoryupdaterequest {
	name := categoryIn["name"].(string)

	categoryOut := platformclientv2.Categoryupdaterequest{
		Name: &name,
	}

	if description, ok := categoryIn["description"].(string); ok && description != "" {
		categoryOut.Description = &description
	}

	if parentId, ok := categoryIn["parent_id"].(string); ok && parentId != "" {
		if strings.Contains(parentId, ",") {
			ids := strings.Split(parentId, ",")
			parent_Id := ids[0]
			categoryOut.ParentCategoryId = &parent_Id
		} else {
			categoryOut.ParentCategoryId = &parentId
		}
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
		if strings.Contains(parentId, ",") {
			ids := strings.Split(parentId, ",")
			parent_Id := ids[0]
			categoryOut.ParentCategoryId = &parent_Id
		} else {
			categoryOut.ParentCategoryId = &parentId
		}
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
		categoryOut["parent_id"] = *(*categoryIn.ParentCategory).Id + "," + *(*categoryIn.KnowledgeBase).Id
	}

	return []interface{}{categoryOut}
}
