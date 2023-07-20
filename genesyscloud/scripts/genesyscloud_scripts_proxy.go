package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	files "terraform-provider-genesyscloud/genesyscloud/util/files"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

// ScriptsProxy contains all of the method used to interact with the Genesys Scripts SDK
type ScriptsProxy struct {
	clientConfig *platformclientv2.Configuration
	scriptsApi   *platformclientv2.ScriptsApi
	basePath     string
	accessToken  string
}

// NewScriptsProxy initializes the Scripts proxy with all of the data needed to communicate with Genesys Cloud
func NewScriptsProxy(clientConfig *platformclientv2.Configuration) *ScriptsProxy {
	scriptsAPI := platformclientv2.NewScriptsApiWithConfig(clientConfig)
	return &ScriptsProxy{
		clientConfig: clientConfig,
		scriptsApi:   scriptsAPI,
		basePath:     strings.Replace(scriptsAPI.Configuration.BasePath, "api", "apps", -1),
		accessToken:  scriptsAPI.Configuration.AccessToken,
	}
}

// publishScript will publish the script after it has been successfully uploade
func (p *ScriptsProxy) publishScript(scriptId string) error {
	publishScriptBody := &platformclientv2.Publishscriptrequestdata{
		ScriptId: &scriptId,
	}

	if _, _, err := p.scriptsApi.PostScriptsPublished("0", *publishScriptBody); err != nil {
		return err
	}

	return nil
}

// getAllPublishedScripts returns all published scripts within a Genesys Cloud instance
func (p *ScriptsProxy) getAllPublishedScripts(ctx context.Context) (*[]platformclientv2.Script, error) {
	var allScripts []platformclientv2.Script
	pageSize := 50
	for pageNum := 1; ; pageNum++ {
		scripts, _, err := p.scriptsApi.GetScripts(pageSize, pageNum, "", "", "", "", "", "", "", "")

		if err != nil {
			return nil, err
		}

		if scripts.Entities == nil || len(*scripts.Entities) == 0 {
			break
		}

		for _, script := range *scripts.Entities {
			_, resp, err := p.scriptsApi.GetScriptsPublishedScriptId(*script.Id, "")

			//If the item is not found this indicates it is not published
			if resp.StatusCode == http.StatusNotFound && err == nil {
				log.Printf("Script id %s, script %s name is not published and will not be returned for export", *script.Id, *script.Name)
				continue
			}

			//Some APIs will return an error code even if the response code is a 404.
			if resp.StatusCode == http.StatusNotFound && err != nil {
				log.Printf("Script id %s, script %s name is not published and will not be returned for export.  Also an err was returned on call %s", *script.Id, *script.Name, err)
				continue
			}

			//All other errors should be failed
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve publication status for script id %s.  Err: %v", *script.Id, err)
			}

			allScripts = append(allScripts, script)
		}
	}

	return &allScripts, nil
}

