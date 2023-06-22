package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v102/platformclientv2"
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
		GetResourcesFunc: GetAllWithPooledClient(getAllScripts),
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

		CreateContext: CreateWithPooledClient(createScript),
		ReadContext:   ReadWithPooledClient(readScript),
		DeleteContext: DeleteWithPooledClient(deleteScript),
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

	exists, err := scriptExistsWithName(scriptName, meta)
	if err != nil {
		return diag.Errorf("%v", err)
	}
	if exists {
		return diag.Errorf("Script with name '%s' already exists. Please provide a unique name.", scriptName)
	}

	formData, err := createScriptFormData(filePath, scriptName)
	if err != nil {
		return diag.Errorf("failed to create form data for script: %v", err)
	}

	headers := make(map[string]string, 0)
	headers["Authorization"] = "Bearer " + accessToken

	s3Uploader := NewS3Uploader(nil, formData, substitutions, headers, "POST", basePath+"/uploads/v2/scripter")
	resp, err := s3Uploader.Upload()
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
	sdkScripts, err := getScriptsWithName(scriptName, meta)
	if err != nil {
		return diag.Errorf("%v", err)
	}
	if len(sdkScripts) > 1 {
		return diag.Errorf("More than one script found with name %s", scriptName)
	}
	if len(sdkScripts) == 0 {
		return diag.Errorf("Script %s not found after creation.", scriptName)
	}

	d.SetId(*sdkScripts[0].Id)

	log.Printf("Created script %s. ", d.Id())
	return readScript(ctx, d, meta)
}

func readScript(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	scriptsApi := platformclientv2.NewScriptsApiWithConfig(sdkConfig)

	return WithRetriesForRead(ctx, d, func() *resource.RetryError {
		script, resp, err := scriptsApi.GetScript(d.Id())
		if err != nil {
			if IsStatus404(resp) {
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

func scriptExistsWithName(scriptName string, meta interface{}) (bool, error) {
	sdkScripts, err := getScriptsWithName(scriptName, meta)
	if err != nil {
		return true, err
	}
	if len(sdkScripts) < 1 {
		return false, nil
	}
	return true, nil
}

func getScriptsWithName(scriptName string, meta interface{}) ([]platformclientv2.Script, error) {
	var (
		scripts []platformclientv2.Script

		sdkConfig  = meta.(*ProviderMeta).ClientConfig
		scriptsApi = platformclientv2.NewScriptsApiWithConfig(sdkConfig)
	)
	log.Printf("Retrieving scripts with name '%s'", scriptName)
	pageSize := 50
	for i := 0; ; i++ {
		pageNumber := i + 1
		data, _, err := scriptsApi.GetScripts(pageSize, pageNumber, "", scriptName, "", "", "", "", "", "")
		if err != nil {
			return scripts, err
		}

		if data.Entities == nil || len(*data.Entities) == 0 {
			break
		}

		for _, script := range *data.Entities {
			if *script.Name == scriptName {
				scripts = append(scripts, script)
			}
		}
	}

	return scripts, nil
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

func createScriptFormData(filePath, scriptName string) (map[string]io.Reader, error) {
	fileReader, _, err := downloadOrOpenFile(filePath)
	if err != nil {
		return nil, err
	}
	formData := make(map[string]io.Reader, 0)
	formData["file"] = fileReader
	formData["scriptName"] = strings.NewReader(scriptName)
	return formData, nil
}
