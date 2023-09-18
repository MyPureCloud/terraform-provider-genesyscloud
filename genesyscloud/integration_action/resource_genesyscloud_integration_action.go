package integration_action

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

func getAllIntegrationActions(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	iap := getIntegrationActionsProxy(clientConfig)

	actions, err := iap.getAllIntegrationActions(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get page of integration actions: %v", err)
	}

	for _, action := range *actions {
		// Don't include "static" actions
		if strings.HasPrefix(*action.Id, "static") {
			continue
		}
		resources[*action.Id] = &resourceExporter.ResourceMeta{Name: *action.Name}
	}

	return resources, nil
}

func createIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	category := d.Get("category").(string)
	integrationId := d.Get("integration_id").(string)
	secure := d.Get("secure").(bool)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Creating integration action %s", name)

	actionContract, diagErr := buildSdkActionContract(d)
	if diagErr != nil {
		return diagErr
	}

	diagErr = gcloud.RetryWhen(gcloud.IsStatus400, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		action, resp, err := iap.createIntegrationAction(ctx, &IntegrationAction{
			Name:          &name,
			Category:      &category,
			IntegrationId: &integrationId,
			Secure:        &secure,
			Contract:      actionContract,
			Config:        buildSdkActionConfig(d),
		})
		if err != nil {
			return resp, diag.Errorf("Failed to create integration action %s: %s", name, err)
		}
		d.SetId(*action.Id)

		log.Printf("Created integration action %s %s", name, *action.Id)
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	return readIntegrationAction(ctx, d, meta)
}

func readIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Reading integration action %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		action, resp, err := iap.getIntegrationActionById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read integration action %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read integration action %s: %s", d.Id(), err))
		}

		// Retrieve config request/response templates
		reqTemp, resp, err := iap.getIntegrationActionTemplate(ctx, d.Id(), reqTemplateFileName)
		if err != nil {
			if gcloud.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read request template for integration action %s: %s", d.Id(), err))
		}

		successTemp, resp, err := iap.getIntegrationActionTemplate(ctx, d.Id(), successTemplateFileName)
		if err != nil {
			if gcloud.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read success template for integration action %s: %s", d.Id(), err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationAction())
		if action.Name != nil {
			d.Set("name", *action.Name)
		} else {
			d.Set("name", nil)
		}

		if action.Category != nil {
			d.Set("category", *action.Category)
		} else {
			d.Set("category", nil)
		}

		if action.IntegrationId != nil {
			d.Set("integration_id", *action.IntegrationId)
		} else {
			d.Set("integration_id", nil)
		}

		if action.Secure != nil {
			d.Set("secure", *action.Secure)
		} else {
			d.Set("secure", nil)
		}

		if action.Config.TimeoutSeconds != nil {
			d.Set("config_timeout_seconds", *action.Config.TimeoutSeconds)
		} else {
			d.Set("config_timeout_seconds", nil)
		}

		if action.Contract != nil && action.Contract.Input != nil && action.Contract.Input.InputSchema != nil {
			input, err := flattenActionContract(*action.Contract.Input.InputSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			d.Set("contract_input", input)
		} else {
			d.Set("contract_input", nil)
		}

		if action.Contract != nil && action.Contract.Output != nil && action.Contract.Output.SuccessSchema != nil {
			output, err := flattenActionContract(*action.Contract.Output.SuccessSchema)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("%v", err))
			}
			d.Set("contract_output", output)
		} else {
			d.Set("contract_output", nil)
		}

		if action.Config != nil && action.Config.Request != nil {
			action.Config.Request.RequestTemplate = reqTemp
			d.Set("config_request", flattenActionConfigRequest(*action.Config.Request))
		} else {
			d.Set("config_request", nil)
		}

		if action.Config != nil && action.Config.Response != nil {
			action.Config.Response.SuccessTemplate = successTemp
			d.Set("config_response", flattenActionConfigResponse(*action.Config.Response))
		} else {
			d.Set("config_response", nil)
		}

		log.Printf("Read integration action %s %s", d.Id(), *action.Name)
		return cc.CheckState()
	})
}

func updateIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	name := d.Get("name").(string)
	category := d.Get("category").(string)

	log.Printf("Updating integration action %s", name)

	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest action version to send with PATCH
		action, resp, err := iap.getIntegrationActionById(ctx, d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to read integration action %s: %s", d.Id(), err)
		}

		_, resp, err = iap.updateIntegrationAction(ctx, d.Id(), &platformclientv2.Updateactioninput{
			Name:     &name,
			Category: &category,
			Version:  action.Version,
			Config:   buildSdkActionConfig(d),
		})
		if err != nil {
			return resp, diag.Errorf("Failed to update integration action %s: %s", name, err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated integration action %s", name)
	return readIntegrationAction(ctx, d, meta)
}

func deleteIntegrationAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	iap := getIntegrationActionsProxy(sdkConfig)

	log.Printf("Deleting integration action %s", name)
	resp, err := iap.deleteIntegrationAction(ctx, d.Id())
	if err != nil {
		if gcloud.IsStatus404(resp) {
			// Parent integration was probably deleted which caused the action to be deleted
			log.Printf("Integration action already deleted %s", d.Id())
			return nil
		}
		return diag.Errorf("Failed to delete Integration action %s: %s", d.Id(), err)
	}

	return gcloud.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := iap.getIntegrationActionById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				// Integration action deleted
				log.Printf("Deleted Integration action %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting integration action %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("integration action %s still exists", d.Id()))
	})
}
