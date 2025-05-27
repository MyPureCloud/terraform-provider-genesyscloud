package scripts

import (
	"context"
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// getAllScripts returns all the published scripts
func getAllScripts(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	scriptsProxy := getScriptsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	scripts, resp, err := scriptsProxy.getAllPublishedScripts(ctx)
	if err != nil {
		return resources, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of scripts error: %s", err), resp)
	}

	for _, script := range *scripts {
		resources[*script.Id] = &resourceExporter.ResourceMeta{BlockLabel: *script.Name}
	}

	return resources, nil
}

// createScript providers the Terraform resource logic for creating a Script object
func createScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	filePath := d.Get("filepath").(string)
	scriptName := d.Get("script_name").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})
	divisionId := d.Get("division_id").(string)
	log.Printf("Creating script %s", scriptName)
	scriptId, err := scriptsProxy.createScript(ctx, filePath, scriptName, divisionId, substitutions)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(scriptId)

	log.Printf("Created script %s. ", d.Id())
	return readScript(ctx, d, meta)
}

func updateScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	filePath := d.Get("filepath").(string)
	scriptName := d.Get("script_name").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})
	divisionId := d.Get("division_id").(string)

	log.Printf("Updating script '%s' %s", scriptName, d.Id())

	scriptId, err := scriptsProxy.updateScript(ctx, filePath, scriptName, d.Id(), divisionId, substitutions)
	if err != nil {
		return diag.FromErr(err)
	}
	if scriptId != d.Id() {
		log.Printf("ID of script '%s' changed from '%s' to '%s' after update.", scriptName, d.Id(), scriptId)
		d.SetId(scriptId)
	}

	log.Printf("Updated script '%s' %s", scriptName, d.Id())
	return readScript(ctx, d, meta)
}

// readScript contains all of the logic needed to read resource data from Genesys Cloud
func readScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceScript(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		script, resp, err := scriptsProxy.getScriptById(ctx, d.Id())
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read flow %s | error: %s", d.Id(), err), resp))
		}

		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read flow %s | error: %s", d.Id(), err), resp))
		}

		if script.Name != nil {
			_ = d.Set("script_name", *script.Name)
		}
		if script.Division != nil && script.Division.Id != nil {
			_ = d.Set("division_id", *script.Division.Id)
		}

		log.Printf("Read script %s %s", d.Id(), *script.Name)
		return cc.CheckState(d)
	})
}

// deleteScript contains all the logic needed to delete a resource from Genesys Cloud
func deleteScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	log.Printf("Deleting script %s", d.Id())
	if err := scriptsProxy.deleteScript(ctx, d.Id()); err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("failed to delete script %s", d.Id()), err)
	}

	log.Printf("Successfully deleted script %s", d.Id())
	return nil
}
