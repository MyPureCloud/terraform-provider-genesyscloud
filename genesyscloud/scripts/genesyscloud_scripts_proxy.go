package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The genesyscloud_scripts_proxy.go file contains all of the logic associated with calling the Genesys cloud API for scripts.
*/
var internalProxy *scriptsProxy

type createScriptFunc func(ctx context.Context, filePath, scriptName string, substitutions map[string]interface{}, p *scriptsProxy) (scriptId string, err error)
type updateScriptFunc func(ctx context.Context, filePath, scriptName, scriptId string, substitutions map[string]interface{}, p *scriptsProxy) (id string, err error)
type getAllPublishedScriptsFunc func(ctx context.Context, p *scriptsProxy) (*[]platformclientv2.Script, *platformclientv2.APIResponse, error)
type publishScriptFunc func(ctx context.Context, p *scriptsProxy, scriptId string) (*platformclientv2.APIResponse, error)
type getScriptsByNameFunc func(ctx context.Context, p *scriptsProxy, scriptName string) ([]platformclientv2.Script, *platformclientv2.APIResponse, error)
type getScriptIdByNameFunc func(ctx context.Context, p *scriptsProxy, name string) (scriptId string, retryable bool, resp *platformclientv2.APIResponse, err error)
type verifyScriptUploadSuccessFunc func(ctx context.Context, p *scriptsProxy, body []byte) (bool, error)
type scriptWasUploadedSuccessfullyFunc func(ctx context.Context, p *scriptsProxy, uploadId string) (bool, *platformclientv2.APIResponse, error)
type getScriptExportUrlFunc func(ctx context.Context, p *scriptsProxy, scriptId string) (string, *platformclientv2.APIResponse, error)
type deleteScriptFunc func(ctx context.Context, p *scriptsProxy, scriptId string) error
type getScriptByIdFunc func(ctx context.Context, p *scriptsProxy, scriptId string) (script *platformclientv2.Script, resp *platformclientv2.APIResponse, err error)
type getPublishedScriptsByNameFunc func(ctx context.Context, p *scriptsProxy, name string) (*[]platformclientv2.Script, *platformclientv2.APIResponse, error)

// scriptsProxy contains all of the method used to interact with the Genesys Scripts SDK
type scriptsProxy struct {
	clientConfig                      *platformclientv2.Configuration
	scriptsApi                        *platformclientv2.ScriptsApi
	basePath                          string
	accessToken                       string
	createScriptAttr                  createScriptFunc
	updateScriptAttr                  updateScriptFunc
	getAllScriptsAttr                 getAllPublishedScriptsFunc
	publishScriptAttr                 publishScriptFunc
	getScriptIdByNameAttr             getScriptIdByNameFunc
	getScriptsByNameAttr              getScriptsByNameFunc
	verifyScriptUploadSuccessAttr     verifyScriptUploadSuccessFunc
	scriptWasUploadedSuccessfullyAttr scriptWasUploadedSuccessfullyFunc
	getScriptExportUrlAttr            getScriptExportUrlFunc
	deleteScriptAttr                  deleteScriptFunc
	getScriptByIdAttr                 getScriptByIdFunc
	getPublishedScriptsByNameAttr     getPublishedScriptsByNameFunc
	scriptCache                       rc.CacheInterface[platformclientv2.Script]
}

// getScriptsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getScriptsProxy(clientConfig *platformclientv2.Configuration) *scriptsProxy {
	if internalProxy == nil {
		internalProxy = newScriptsProxy(clientConfig)
	}
	return internalProxy
}

