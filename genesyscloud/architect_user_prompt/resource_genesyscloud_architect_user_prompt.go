package architect_user_prompt

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func getAllUserPrompts(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	proxy := getArchitectUserPromptProxy(clientConfig)

	userPrompts, resp, err := proxy.getAllArchitectUserPrompts(ctx, true, true, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("failed to get user prompts: %s", err), resp)
	}
	for _, userPrompt := range *userPrompts {
		resources[*userPrompt.Id] = &resourceExporter.ResourceMeta{BlockLabel: *userPrompt.Name}
	}

	return resources, nil
}

func createUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)

	name := d.Get("name").(string)
	prompt := buildUserPromptFromResourceData(d)

	log.Printf("Creating user prompt %s", name)
	userPrompt, resp, err := proxy.createArchitectUserPrompt(ctx, prompt)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create user prompt %s: %s", name, err), resp)
	}

	log.Printf("Creating prompt resources. Prompt ID: '%s'", *userPrompt.Id)
	resp, err = proxy.createOrUpdateArchitectUserPromptResources(ctx, d, *userPrompt.Id, true)
	if err != nil {
		// Cleanup user prompt that was created if the resource creation fails
		d.SetId(*userPrompt.Id) // has to be set for the delete function below to work
		if diagErr := deleteUserPrompt(ctx, d, meta); diagErr != nil {
			log.Println(diagErr)
		}
		d.SetId("")

		if resp != nil {
			return util.BuildAPIDiagnosticError(ResourceType, err.Error(), resp)
		}
		return util.BuildDiagnosticError(ResourceType, err.Error(), err)
	}
	log.Printf("Updated prompt resources. Prompt ID: '%s'", *userPrompt.Id)

	d.SetId(*userPrompt.Id)
	log.Printf("Created user prompt %s %s", name, *userPrompt.Id)
	return readUserPrompt(ctx, d, meta)
}

func readUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectUserPrompt(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading User Prompt %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		userPrompt, resp, getErr := proxy.getArchitectUserPrompt(ctx, d.Id(), true, true, nil, true)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read User Prompt %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("failed to read User Prompt %s | error: %s", d.Id(), getErr), resp))
		}

		resourcedata.SetNillableValue(d, "name", userPrompt.Name)
		resourcedata.SetNillableValue(d, "description", userPrompt.Description)
		_ = d.Set("resources", flattenPromptResources(d, userPrompt.Resources))

		log.Printf("Read Audio Prompt %s %s", d.Id(), *userPrompt.Id)
		return cc.CheckState(d)
	})
}

func updateUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)

	name := d.Get("name").(string)
	prompt := buildUserPromptFromResourceData(d)

	log.Printf("Updating user prompt %s", name)
	_, resp, err := proxy.updateArchitectUserPrompt(ctx, d.Id(), prompt)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update user prompt %s: %s", name, err), resp)
	}

	log.Printf("Updating prompt resources. Prompt ID: '%s'", d.Id())
	resp, err = proxy.createOrUpdateArchitectUserPromptResources(ctx, d, d.Id(), false)
	if err != nil {
		if resp != nil {
			return util.BuildAPIDiagnosticError(ResourceType, err.Error(), resp)
		}
		return util.BuildDiagnosticError(ResourceType, err.Error(), err)
	}
	log.Printf("Updated prompt resources. Prompt ID: '%s'", d.Id())

	log.Printf("Updated User Prompt %s", d.Id())
	return readUserPrompt(ctx, d, meta)
}

func deleteUserPrompt(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getArchitectUserPromptProxy(sdkConfig)

	log.Printf("Deleting user prompt %s", name)
	if resp, err := proxy.deleteArchitectUserPrompt(ctx, d.Id(), true); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete user prompt %s: %s", name, err), resp)
	}
	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getArchitectUserPrompt(ctx, d.Id(), false, false, nil, false)
		if err != nil {
			if util.IsStatus404(resp) {
				// User prompt deleted
				log.Printf("Deleted user prompt %s", name)
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("error deleting user prompt %s | error: %s", name, err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("user prompt %s still exists", name), resp))
	})
}
