package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/validators"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var (
	knowledgeCategoryV1 = &schema.Resource{
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

func getAllKnowledgeCategoriesV1(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
	categoryEntities := make([]platformclientv2.Knowledgecategory, 0)
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
		partialEntities, err := getAllKnowledgeV1CategoryEntities(*knowledgeAPI, &knowledgeBase)
		if err != nil {
			return nil, err
		}
		categoryEntities = append(categoryEntities, *partialEntities...)
	}

	for _, knowledgeCategory := range categoryEntities {
		id := fmt.Sprintf("%s %s %s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id, *knowledgeCategory.LanguageCode)
		resources[id] = &resourceExporter.ResourceMeta{Name: *knowledgeCategory.Name}
	}

	return resources, nil
}

func getAllKnowledgeV1CategoryEntities(knowledgeAPI platformclientv2.KnowledgeApi, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Knowledgecategory, diag.Diagnostics) {
	var (
		after    string
		entities []platformclientv2.Knowledgecategory
	)

	const pageSize = 100
	for i := 0; ; i++ {
		knowledgeCategories, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategories(*knowledgeBase.Id, *knowledgeBase.CoreLanguage, "", after, "", fmt.Sprintf("%v", pageSize), "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_category", fmt.Sprintf("Failed to get knowledge categorys error: %s", getErr), resp)
		}

		if knowledgeCategories.Entities == nil || len(*knowledgeCategories.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeCategories.Entities...)

		if knowledgeCategories.NextUri == nil || *knowledgeCategories.NextUri == "" {
			break
		}

		after, err := util.GetQueryParamValueFromUri(*knowledgeCategories.NextUri, "after")
		if err != nil {
			return nil, util.BuildDiagnosticError("genesyscloud_knowledge_v1_category", fmt.Sprintf("Failed to parse after cursor from knowledge category nextUri"), err)
		}
		if after == "" {
			break
		}
	}

	return &entities, nil
}

func KnowledgeCategoryExporterV1() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllKnowledgeCategoriesV1),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"knowledge_base_id": {RefType: "genesyscloud_knowledge_knowledgebase"},
		},
	}
}

func ResourceKnowledgeCategoryV1() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge v1 Category",

		CreateContext: provider.CreateWithPooledClient(createKnowledgeCategoryV1),
		ReadContext:   provider.ReadWithPooledClient(readKnowledgeCategoryV1),
		UpdateContext: provider.UpdateWithPooledClient(updateKnowledgeCategoryV1),
		DeleteContext: provider.DeleteWithPooledClient(deleteKnowledgeCategoryV1),
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
				Description:      "language code of the category",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateLanguageCode,
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

func createKnowledgeCategoryV1(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	languageCode := d.Get("language_code").(string)
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	knowledgeCategory := d.Get("knowledge_category").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	knowledgeCategoryRequest := buildKnowledgeCategoryV1(knowledgeCategory)

	log.Printf("Creating knowledge category %s", knowledgeCategory["name"].(string))
	knowledgeCategoryResponse, resp, err := knowledgeAPI.PostKnowledgeKnowledgebaseLanguageCategories(knowledgeBaseId, languageCode, knowledgeCategoryRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_category", fmt.Sprintf("Failed to create knowledge category %s error: %s", knowledgeBaseId, err), resp)
	}

	id := fmt.Sprintf("%s %s %s", *knowledgeCategoryResponse.Id, *knowledgeCategoryResponse.KnowledgeBase.Id, *knowledgeCategoryResponse.LanguageCode)
	d.SetId(id)

	log.Printf("Created knowledge category %s", *knowledgeCategoryResponse.Id)
	return readKnowledgeCategoryV1(ctx, d, meta)
}

func readKnowledgeCategoryV1(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeCategory(), constants.DefaultConsistencyChecks, "genesyscloud_knowledge_v1_category")

	log.Printf("Reading knowledge category %s", knowledgeCategoryId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeCategory, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, languageCode)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_v1_category", fmt.Sprintf("Failed to read knowledge category %s | error: %s", knowledgeCategoryId, getErr), resp))
			}
			log.Printf("%s", getErr)
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_v1_category", fmt.Sprintf("Failed to read knowledge category %s | error: %s", knowledgeCategoryId, getErr), resp))
		}

		newId := fmt.Sprintf("%s %s %s", *knowledgeCategory.Id, *knowledgeCategory.KnowledgeBase.Id, *knowledgeCategory.LanguageCode)
		d.SetId(newId)
		d.Set("knowledge_base_id", *knowledgeCategory.KnowledgeBase.Id)
		d.Set("language_code", *knowledgeCategory.LanguageCode)
		d.Set("knowledge_category", flattenKnowledgeCategoryV1(knowledgeCategory))
		log.Printf("Read knowledge category %s", knowledgeCategoryId)
		return cc.CheckState(d)
	})
}

func updateKnowledgeCategoryV1(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]
	knowledgeCategory := d.Get("knowledge_category").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating knowledge category %s", knowledgeCategory["name"].(string))
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge category version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, languageCode)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_category", fmt.Sprintf("Failed to read knowledge category %s error: %s", knowledgeCategoryId, getErr), resp)
		}

		knowledgeCategoryUpdate := buildKnowledgeCategoryV1(knowledgeCategory)

		log.Printf("Updating knowledge category %s", knowledgeCategory["name"].(string))
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, languageCode, knowledgeCategoryUpdate)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_category", fmt.Sprintf("Failed to update knowledge category %s error: %s", knowledgeCategoryId, putErr), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated knowledge category %s %s", knowledgeCategory["name"].(string), knowledgeCategoryId)
	return readKnowledgeCategoryV1(ctx, d, meta)
}

func deleteKnowledgeCategoryV1(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := strings.Split(d.Id(), " ")
	knowledgeCategoryId := id[0]
	knowledgeBaseId := id[1]
	languageCode := id[2]

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting knowledge category %s", id)
	_, resp, err := knowledgeAPI.DeleteKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, languageCode)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_v1_category", fmt.Sprintf("Failed to delete knowledge category %s error: %s", id, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLanguageCategory(knowledgeCategoryId, knowledgeBaseId, languageCode)
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge category deleted
				log.Printf("Deleted knowledge category %s", knowledgeCategoryId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_v1_category", fmt.Sprintf("Error deleting knowledge category %s | error: %s", knowledgeCategoryId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_v1_category", fmt.Sprintf("Knowledge category %s still exists", knowledgeCategoryId), resp))
	})
}

func buildKnowledgeCategoryV1(categoryIn map[string]interface{}) platformclientv2.Knowledgecategoryrequest {
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

func flattenKnowledgeCategoryV1(categoryIn *platformclientv2.Knowledgeextendedcategory) []interface{} {
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