// newScriptsProxy initializes the Scripts proxy with all of the data needed to communicate with Genesys Cloud
func newScriptsProxy(clientConfig *platformclientv2.Configuration) *scriptsProxy {
	scriptsAPI := platformclientv2.NewScriptsApiWithConfig(clientConfig)
	scriptCache := rc.NewResourceCache[platformclientv2.Script]()
	return &scriptsProxy{
		clientConfig:                      clientConfig,
		scriptsApi:                        scriptsAPI,
		basePath:                          strings.Replace(scriptsAPI.Configuration.BasePath, "api", "apps", -1),
		accessToken:                       scriptsAPI.Configuration.AccessToken,
		createScriptAttr:                  createScriptFn,
		updateScriptAttr:                  updateScriptFn,
		getAllScriptsAttr:                 getAllPublishedScriptsFn,
		publishScriptAttr:                 publishScriptFn,
		getScriptIdByNameAttr:             getScriptIdByNameFn,
		getScriptsByNameAttr:              getScriptsByNameFn,
		verifyScriptUploadSuccessAttr:     verifyScriptUploadSuccessFn,
		scriptWasUploadedSuccessfullyAttr: scriptWasUploadedSuccessfullyFn,
		getScriptExportUrlAttr:            getScriptExportUrlFn,
		deleteScriptAttr:                  deleteScriptFn,
		getScriptByIdAttr:                 getScriptByIdFn,
		getPublishedScriptsByNameAttr:     getPublishedScriptsByNameFn,
		scriptCache:                       scriptCache,
	}
}

// createScript creates a Genesys Cloud Script
func (p *scriptsProxy) createScript(ctx context.Context, filePath, scriptName string, substitutions map[string]interface{}) (string, error) {
	return p.createScriptAttr(ctx, filePath, scriptName, substitutions, p)
}

// updateScript updates a Genesys Cloud Script
func (p *scriptsProxy) updateScript(ctx context.Context, filePath, scriptName, scriptId string, substitutions map[string]interface{}) (string, error) {
	return p.updateScriptAttr(ctx, filePath, scriptName, scriptId, substitutions, p)
}

func (p *scriptsProxy) getAllPublishedScripts(ctx context.Context) (*[]platformclientv2.Script, *platformclientv2.APIResponse, error) {
	return p.getAllScriptsAttr(ctx, p)
}

func (p *scriptsProxy) publishScript(ctx context.Context, scriptId string) (*platformclientv2.APIResponse, error) {
	return p.publishScriptAttr(ctx, p, scriptId)
}

func (p *scriptsProxy) getScriptsByName(ctx context.Context, scriptName string) ([]platformclientv2.Script, *platformclientv2.APIResponse, error) {
	return p.getScriptsByNameAttr(ctx, p, scriptName)
}

func (p *scriptsProxy) getScriptIdByName(ctx context.Context, name string) (string, bool, *platformclientv2.APIResponse, error) {
	return p.getScriptIdByNameAttr(ctx, p, name)
}

func (p *scriptsProxy) verifyScriptUploadSuccess(ctx context.Context, body []byte) (bool, error) {
	return p.verifyScriptUploadSuccessAttr(ctx, p, body)
}

func (p *scriptsProxy) scriptWasUploadedSuccessfully(ctx context.Context, uploadId string) (bool, *platformclientv2.APIResponse, error) {
	return p.scriptWasUploadedSuccessfullyAttr(ctx, p, uploadId)
}

func (p *scriptsProxy) getScriptExportUrl(ctx context.Context, scriptId string) (string, *platformclientv2.APIResponse, error) {
	return p.getScriptExportUrlAttr(ctx, p, scriptId)
}

func (p *scriptsProxy) deleteScript(ctx context.Context, scriptId string) error {
	return p.deleteScriptAttr(ctx, p, scriptId)
}

func (p *scriptsProxy) getScriptById(ctx context.Context, scriptId string) (script *platformclientv2.Script, resp *platformclientv2.APIResponse, err error) {
	return p.getScriptByIdAttr(ctx, p, scriptId)
}

func (p *scriptsProxy) getPublishedScriptsByName(ctx context.Context, name string) (*[]platformclientv2.Script, *platformclientv2.APIResponse, error) {
	return p.getPublishedScriptsByNameAttr(ctx, p, name)
}