// Retrieves all scripts instances that matrch the name passed in
func (p *ScriptsProxy) getScriptsByName(scriptName string) ([]platformclientv2.Script, error) {
	var scripts []platformclientv2.Script

	log.Printf("Retrieving scripts with name '%s'", scriptName)
	pageSize := 50
	for i := 0; ; i++ {
		pageNumber := i + 1
		data, _, err := p.scriptsApi.GetScripts(pageSize, pageNumber, "", scriptName, "", "", "", "", "", "")
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

// createScriptFormData creates the form data attributes to create a script in Genesys Cloud
func (p *ScriptsProxy) createScriptFormData(filePath, scriptName string) (map[string]io.Reader, error) {
	fileReader, _, err := files.DownloadOrOpenFile(filePath)
	if err != nil {
		return nil, err
	}
	formData := make(map[string]io.Reader, 0)
	formData["file"] = fileReader
	formData["scriptName"] = strings.NewReader(scriptName)
	return formData, nil
}

// uploadScriptFile uploads a script file to S3
func (p *ScriptsProxy) uploadScriptFile(filePath string, scriptName string, substitutions map[string]interface{}) ([]byte, error) {
	formData, err := p.createScriptFormData(filePath, scriptName)
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string, 0)
	headers["Authorization"] = "Bearer " + p.accessToken

	s3Uploader := files.NewS3Uploader(nil, formData, substitutions, headers, "POST", p.basePath+"/uploads/v2/scripter")
	resp, err := s3Uploader.Upload()
	return resp, err
}

// verifyScriptFile checks to see if a file has successfully uploaded
func (p *ScriptsProxy) verifyScriptUploadSuccess(body []byte) (bool, error) {
	uploadId := p.getUploadIdFromBody(body)

	maxRetries := 3
	for i := 1; i < maxRetries; i++ {
		time.Sleep(2 * time.Second)
		isUploadSuccess, err := p.scriptWasUploadedSuccessfully(uploadId)
		if err != nil {
			return false, err
		}
		if isUploadSuccess {
			break
		} else if i == maxRetries {
			return false, nil
		}
	}

	return true, nil
}

// getUploadIdFromBody retrieves the upload Id from the json file being uploade
func (p *ScriptsProxy) getUploadIdFromBody(body []byte) string {
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

// scriptWasUploadedSuccessfully checks the Genesys Cloud API to see if the script was successfully uploaded
func (p *ScriptsProxy) scriptWasUploadedSuccessfully(uploadId string) (bool, error) {

	data, resp, err := p.scriptsApi.GetScriptsUploadStatus(uploadId, false)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("error calling GetScriptsUploadStatus: %v", resp.Status)
	}

	return *data.Succeeded, nil
}

// getScriptExportUrl retrieves the export URL for a targeted script
func (p *ScriptsProxy) getScriptExportUrl(scriptId string) (string, error) {
	var (
		body platformclientv2.Exportscriptrequest
	)

	data, resp, err := p.scriptsApi.PostScriptExport(scriptId, body)
	if err != nil {
		return "", fmt.Errorf("error calling PostScriptExport: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error calling PostScriptExport: %v", resp.Status)
	}

	return *data.Url, nil
}

// deleteScript deletes a script from Genesys Cloud
func (p *ScriptsProxy) deleteScript(ctx context.Context, scriptId string) error {
	fullPath := p.scriptsApi.Configuration.BasePath + "/api/v2/scripts/" + scriptId
	r, _ := http.NewRequest(http.MethodDelete, fullPath, nil)
	r.Header.Set("Authorization", "Bearer "+p.scriptsApi.Configuration.AccessToken)
	r.Header.Set("Content-Type", "application/json")

	log.Printf("Deleting script %s", scriptId)
	client := &http.Client{}
	resp, err := client.Do(r)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {

		return fmt.Errorf("failed to delete script %s: %s", scriptId, resp.Status)
	}

	log.Printf("Successfully deleted script %s", scriptId)
	return nil
}

// getScriptById retrieves a script by Id
func (p *ScriptsProxy) getScriptById(scriptId string) (script *platformclientv2.Script, statusCode int, err error) {
	script, resp, err := p.scriptsApi.GetScript(scriptId)

	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, resp.StatusCode, nil
		}
		return nil, 0, err
	}

	return script, 0, nil
}

// getsPublishedScriptsByName returns all of the published scripts that match a name.  Note:  Genesys Cloud allows two script to have the same name and published so we have to return all of the published scripts and let the consumer sort it out.
func (p *ScriptsProxy) getPublishedScriptsByName(name string) (*[]platformclientv2.Script, error) {
	const pageSize = 100
	var allPublishedScripts []platformclientv2.Script

	for i := 0; ; i++ {
		pageNumber := i + 1
		data, _, err := p.scriptsApi.GetScriptsPublished(pageSize, pageNumber, "", name, "", "", "", "")
		if err != nil {
			return nil, err
		}

		if data.Entities == nil || len(*data.Entities) == 0 {
			break
		}

		for _, script := range *data.Entities {
			if *script.Name == name {
				allPublishedScripts = append(allPublishedScripts, script)
			}
		}
	}

	return &allPublishedScripts, nil
}
