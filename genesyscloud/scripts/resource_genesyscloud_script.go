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
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"

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
func createScript(ctx context.Context, d *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getScriptsProxy(sdkConfig)

	filePath := d.Get("filepath").(string)
	scriptName := d.Get("script_name").(string)
	substitutions := d.Get("substitutions").(map[string]any)
	divisionId := d.Get("division_id").(string)

	if fch := d.Get("file_content_hash").(string); fch != "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "file_content_hash will become a read-only attribute in a future release and should not be set",
		})
	}

	log.Printf("Creating script %s", scriptName)
	scriptId, err := proxy.createScript(ctx, filePath, scriptName, divisionId, substitutions)
	if err != nil {
		return append(diags, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to create script '%s': %s", scriptName, err.Error()), err)...)
	}

	fileHash, err := files.HashFileContent(ctx, filePath, S3Enabled)
	if err != nil {
		return append(diags, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("failed to get file content hash: %s | error: %s", filePath, err.Error()), err)...)
	}
	_ = d.Set("file_content_hash", fileHash)

	d.SetId(scriptId)

	log.Printf("Created script %s. ", d.Id())
	return append(diags, readScript(ctx, d, meta)...)
}

// updateScript providers the Terraform resource logic for updating a Script in Genesys Cloud
func updateScript(ctx context.Context, d *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getScriptsProxy(sdkConfig)

	filePath := d.Get("filepath").(string)
	scriptName := d.Get("script_name").(string)
	substitutions := d.Get("substitutions").(map[string]any)
	divisionId := d.Get("division_id").(string)

	if d.HasChange("file_content_hash") {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "file_content_hash will become a read-only attribute in a future release and should not be set",
		})
	}

	log.Printf("Updating script '%s' %s", scriptName, d.Id())

	scriptId, err := proxy.updateScript(ctx, filePath, scriptName, d.Id(), divisionId, substitutions)
	if err != nil {
		return append(diags, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Failed to update script '%s': %s", scriptName, err.Error()), err)...)
	}
	if scriptId != d.Id() {
		log.Printf("ID of script '%s' changed from '%s' to '%s' after update.", scriptName, d.Id(), scriptId)
		d.SetId(scriptId)
	}

	fileHash, err := files.HashFileContent(ctx, filePath, S3Enabled)
	if err != nil {
		return append(diags, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("failed to get file content hash: %s | error: %s", filePath, err.Error()), err)...)
	}
	_ = d.Set("file_content_hash", fileHash)

	log.Printf("Updated script '%s' %s", scriptName, d.Id())
	return append(diags, readScript(ctx, d, meta)...)
}

// readScript contains all logic needed to read resource data from Genesys Cloud
func readScript(ctx context.Context, d *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getScriptsProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceScript(), constants.ConsistencyChecks(), ResourceType)

	diags = append(diags, util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		script, resp, err := proxy.getScriptById(ctx, d.Id())
		if err != nil {
			diagErr := util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read flow %s | error: %s", d.Id(), err), resp)
			if util.IsStatus404(resp) {
				return retry.RetryableError(diagErr)
			}
			return retry.NonRetryableError(diagErr)
		}

		// Check if script is nil (resource was deleted outside Terraform)
		if script == nil {
			d.SetId("") // Remove from state
			return nil
		}

		resourcedata.SetNillableValue(d, "script_name", script.Name)
		resourcedata.SetNillableReferenceDivision(d, "division_id", script.Division)

		log.Printf("Read script %s %s", d.Id(), *script.Name)
		return cc.CheckState(d)
	})...)

	return diags
}

// deleteScript contains all the logic needed to delete a resource from Genesys Cloud
func deleteScript(ctx context.Context, d *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getScriptsProxy(sdkConfig)

	log.Printf("Deleting script %s", d.Id())
	if err := proxy.deleteScript(ctx, d.Id()); err != nil {
		return append(diags, util.BuildDiagnosticError(ResourceType, fmt.Sprintf("failed to delete script %s", d.Id()), err)...)
	}
	log.Printf("Successfully deleted script %s", d.Id())

	return diags
}