// publishScriptFn will publish the script after it has been successfully upload
func publishScriptFn(_ context.Context, p *scriptsProxy, scriptId string) (*platformclientv2.APIResponse, error) {
	publishScriptBody := &platformclientv2.Publishscriptrequestdata{
		ScriptId: &scriptId,
	}
	_, resp, err := p.scriptsApi.PostScriptsPublished("0", *publishScriptBody)
	return resp, err
}

// getAllPublishedScriptsFn returns all published scripts within a Genesys Cloud instance
func getAllPublishedScriptsFn(_ context.Context, p *scriptsProxy) (*[]platformclientv2.Script, *platformclientv2.APIResponse, error) {
	var allScripts []platformclientv2.Script
	var response *platformclientv2.APIResponse
	pageSize := 50
	for pageNum := 1; ; pageNum++ {
		scripts, resp, err := p.scriptsApi.GetScripts(pageSize, pageNum, "", "", "", "", "", "", "", "")
		response = resp
		if err != nil {
			return nil, resp, err
		}

		if scripts.Entities == nil || len(*scripts.Entities) == 0 {
			break
		}

		for _, script := range *scripts.Entities {
			_, resp, err := p.scriptsApi.GetScriptsPublishedScriptId(*script.Id, "")
			response = resp
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
				return nil, resp, fmt.Errorf("failed to retrieve publication status for script id %s.  Err: %v", *script.Id, err)
			}

			allScripts = append(allScripts, script)
		}
	}

	for _, script := range allScripts {
		rc.SetCache(p.scriptCache, *script.Id, script)
	}

	return &allScripts, response, nil
}

// getScriptsByNameFn Retrieves all scripts instances that match the name passed in
func getScriptsByNameFn(_ context.Context, p *scriptsProxy, scriptName string) ([]platformclientv2.Script, *platformclientv2.APIResponse, error) {
	const pageSize = 50
	var (
		scripts            []platformclientv2.Script
		response           *platformclientv2.APIResponse
		processedScriptIds []string
	)

	log.Printf("Retrieving scripts with name '%s'", scriptName)
	for pageNum := 1; ; pageNum++ {
		data, response, err := p.scriptsApi.GetScripts(pageSize, pageNum, "", scriptName, "", "", "", "", "", "")
		if err != nil {
			return scripts, response, err
		}

		if data.Entities == nil || len(*data.Entities) == 0 {
			break
		}

		for _, script := range *data.Entities {
			if *script.Name == scriptName {
				scripts = append(scripts, script)
				processedScriptIds = append(processedScriptIds, *script.Id)
			}
		}
	}

	for pageNum := 1; ; pageNum++ {
		data, response, err := p.scriptsApi.GetScriptsPublished(pageSize, pageNum, "", scriptName, "", "", "", "")
		if err != nil {
			return nil, response, err
		}
		if data.Entities == nil || len(*data.Entities) == 0 {
			break
		}
		for _, script := range *data.Entities {
			if *script.Name == scriptName && !util.StringExists(*script.Id, processedScriptIds) {
				scripts = append(scripts, script)
				processedScriptIds = append(processedScriptIds, *script.Id)
			}
		}
	}

	return scripts, response, nil
}

// createScriptFormData creates the form data attributes to create a script in Genesys Cloud
func (p *scriptsProxy) createScriptFormData(filePath, scriptName, scriptId string) (map[string]io.Reader, error) {
	fileReader, _, err := files.DownloadOrOpenFile(filePath)
	if err != nil {
		return nil, err
	}
	formData := make(map[string]io.Reader, 0)
	formData["file"] = fileReader
	formData["scriptName"] = strings.NewReader(scriptName)
	if scriptId != "" {
		formData["scriptIdToReplace"] = strings.NewReader(scriptId)
	}
	return formData, nil
}

// uploadScriptFile uploads a script file to S3
// For creates, scriptId should be an empty string
func (p *scriptsProxy) uploadScriptFile(filePath, scriptName, scriptId string, substitutions map[string]interface{}) ([]byte, error) {
	formData, err := p.createScriptFormData(filePath, scriptName, scriptId)
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + p.accessToken

	s3Uploader := files.NewS3Uploader(nil, formData, substitutions, headers, "POST", p.basePath+"/uploads/v2/scripter")
	resp, err := s3Uploader.Upload()
	return resp, err
}

