package scripts

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// getAllScripts returns all the published scripts
func getAllScripts(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getScriptsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	scripts, resp, err := proxy.getAllPublishedScripts(ctx)
	if err != nil {
		return resources, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get page of scripts error: %s", err), resp)
	}

	for _, script := range *scripts {
		if isDefaultScriptById(*script.Id) {
			continue
		}
		resources[*script.Id] = &resourceExporter.ResourceMeta{BlockLabel: *script.Name}
	}

	return resources, nil
}

// createScript providers the Terraform resource logic for creating a Script in Genesys Cloud
func createScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getScriptsProxy(sdkConfig)

	filePath := d.Get("filepath").(string)
	scriptName := d.Get("script_name").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})
	divisionId := d.Get("division_id").(string)

	log.Printf("Creating script %s", scriptName)
	scriptId, err := proxy.createScript(ctx, filePath, scriptName, divisionId, substitutions)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to create script '%s': %s", scriptName, err.Error()), err)
	}

	d.SetId(scriptId)

	log.Printf("Created script %s. ", d.Id())
	return readScript(ctx, d, meta)
}

// updateScript providers the Terraform resource logic for updating a Script in Genesys Cloud
func updateScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getScriptsProxy(sdkConfig)

	filePath := d.Get("filepath").(string)
	scriptName := d.Get("script_name").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})
	divisionId := d.Get("division_id").(string)

	log.Printf("Updating script '%s' %s", scriptName, d.Id())

	scriptId, err := proxy.updateScript(ctx, filePath, scriptName, d.Id(), divisionId, substitutions)
	if err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to update script '%s': %s", scriptName, err.Error()), err)
	}
	if scriptId != d.Id() {
		log.Printf("ID of script '%s' changed from '%s' to '%s' after update.", scriptName, d.Id(), scriptId)
		d.SetId(scriptId)
	}

	log.Printf("Updated script '%s' %s", scriptName, d.Id())
	return readScript(ctx, d, meta)
}

// readScript contains all logic needed to read resource data from Genesys Cloud
func readScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getScriptsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceScript(), constants.ConsistencyChecks(), ResourceType)

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		script, resp, err := proxy.getScriptById(ctx, d.Id())
		if err != nil {
			diagErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read flow %s | error: %s", d.Id(), err), resp)
			if util.IsStatus404(resp) {
				return retry.RetryableError(diagErr)
			}
			return retry.NonRetryableError(diagErr)
		}

		resourcedata.SetNillableValue(d, "script_name", script.Name)
		resourcedata.SetNillableReferenceDivision(d, "division_id", script.Division)

		log.Printf("Read script %s %s", d.Id(), *script.Name)
		return cc.CheckState(d)
	})
}

// deleteScript contains all the logic needed to delete a resource from Genesys Cloud
func deleteScript(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getScriptsProxy(sdkConfig)

	log.Printf("Deleting script %s", d.Id())
	if err := proxy.deleteScript(ctx, d.Id()); err != nil {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("failed to delete script %s", d.Id()), err)
	}
	log.Printf("Successfully deleted script %s", d.Id())

	return
}
