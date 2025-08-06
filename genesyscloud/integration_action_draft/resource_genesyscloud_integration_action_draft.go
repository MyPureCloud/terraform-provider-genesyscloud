package integration_action_draft

import (
	"context"
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
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
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Creating integration action draft %s", name)

	draftRequest := buildActionDraftFromResourceData(d)

	draft, resp, err := iap.createIntegrationActionDraft(ctx, *draftRequest)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create integration action draft %s error: %s", name, err), resp)
	}

	d.SetId(*draft.Id)
	log.Printf("Created integration action draft %s %s", name, *draft.Id)

	return readIntegrationActionDraft(ctx, d, meta)
}

func readIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationActionDraft(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading integration action draft %s", d.Id())

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
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read request template for integration action draft %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read request template for integration action draft %s | error: %s", d.Id(), err), resp))
		}

		successTemp, resp, err := iap.getIntegrationActionDraftTemplate(ctx, d.Id(), successTemplateFileName)
		if err != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read success template for integration action draft %s | error: %s", d.Id(), err), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read success template for integration action draft %s | error: %s", d.Id(), err), resp))
		}

		resourcedata.SetNillableValue(d, "name", draft.Name)
		resourcedata.SetNillableValue(d, "category", draft.Category)
		resourcedata.SetNillableValue(d, "integration_id", draft.IntegrationId)
		resourcedata.SetNillableValue(d, "secure", draft.Secure)
		resourcedata.SetNillableValue(d, "config_timeout_seconds", draft.Config.TimeoutSeconds)

		if draft.Contract != nil && draft.Contract.Input != nil && draft.Contract.Input.InputSchema != nil {
			input, err := flattenActionDraftContract(*draft.Contract.Input.InputSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			_ = d.Set("contract_input", input)
		} else {
			_ = d.Set("contract_input", nil)
		}

		if draft.Contract != nil && draft.Contract.Output != nil && draft.Contract.Output.SuccessSchema != nil {
			output, err := flattenActionDraftContract(*draft.Contract.Output.SuccessSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			_ = d.Set("contract_output", output)
		} else {
			_ = d.Set("contract_output", nil)
		}

		if draft.Config != nil && draft.Config.Request != nil {
			draft.Config.Request.RequestTemplate = reqTemp
			_ = d.Set("config_request", flattenActionConfigRequest(*draft.Config.Request))
		} else {
			_ = d.Set("config_request", nil)
		}

		if draft.Config != nil && draft.Config.Response != nil {
			draft.Config.Response.SuccessTemplate = successTemp
			_ = d.Set("config_response", flattenActionConfigResponse(*draft.Config.Response))
		} else {
			_ = d.Set("config_response", nil)
		}

		log.Printf("Read integration action draft %s %s", d.Id(), *draft.Name)
		return cc.CheckState(d)
	})
}

func updateIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)
	name := d.Get("name").(string)
	category := d.Get("category").(string)
	secure := d.Get("secure").(bool)

	log.Printf("Updating integration action draft %s\n", name)

	// retrieve the latest draft version to send with PATCH
	draft, resp, err := iap.getIntegrationActionDraftById(ctx, d.Id())
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to read integration action draft %s error: %s", d.Id(), err), resp)
	}

	contract, diagErr := buildDraftContract(d)
	if diagErr != nil {
		return diag.Errorf("Failed to build contract %s", err)
	}

	_, resp, err = iap.updateIntegrationActionDraft(ctx, d.Id(), platformclientv2.Updatedraftinput{
		Category: &category,
		Name:     &name,
		Config:   buildSdkActionConfig(d),
		Contract: contract,
		Secure:   &secure,
		Version:  draft.Version,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update integration action draft %s error: %s", name, err), resp)
	}

	log.Printf("Update successful for action draft %s\n", name)
	return readIntegrationActionDraft(ctx, d, meta)
}

// deleteIntegrationActionDraft is used by the integration action resource to delete an action from Genesys cloud.
func deleteIntegrationActionDraft(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)
	name := d.Get("name").(string)

	log.Printf("Deleting integration action draft %s", name)

	resp, err := iap.deleteIntegrationActionDraft(ctx, d.Id())
	if err != nil {
		if util.IsStatus404(resp) {
			// Parent integration was probably deleted which caused the action draft to be deleted
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