// getScriptIdByNameFn is the implementation function for retrieving a script ID by name, if no other scripts have the same name
func getScriptIdByNameFn(ctx context.Context, p *scriptsProxy, name string) (string, bool, *platformclientv2.APIResponse, error) {
	sdkScripts, resp, err := p.getScriptsByName(ctx, name)
	if err != nil {
		return "", false, resp, err
	}
	if len(sdkScripts) > 1 {
		var extraErrorInfo string
		if isNameOfDefaultScript(name) {
			extraErrorInfo = fmt.Sprintf("'%s' is the name of a reserved script in Genesys Cloud that cannot be deleted. Please select another name.", name)
		}
		return "", false, resp, fmt.Errorf("more than one script found with name '%s'. %s", name, extraErrorInfo)
	}
	if len(sdkScripts) == 0 {
		return "", true, resp, fmt.Errorf("no script found with name '%s'", name)
	}
	return *sdkScripts[0].Id, false, resp, nil
}

func isNameOfDefaultScript(name string) bool {
	return name == constants.DefaultOutboundScriptName || name == constants.DefaultInboundScriptName || name == constants.DefaultCallbackScriptName
}

// verifyScriptUploadSuccessFn checks to see if a file has successfully uploaded
func verifyScriptUploadSuccessFn(ctx context.Context, p *scriptsProxy, body []byte) (bool, error) {
	uploadId, err := p.getUploadIdFromBody(body)
	if err != nil {
		return false, err
	}

	maxRetries := 3
	for i := 1; i <= maxRetries; i++ {
		time.Sleep(2 * time.Second)
		isUploadSuccess, _, err := p.scriptWasUploadedSuccessfully(ctx, uploadId)
		if err != nil {
			return false, err
		}
		if isUploadSuccess {
			return true, nil
		}
	}
	return false, nil
}

// getUploadIdFromBody retrieves the upload Id from the json file being uploade
func (p *scriptsProxy) getUploadIdFromBody(body []byte) (string, error) {
	var (
		jsonData interface{}
		uploadId string
	)

	if err := json.Unmarshal(body, &jsonData); err != nil {
		return "", fmt.Errorf("error unmarshalling json: %v", err)
	}

	if jsonMap, ok := jsonData.(map[string]interface{}); ok {
		uploadId = jsonMap["correlationId"].(string)
	}

	return uploadId, nil
}

// scriptWasUploadedSuccessfullyFn checks the Genesys Cloud API to see if the script was successfully uploaded
func scriptWasUploadedSuccessfullyFn(_ context.Context, p *scriptsProxy, uploadId string) (bool, *platformclientv2.APIResponse, error) {

	data, resp, err := p.scriptsApi.GetScriptsUploadStatus(uploadId, false)
	if err != nil {
		return false, resp, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, resp, fmt.Errorf("error calling GetScriptsUploadStatus: %v", resp.Status)
	}
	return *data.Succeeded, resp, nil
}

