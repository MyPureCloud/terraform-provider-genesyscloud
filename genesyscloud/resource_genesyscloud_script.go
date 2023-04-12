package genesyscloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v95/platformclientv2"
)

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
			"filename": {
				Description:  "Path to the script file to upload.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validatePath,
			},
			"file_content_hash": {
				Description: "Hash value of the script file content. Used to detect changes.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"script_name": {
				Description: "Display name for the script. A reliably unique name is recommended. Default value contains unique identifier.",
				Type:        schema.TypeString,
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

		bodyBuf = bytes.Buffer{}
		w       = multipart.NewWriter(&bodyBuf)
	)

	basePath := strings.Replace(scriptsAPI.Configuration.BasePath, "api", "apps", -1)
	postUrl := basePath + "/uploads/v2/scripter"

	fileName := d.Get("filename").(string)
	scriptName := d.Get("script_name").(string)

	// Check if script already exists with this name
	sdkScript, err := getScriptByName(scriptName, meta)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	if sdkScript.Id != nil {
		return diag.Errorf("Script with name '%s' already exists. Please provide a unique name.", scriptName)
	}

	if err := createScriptFormData(fileName, scriptName, &bodyBuf, w); err != nil {
		return diag.Errorf("%v", err)
	}

	// using newrequest
	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, postUrl, &bodyBuf)
	r.Header.Set("Authorization", "Bearer "+scriptsAPI.Configuration.AccessToken)
	r.Header.Set("Content-Type", w.FormDataContentType())

	log.Printf("Creating script '%s'", scriptName)
	resp, err := client.Do(r)
	if err != nil {
		return diag.Errorf("%v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("error: %v", resp.Status)
	}

	time.Sleep(3 * time.Second)

	uploadId, err := getUploadIdFromBody(resp.Body)
	if err != nil {
		return diag.Errorf("Failed to retrieve upload ID from response body: %v", err)
	}

	maxRetries := 3
	for i := 1; i < maxRetries; i++ {
		isUploadSucces, err := scriptWasUploadedSuccessfully(uploadId, meta)
		if err != nil {
			return diag.Errorf("%v", err)
		}
		if isUploadSucces {
			break
		} else if i == maxRetries {
			return diag.Errorf("Script '%s' failed to upload", scriptName)
		}
	}

	sdkScript, err = getScriptByName(scriptName, meta)
	if err != nil {
		return diag.Errorf("%v", err)
	}

	if sdkScript.Id != nil {
		d.SetId(*sdkScript.Id)
	} else {
		return diag.Errorf("Could not locate script '%s'.", scriptName)
	}

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
	return nil
}

func createScriptFormData(fileName, scriptName string, bodyBuf *bytes.Buffer, w *multipart.Writer) error {
	scriptFile, err := os.Open(fileName)
	if err != nil {
		return err
	}

	readers := map[string]io.Reader{
		"file":       scriptFile,
		"scriptName": strings.NewReader(scriptName),
	}

	for key, r := range readers {
		var (
			fw  io.Writer
			err error
		)
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			fw, err = w.CreateFormFile(key, x.Name())
		} else {
			// Add other fields
			fw, err = w.CreateFormField(key)
		}
		if err != nil {
			return err
		}
		if _, err := io.Copy(fw, r); err != nil {
			return err
		}
	}

	w.Close()
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

	return script, nil
}

func getUploadIdFromBody(body io.ReadCloser) (string, error) {
	var (
		jsonData interface{}
		uploadId string
	)

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	json.Unmarshal(bodyBytes, &jsonData)

	if jsonMap, ok := jsonData.(map[string]interface{}); ok {
		uploadId = jsonMap["correlationId"].(string)
	}

	return uploadId, nil
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
		return false, fmt.Errorf("error: %v", resp.Status)
	}

	return *data.Succeeded, nil
}
