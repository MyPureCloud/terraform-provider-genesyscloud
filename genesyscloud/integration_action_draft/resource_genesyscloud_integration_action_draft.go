package integration_action_draft

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v154/platformclientv2"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"
)

func getAllIntegrationActionDrafts(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	iap := getIntegrationActionsProxy(clientConfig)

	actions, resp, err := iap.getAllIntegrationActionDrafts(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get integration action drafts %s", err), resp)
	}

	for _, action := range *actions {
		// Don't include "static" actions
		if strings.HasPrefix(*action.Id, "static") {
			continue
		}
		resources[*action.Id] = &resourceExporter.ResourceMeta{BlockLabel: *action.Name}
	}
	return resources, nil
}

func createIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Creating integration action draft %s", name)

	draftRequest := buildActionDraftFromResourceData(d)

	log.Println("Create Contract: ", draftRequest.Contract.Input.String())
	draft, resp, err := iap.createIntegrationActionDraft(ctx, *draftRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create integration action draft %s error: %s", name, err), resp)
	}

	d.SetId(*draft.Id)
	log.Printf("Created integration action draft %s %s", name, *draft.Id)

	fmt.Println("Before: ", d.State())
	return readIntegrationActionDraft(ctx, d, meta)
}

func readIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Reading integration action draft %s", d.Id())
	//cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationActionDraft(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		draft, resp, getErr := iap.getIntegrationActionDraftById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read action draft %s: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read action draft %s: %s", d.Id(), getErr), resp))
		}

		// Retrieve config request/response templates
		reqTemp, resp, err := iap.getIntegrationActionDraftTemplate(ctx, d.Id(), reqTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read request template for integration action draft %s | error: %s", d.Id(), err), resp))
		}

		successTemp, resp, err := iap.getIntegrationActionDraftTemplate(ctx, d.Id(), successTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read success template for integration action draft %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", draft.Name)
		resourcedata.SetNillableValue(d, "category", draft.Category)
		resourcedata.SetNillableValue(d, "integration_id", draft.IntegrationId)
		resourcedata.SetNillableValue(d, "secure", draft.Secure)
		resourcedata.SetNillableValue(d, "config_timeout_seconds", draft.Config.TimeoutSeconds)

		if draft.Contract != nil {
			contract, err := flattenActionDraftContract(*draft.Contract)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			log.Println("Flattened contract", contract)
			_ = d.Set("contract", contract)
		}

		if draft.Config != nil && draft.Config.Request != nil {
			draft.Config.Request.RequestTemplate = reqTemp
			_ = d.Set("config_request", FlattenActionConfigRequest(*draft.Config.Request))
		} else {
			_ = d.Set("config_request", nil)
		}

		if draft.Config != nil && draft.Config.Response != nil {
			draft.Config.Response.SuccessTemplate = successTemp
			_ = d.Set("config_response", FlattenActionConfigResponse(*draft.Config.Response))
		} else {
			_ = d.Set("config_response", nil)
		}

		fmt.Println("After: ", d.State())
		log.Printf("Read integration action draft %s %s", d.Id(), *draft.Name)
		//return cc.CheckState(d)
		return nil
	})
}

func updateIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)
	name := d.Get("name").(string)
	category := d.Get("category").(string)
	secure := d.Get("secure").(bool)
	fmt.Printf("Starting update for action draft %s\n", name)

	draft, resp, err := iap.getIntegrationActionDraftById(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read integration action draft %s error: %s", d.Id(), err), resp)
	}

	_, resp, err = iap.updateIntegrationActionDraft(ctx, d.Id(), platformclientv2.Updatedraftinput{
		Category: &category,
		Name:     &name,
		Config:   buildSdkActionConfig(d),
		Contract: BuildDraftContract(d),
		Secure:   &secure,
		Version:  draft.Version,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update integration action draft %s error: %s", name, err), resp)
	}

	fmt.Printf("Update successful for action draft %s\n", name)
	fmt.Println("After Update: ", d.State())
	return readIntegrationActionDraft(ctx, d, meta)
}

// deleteIntegrationActionDraft is used by the integration action resource to delete an action from Genesys cloud.
func deleteIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Deleting integration action draft %s", name)
	resp, err := iap.deleteIntegrationActionDraft(ctx, d.Id())
	if err != nil {
		if util.IsStatus404(resp) {
			log.Printf("Integration action draft already deleted %s", d.Id())
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Integration action draft %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := iap.getIntegrationActionDraftById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted Integration action draft %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting integration action draft %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("integration action draft %s still exists", d.Id()), resp))
	})
}