// getScriptExportUrlFn retrieves the export URL for a targeted script
func getScriptExportUrlFn(_ context.Context, p *scriptsProxy, scriptId string) (string, *platformclientv2.APIResponse, error) {
	var (
		body platformclientv2.Exportscriptrequest
	)

	data, resp, err := p.scriptsApi.PostScriptExport(scriptId, body)
	if err != nil {
		return "", resp, fmt.Errorf("error calling PostScriptExport: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", resp, fmt.Errorf("error calling PostScriptExport: %v", resp.Status)
	}

	return *data.Url, resp, nil
}

// deleteScriptFn deletes a script from Genesys Cloud
func deleteScriptFn(_ context.Context, p *scriptsProxy, scriptId string) error {
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

// getScriptByIdFn retrieves a script by Id
func getScriptByIdFn(_ context.Context, p *scriptsProxy, scriptId string) (script *platformclientv2.Script, resp *platformclientv2.APIResponse, err error) {
	if script := rc.GetCacheItem(p.scriptCache, scriptId); script != nil {
		return script, nil, nil
	}

	script, resp, err = p.scriptsApi.GetScript(scriptId)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return script, resp, nil
}

// getPublishedScriptsByNameFn returns all of the published scripts that match a name.  Note:  Genesys Cloud allows two script to have the same name and published so we have to return all of the published scripts and let the consumer sort it out.
func getPublishedScriptsByNameFn(_ context.Context, p *scriptsProxy, name string) (*[]platformclientv2.Script, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var allPublishedScripts []platformclientv2.Script
	var response *platformclientv2.APIResponse

	for i := 0; ; i++ {
		pageNumber := i + 1
		data, resp, err := p.scriptsApi.GetScriptsPublished(pageSize, pageNumber, "", name, "", "", "", "")
		if err != nil {
			return nil, resp, err
		}
		response = resp
		if data.Entities == nil || len(*data.Entities) == 0 {
			break
		}

		for _, script := range *data.Entities {
			if *script.Name == name {
				allPublishedScripts = append(allPublishedScripts, script)
			}
		}
	}

	return &allPublishedScripts, response, nil
}

// createScriptFn is an implementation function for creating a Genesys Cloud Script
func createScriptFn(ctx context.Context, filePath, scriptName string, substitutions map[string]interface{}, p *scriptsProxy) (string, error) {
	exists, err := scriptExistsWithName(ctx, p, scriptName)
	if err != nil {
		return "", err
	}

	if exists {
		return "", fmt.Errorf("script with name '%s' already exists. Please provide a unique name", scriptName)
	}

	resp, err := p.uploadScriptFile(filePath, scriptName, "", substitutions)
	if err != nil {
		return "", err
	}

	success, err := p.verifyScriptUploadSuccess(ctx, resp)
	if err != nil {
		return "", err
	} else if !success {
		return "", fmt.Errorf("script '%s' failed to upload successfully", scriptName)
	}

	scriptId, _, _, err := p.getScriptIdByName(ctx, scriptName)
	if err != nil {
		return "", err
	}

	if resp, err := p.publishScript(ctx, scriptId); err != nil {
		return "", fmt.Errorf("script '%s' with id '%s' was not successfully published: %v %v", scriptName, scriptId, err, resp)
	}
	return scriptId, nil
}

// updateScriptFn is an implementation function for updating a Genesys Cloud Script
func updateScriptFn(ctx context.Context, filePath, scriptName, scriptId string, substitutions map[string]interface{}, p *scriptsProxy) (string, error) {
	resp, err := p.uploadScriptFile(filePath, scriptName, scriptId, substitutions)
	if err != nil {
		return "", err
	}

	success, err := p.verifyScriptUploadSuccess(ctx, resp)
	if err != nil {
		return "", err
	} else if !success {
		return "", fmt.Errorf("script '%s' failed to upload successfully", scriptName)
	}

	scriptIdAfterUpdate, _, _, err := p.getScriptIdByName(ctx, scriptName)
	if err != nil {
		return "", err
	}

	if resp, err := p.publishScript(ctx, scriptIdAfterUpdate); err != nil {
		return "", fmt.Errorf("script '%s' with id '%s' was not successfully published: %v %v", scriptName, scriptIdAfterUpdate, err, resp)
	}
	return scriptIdAfterUpdate, nil
}

// scriptExistsWithName is a helper method to determine if a script already exists with the name the user is trying to create a script with
func scriptExistsWithName(ctx context.Context, scriptsProxy *scriptsProxy, scriptName string) (bool, error) {
	sdkScripts, _, err := scriptsProxy.getScriptsByName(ctx, scriptName)
	if err != nil {
		return true, err
	}
	if len(sdkScripts) < 1 {
		return false, nil
	}
	return true, nil
}
