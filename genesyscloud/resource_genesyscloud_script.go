package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v95/platformclientv2"
)

func getAllScripts(ctx context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	scriptsAPI := platformclientv2.NewScriptsApiWithConfig(clientConfig)
	pageSize := 50

	for pageNum := 1; ; pageNum++ {
		scripts, _, err := scriptsAPI.GetScripts(pageSize, pageNum, "", "", "", "", "", "", "", "")
		if err != nil {
			return resources, diag.Errorf("Failed to get page of scripts: %v", err)
		}
		if scripts.Entities == nil || len(*scripts.Entities) == 0 {
			break
		}
		for _, script := range *scripts.Entities {
			resources[*script.Id] = &ResourceMeta{Name: *script.Name}
		}
	}

	return resources, nil
}

func scriptExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllScripts),
		RefAttrs:         map[string]*RefAttrSettings{},
		CustomFileWriter: CustomFileWriterSettings{
			RetrieveAndWriteFilesFunc: ScriptResolver,
			SubDirectory:              "scripts",
		},
	}
}

func resourceScript() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Script",

		CreateContext: createWithPooledClient(createScript),
		ReadContext:   readWithPooledClient(readScript),
		DeleteContext: deleteWithPooledClient(deleteScript),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"script_name": {
				Description: "Display name for the script. A reliably unique name is recommended.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"filepath": {
				Description:  "Path to the script file to upload.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validatePath,
				ForceNew:     true,
			},
			"file_content_hash": {
				Description: "Hash value of the script file content. Used to detect changes.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"substitutions": {
				Description: "A substitution is a key value pair where the key is the value you want to replace, and the value is the value to substitute in its place.",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func createScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		sdkConfig  = meta.(*ProviderMeta).ClientConfig
		scriptsAPI = platformclientv2.NewScriptsApiWithConfig(sdkConfig)

		basePath    = strings.Replace(scriptsAPI.Configuration.BasePath, "api", "apps", -1)
		accessToken = scriptsAPI.Configuration.AccessToken
	)

	filePath := d.Get("filepath").(string)
	scriptName := d.Get("script_name").(string)
	substitutions := d.Get("substitutions").(map[string]interface{})

	log.Printf("Creating script %s", scriptName)

	// Check if a script with this name already exists
	if err := scriptExistsWithName(scriptName, meta); err != nil {
		return diag.Errorf("%v", err)
	}

	scriptUploader := NewScriptUploaderObject(filePath, scriptName, basePath, accessToken, substitutions)

	resp, err := scriptUploader.Upload()
	if err != nil {
		return diag.Errorf("%v", err)
	}

	success, err := verifyScriptUploadSuccess(resp, meta)
	if err != nil {
		return diag.Errorf("%v", err)
	} else if !success {
		return diag.Errorf("Script '%s' failed to upload successfully.", scriptName)
	}

	// Retrieve script ID using getAll function
	sdkScript, err := getScriptByName(scriptName, meta)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	d.SetId(*sdkScript.Id)

	log.Printf("Created script %s. ", d.Id())
	return readScript(ctx, d, meta)
}

func readScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	scriptsApi := platformclientv2.NewScriptsApiWithConfig(sdkConfig)

	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		script, resp, err := scriptsApi.GetScript(d.Id())
		if err != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read flow %s: %s", d.Id(), err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceScript())

		if script.Name != nil {
			_ = d.Set("script_name", *script.Name)
		}

		log.Printf("Read script %s %s", d.Id(), *script.Name)
		return cc.CheckState()
	})
}

func deleteScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		sdkConfig  = meta.(*ProviderMeta).ClientConfig
		scriptsApi = platformclientv2.NewScriptsApiWithConfig(sdkConfig)

		fullPath = scriptsApi.Configuration.BasePath + "/api/v2/scripts/" + d.Id()
	)

	r, _ := http.NewRequest(http.MethodDelete, fullPath, nil)
	r.Header.Set("Authorization", "Bearer "+scriptsApi.Configuration.AccessToken)
	r.Header.Set("Content-Type", "application/json")

	log.Printf("Deleting script %s", d.Id())
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return diag.Errorf("failed to delete script %s: %s", d.Id(), err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("failed to delete script %s: %s", d.Id(), resp.Status)
	}

	log.Printf("Successfully deleted script %s", d.Id())
	return nil
}

func scriptExistsWithName(scriptName string, meta interface{}) error {
	sdkScript, err := getScriptByName(scriptName, meta)
	if err != nil {
		return err
	}
	if sdkScript.Id != nil {
		return fmt.Errorf("Script with name '%s' already exists. Please provide a unique name.", scriptName)
	}
	return nil
}

func getScriptByName(scriptName string, meta interface{}) (platformclientv2.Script, error) {
	var (
		script platformclientv2.Script

		sdkConfig  = meta.(*ProviderMeta).ClientConfig
		scriptsApi = platformclientv2.NewScriptsApiWithConfig(sdkConfig)
	)
	log.Printf("Retrieving script by name '%s'", scriptName)
	pageSize := 50
	pageNumber := 1
	data, _, err := scriptsApi.GetScripts(pageSize, pageNumber, "", scriptName, "", "", "", "", "", "")
	if err != nil {
		return script, err
	}

	if data.Entities != nil && len(*data.Entities) > 0 {
		script = (*data.Entities)[0]
	}

	if len(*data.Entities) > 1 {
		return script, fmt.Errorf("more than one script was found with name '%s'. Please use a unique name.", scriptName)
	}

	return script, nil
}

func verifyScriptUploadSuccess(body []byte, meta interface{}) (bool, error) {
	uploadId := getUploadIdFromBody(body)

	maxRetries := 3
	for i := 1; i < maxRetries; i++ {
		time.Sleep(2 * time.Second)
		isUploadSucces, err := scriptWasUploadedSuccessfully(uploadId, meta)
		if err != nil {
			return false, err
		}
		if isUploadSucces {
			break
		} else if i == maxRetries {
			return false, nil
		}
	}

	return true, nil
}

func getUploadIdFromBody(body []byte) string {
	var (
		jsonData interface{}
		uploadId string
	)

	json.Unmarshal(body, &jsonData)

	if jsonMap, ok := jsonData.(map[string]interface{}); ok {
		uploadId = jsonMap["correlationId"].(string)
	}

	return uploadId
}

func scriptWasUploadedSuccessfully(uploadId string, meta interface{}) (bool, error) {
	var (
		sdkConfig  = meta.(*ProviderMeta).ClientConfig
		scriptsApi = platformclientv2.NewScriptsApiWithConfig(sdkConfig)
	)

	data, resp, err := scriptsApi.GetScriptsUploadStatus(uploadId, false)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("error calling GetScriptsUploadStatus: %v", resp.Status)
	}

	return *data.Succeeded, nil
}

func getScriptExportUrl(scriptId string, meta interface{}) (string, error) {
	var (
		sdkConfig  = meta.(*ProviderMeta).ClientConfig
		scriptsApi = platformclientv2.NewScriptsApiWithConfig(sdkConfig)
		body       platformclientv2.Exportscriptrequest
	)

	data, resp, err := scriptsApi.PostScriptExport(scriptId, body)
	if err != nil {
		return "", fmt.Errorf("error calling PostScriptExport: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error calling PostScriptExport: %v", resp.Status)
	}

	return *data.Url, nil
}
