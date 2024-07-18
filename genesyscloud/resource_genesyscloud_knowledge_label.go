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
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var (
	knowledgeLabel = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the label.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"color": {
				Description: "The color for the label.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
)

func getAllKnowledgeLabels(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	knowledgeBaseList := make([]platformclientv2.Knowledgebase, 0)
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
		labelEntities, err := getAllKnowledgeLabelEntities(*knowledgeAPI, &knowledgeBase)
		if err != nil {
			return nil, err
		}

		for _, knowledgeLabel := range *labelEntities {
			id := fmt.Sprintf("%s,%s", *knowledgeLabel.Id, *knowledgeBase.Id)
			resources[id] = &resourceExporter.ResourceMeta{Name: *knowledgeLabel.Name}
		}
	}

	return resources, nil
}

func getAllKnowledgeLabelEntities(knowledgeAPI platformclientv2.KnowledgeApi, knowledgeBase *platformclientv2.Knowledgebase) (*[]platformclientv2.Labelresponse, diag.Diagnostics) {
	var (
		after    string
		entities []platformclientv2.Labelresponse
	)

	const pageSize = 100
	for i := 0; ; i++ {
		knowledgeLabels, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabels(*knowledgeBase.Id, "", after, fmt.Sprintf("%v", pageSize), "", false)
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to get knowledge labels error: %s", getErr), resp)
		}

		if knowledgeLabels.Entities == nil || len(*knowledgeLabels.Entities) == 0 {
			break
		}

		entities = append(entities, *knowledgeLabels.Entities...)

		if knowledgeLabels.NextUri == nil || *knowledgeLabels.NextUri == "" {
			break
		}

		after, err := util.GetQueryParamValueFromUri(*knowledgeLabels.NextUri, "after")
		if err != nil {
			return nil, util.BuildDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to parse after cursor from knowledge label nextUri"), err)
		}
		if after == "" {
			break
		}
	}

	return &entities, nil
}

func KnowledgeLabelExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllKnowledgeLabels),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"knowledge_base_id": {RefType: "genesyscloud_knowledge_knowledgebase"},
		},
	}
}

func ResourceKnowledgeLabel() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Knowledge Label",

		CreateContext: provider.CreateWithPooledClient(createKnowledgeLabel),
		ReadContext:   provider.ReadWithPooledClient(readKnowledgeLabel),
		UpdateContext: provider.UpdateWithPooledClient(updateKnowledgeLabel),
		DeleteContext: provider.DeleteWithPooledClient(deleteKnowledgeLabel),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"knowledge_base_id": {
				Description: "Knowledge base id of the label",
				Type:        schema.TypeString,
				Required:    true,
			},
			"knowledge_label": {
				Description: "Knowledge label id",
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Elem:        knowledgeLabel,
			},
		},
	}
}

func createKnowledgeLabel(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	knowledgeBaseId := d.Get("knowledge_base_id").(string)
	knowledgeLabel := d.Get("knowledge_label").([]interface{})[0].(map[string]interface{})

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	knowledgeLabelRequest := buildKnowledgeLabel(knowledgeLabel)

	log.Printf("Creating knowledge label %s", knowledgeLabel["name"].(string))
	knowledgeLabelResponse, resp, err := knowledgeAPI.PostKnowledgeKnowledgebaseLabels(knowledgeBaseId, knowledgeLabelRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to create knowledge label %s error: %s", knowledgeBaseId, err), resp)
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
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceKnowledgeLabel(), constants.DefaultConsistencyChecks, "genesyscloud_knowledge_label")

	log.Printf("Reading knowledge label %s", knowledgeLabelId)
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		knowledgeLabel, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabel(knowledgeBaseId, knowledgeLabelId)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to read knowledge label %s | error: %s", knowledgeLabelId, getErr), resp))
			}
			log.Printf("%s", getErr)
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to read knowledge label %s | error: %s", knowledgeLabelId, getErr), resp))
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
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Updating knowledge label %s", knowledgeLabel["name"].(string))
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current knowledge label version
		_, resp, getErr := knowledgeAPI.GetKnowledgeKnowledgebaseLabel(knowledgeBaseId, knowledgeLabelId)
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to read knowledge label %s error: %s", knowledgeLabelId, getErr), resp)
		}

		knowledgeLabelUpdate := buildKnowledgeLabelUpdate(knowledgeLabel)

		log.Printf("Updating knowledge label %s", knowledgeLabel["name"].(string))
		_, resp, putErr := knowledgeAPI.PatchKnowledgeKnowledgebaseLabel(knowledgeBaseId, knowledgeLabelId, knowledgeLabelUpdate)
		if putErr != nil {
			return resp, util.BuildAPIDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to update knowledge label %s error: %s", knowledgeLabelId, putErr), resp)
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
	knowledgeAPI := platformclientv2.NewKnowledgeApiWithConfig(sdkConfig)

	log.Printf("Deleting knowledge label %s", id)
	_, resp, err := knowledgeAPI.DeleteKnowledgeKnowledgebaseLabel(knowledgeBaseId, knowledgeLabelId)
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Failed to delete knowledge label %s error: %s", id, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := knowledgeAPI.GetKnowledgeKnowledgebaseLabel(knowledgeBaseId, knowledgeLabelId)
		if err != nil {
			if util.IsStatus404(resp) {
				// Knowledge label deleted
				log.Printf("Deleted knowledge label %s", knowledgeLabelId)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Error deleting knowledge label %s | error: %s", knowledgeLabelId, err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_knowledge_label", fmt.Sprintf("Knowledge label %s still exists", knowledgeLabelId), resp))
	})
}

func buildKnowledgeLabel(labelIn map[string]interface{}) platformclientv2.Labelcreaterequest {
	name := labelIn["name"].(string)
	color := labelIn["color"].(string)

	labelOut := platformclientv2.Labelcreaterequest{
		Name:  &name,
		Color: &color,
	}

	return labelOut
}

func buildKnowledgeLabelUpdate(labelIn map[string]interface{}) platformclientv2.Labelupdaterequest {
	name := labelIn["name"].(string)
	color := labelIn["color"].(string)

	labelOut := platformclientv2.Labelupdaterequest{
		Name:  &name,
		Color: &color,
	}

	return labelOut
}

func flattenKnowledgeLabel(labelIn *platformclientv2.Labelresponse) []interface{} {
	labelOut := make(map[string]interface{})

	if labelIn.Name != nil {
		labelOut["name"] = *labelIn.Name
	}
	if labelIn.Color != nil {
		labelOut["color"] = *labelIn.Color
	}

	return []interface{}{labelOut}
}
