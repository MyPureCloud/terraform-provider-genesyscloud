package scripts

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// getAllScripts returns all the published scripts
func getAllScripts(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	scriptsProxy := getScriptsProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	scripts, err := scriptsProxy.getAllPublishedScripts(ctx)
	if err != nil {
		return resources, diag.Errorf("Failed to get page of scripts: %v", err)
	}

	for _, script := range *scripts {
		resources[*script.Id] = &resourceExporter.ResourceMeta{Name: *script.Name}
	}

	return resources, nil
}

// createScript providers the Terraform resource logic for creating a Script object
func createScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	filePath := d.Get("filepath").(string)
	scriptName := d.Get("script_name").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})

	log.Printf("Creating script %s", scriptName)
	scriptId, err := scriptsProxy.createScript(ctx, filePath, scriptName, substitutions)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(scriptId)

	log.Printf("Created script %s. ", d.Id())
	return readScript(ctx, d, meta)
}

func updateScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	filePath := d.Get("filepath").(string)
	scriptName := d.Get("script_name").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})

	log.Printf("Updating script '%s' %s", scriptName, d.Id())

	scriptId, err := scriptsProxy.updateScript(ctx, filePath, scriptName, d.Id(), substitutions)
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
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		script, statusCode, err := scriptsProxy.getScriptById(ctx, d.Id())
		if statusCode == http.StatusNotFound {
			return retry.RetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
		}

		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceScript())

		if script.Name != nil {
			_ = d.Set("script_name", *script.Name)
		}

		log.Printf("Read script %s %s", d.Id(), *script.Name)
		return cc.CheckState()
	})
}

// deleteScript contains all the logic needed to delete a resource from Genesys Cloud
func deleteScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	log.Printf("Deleting script %s", d.Id())
	if err := scriptsProxy.deleteScript(ctx, d.Id()); err != nil {
		return diag.Errorf("failed to delete script %s: %s", d.Id(), err)
	}

	log.Printf("Successfully deleted script %s", d.Id())
	return nil
}
