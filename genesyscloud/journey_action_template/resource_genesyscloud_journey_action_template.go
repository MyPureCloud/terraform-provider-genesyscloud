package journey_action_template

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

func getAllJourneyActionTemplates(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	journeyApi := platformclientv2.NewJourneyApiWithConfig(clientConfig)
	pageCount := 1 // Needed because of broken journey common paging
	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		const pageSize = 100
		actionTemplates, resp, getErr := journeyApi.GetJourneyActiontemplates(pageNum, pageSize, "", "", "", nil, "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of journey action maps error: %s", getErr), resp)
		}
		if actionTemplates.Entities == nil || len(*actionTemplates.Entities) == 0 {
			break
		}
		for _, actionTemplate := range *actionTemplates.Entities {
			resources[*actionTemplate.Id] = &resourceExporter.ResourceMeta{BlockLabel: *actionTemplate.Name}
		}
		pageCount = *actionTemplates.PageCount
	}
	return resources, nil
}

func createJourneyActionTemplate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	journeyApi := journeyApiConfig(i)
	actionTemplate := buildSdkActionTemplate(data)
	log.Printf("Creating Journey Action Template %s", *actionTemplate.Name)
	result, resp, err := journeyApi.PostJourneyActiontemplates(*actionTemplate)
	if err != nil {
		input, _ := util.InterfaceToJson(*actionTemplate)
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create journey action template %s (input: %+v) error: %s", *actionTemplate.Name, input, err), resp)
	}
	data.SetId(*result.Id)
	log.Printf("Created Journey Action Template %s %s", *result.Name, *result.Id)
	return readJourneyActionTemplate(ctx, data, i)
}

func readJourneyActionTemplate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	journeyApi := journeyApiConfig(i)
	cc := consistency_checker.NewConsistencyCheck(ctx, data, i, ResourceJourneyActionTemplate(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Journey Action Template %s", data.Id())
	return util.WithRetriesForRead(ctx, data, func() *retry.RetryError {
		actionTemplate, resp, getErr := journeyApi.GetJourneyActiontemplate(data.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Journey Action Template %s | error: %s", data.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read Journey Action Template %s | error: %s", data.Id(), getErr), resp))
		}
		flattenActionTemplate(data, actionTemplate)
		log.Printf("Read Journey Action Template %s %s", data.Id(), *actionTemplate.Name)
		return cc.CheckState(data)
	})
}

func updateJourneyActionTemplate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	journeyApi := journeyApiConfig(i)
	patchActionTemplate := buildSdkPatchActionTemplate(data)
	log.Printf("Updating Journey Action Template %s", data.Id())
	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		actionTemplate, resp, getErr := journeyApi.GetJourneyActiontemplate(data.Id())
		if getErr != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to read journey action template %s error: %s", data.Id(), getErr), resp)
		}
		patchActionTemplate.Version = actionTemplate.Version
		_, resp, patchErr := journeyApi.PatchJourneyActiontemplate(data.Id(), *patchActionTemplate)
		if patchErr != nil {
			input, _ := util.InterfaceToJson(*patchActionTemplate)
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to update journey action template %s (input: %+v) error: %s", *actionTemplate.Name, input, patchErr), resp)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}
	log.Printf("Updated Journey Action Template %s", data.Id())
	return readJourneyActionTemplate(ctx, data, i)
}

func deleteJourneyActionTemplate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	journeyApi := journeyApiConfig(i)
	name := data.Get("name").(string)
	log.Printf("Deleting Journey Action Template with name %s", name)
	if resp, err := journeyApi.DeleteJourneyActiontemplate(data.Id(), true); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("create journey action template %s error: %s", name, err), resp)
	}
	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := journeyApi.GetJourneyActiontemplate(data.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Journey Action Template %s", data.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting journey action template %s | error: %s", data.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("journey action template %s still exists", data.Id()), resp))
	})
}

func journeyApiConfig(meta interface{}) *platformclientv2.JourneyApi {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	journeyApi := platformclientv2.NewJourneyApiWithConfig(sdkConfig)
	return journeyApi
}
