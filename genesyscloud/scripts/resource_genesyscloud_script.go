package scripts

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	files "terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

// createScript providers the Terraform resource logic for creating a Script object
func createScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	filePath := d.Get("filepath").(string)
	scriptName := d.Get("script_name").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})

	log.Printf("Creating script %s", scriptName)
	exists, err := scriptExistsWithName(ctx, scriptsProxy, scriptName)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	if exists {
		return diag.Errorf("Script with name '%s' already exists. Please provide a unique name.", scriptName)
	}

	resp, err := scriptsProxy.uploadScriptFile(filePath, scriptName, substitutions)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	success, err := scriptsProxy.verifyScriptUploadSuccess(ctx, resp)
	if err != nil {
		return diag.Errorf("%v", err)
	} else if !success {
		return diag.Errorf("Script '%s' failed to upload successfully.", scriptName)
	}

	// Retrieve script ID using getScriptByName function
	sdkScripts, err := scriptsProxy.getScriptByName(ctx, scriptName)
	if err != nil {
		return diag.Errorf("%v", err)
	}
	if len(sdkScripts) > 1 {
		return diag.Errorf("More than one script found with name %s", scriptName)
	}
	if len(sdkScripts) == 0 {
		return diag.Errorf("Script %s not found after creation.", scriptName)
	}

	scriptId := *sdkScripts[0].Id
	if err := scriptsProxy.publishScript(ctx, scriptId); err != nil {
		return diag.Errorf("script %s with id id %s was not successfully published due to %s", scriptName, scriptId, err)
	}

	d.SetId(scriptId)

	log.Printf("Created script %s. ", d.Id())
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
			d.Set("script_name", *script.Name)
		}

		log.Printf("Read script %s %s", d.Id(), *script.Name)
		return cc.CheckState()
	})
}

// deleteScript contains all of the logic needed to delete a resource from Genesys Cloud
func deleteScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	scriptId := d.Id()
	err := scriptsProxy.deleteScript(ctx, scriptId)

	if err != nil {
		return diag.Errorf("failed to delete script %s: %s", d.Id(), err)
	}
	log.Printf("Successfully deleted script %s", d.Id())
	return nil
}

// scriptExistsWithName is a helper method to determine if a script already exists with the name the user is trying create a script with
func scriptExistsWithName(ctx context.Context, scriptsProxy *scriptsProxy, scriptName string) (bool, error) {
	sdkScripts, err := scriptsProxy.getScriptByName(ctx, scriptName)
	if err != nil {
		return true, err
	}
	if len(sdkScripts) < 1 {
		return false, nil
	}
	return true, nil
}

// getAllScripts returns all of the published scripts
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

// ScriptResolver is used to download all Genesys Cloud scripts from Genesys Cloud
func ScriptResolver(scriptId, exportDirectory, subDirectory string, configMap map[string]interface{}, meta interface{}) error {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	scriptsProxy := getScriptsProxy(sdkConfig)

	exportFileName := fmt.Sprintf("script-%s.json", scriptId)

	fullPath := path.Join(exportDirectory, subDirectory)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return err
	}
	ctx := context.Background()
	url, err := scriptsProxy.getScriptExportUrl(ctx, scriptId)
	if err != nil {
		return err
	}

	if err := files.DownloadExportFile(fullPath, exportFileName, url); err != nil {
		return err
	}

	// Update filepath field in configMap to point to exported script file
	configMap["filepath"] = path.Join(subDirectory, exportFileName)
	configMap["file_content_hash"] = fmt.Sprintf(`${filesha256("%s")}`, path.Join(subDirectory, exportFileName))

	return err
}
