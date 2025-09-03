package scripts

import (
	"context"
	"encoding/json"
	"fmt"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/files"

	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

/*
The genesyscloud_scripts_proxy.go file contains all of the logic associated with calling the Genesys cloud API for scripts.
*/
type createScriptFunc func(ctx context.Context, filePath, scriptName, divisionId string, substitutions map[string]interface{}, p *scriptsProxy) (scriptId string, err error)
type updateScriptFunc func(ctx context.Context, filePath, scriptName, scriptId, divisionId string, substitutions map[string]interface{}, p *scriptsProxy) (id string, err error)
type getAllPublishedScriptsFunc func(ctx context.Context, p *scriptsProxy) (*[]platformclientv2.Script, *platformclientv2.APIResponse, error)
type publishScriptFunc func(ctx context.Context, p *scriptsProxy, scriptId string) (*platformclientv2.APIResponse, error)
type getScriptsByNameFunc func(ctx context.Context, p *scriptsProxy, scriptName string) ([]platformclientv2.Script, *platformclientv2.APIResponse, error)
type getScriptIdByNameFunc func(ctx context.Context, p *scriptsProxy, name string) (scriptId string, retryable bool, resp *platformclientv2.APIResponse, err error)
type verifyScriptUploadSuccessFunc func(ctx context.Context, p *scriptsProxy, body []byte) (bool, error)
type scriptWasUploadedSuccessfullyFunc func(ctx context.Context, p *scriptsProxy, uploadId string) (bool, *platformclientv2.APIResponse, error)
type getScriptExportUrlFunc func(ctx context.Context, p *scriptsProxy, scriptId string) (string, *platformclientv2.APIResponse, error)
type deleteScriptFunc func(ctx context.Context, p *scriptsProxy, scriptId string) error
type getScriptByIdFunc func(ctx context.Context, p *scriptsProxy, scriptId string) (script *platformclientv2.Script, resp *platformclientv2.APIResponse, err error)

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
	scriptCache                       rc.CacheInterface[platformclientv2.Script]
}

var scriptCache = rc.NewResourceCache[platformclientv2.Script]()

// getScriptsProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
// (abandoned singleton pattern for DEVTOOLING-1081)
func getScriptsProxy(clientConfig *platformclientv2.Configuration) *scriptsProxy {
	return newScriptsProxy(clientConfig)
}

// newScriptsProxy initializes the Scripts proxy with all of the data needed to communicate with Genesys Cloud
func newScriptsProxy(clientConfig *platformclientv2.Configuration) *scriptsProxy {
	scriptsAPI := platformclientv2.NewScriptsApiWithConfig(clientConfig)
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
		scriptCache:                       scriptCache,
	}
}

// createScript creates a Genesys Cloud Script
func (p *scriptsProxy) createScript(ctx context.Context, filePath, scriptName, divisionId string, substitutions map[string]interface{}) (string, error) {
	return p.createScriptAttr(ctx, filePath, scriptName, divisionId, substitutions, p)
}

// updateScript updates a Genesys Cloud Script
func (p *scriptsProxy) updateScript(ctx context.Context, filePath, scriptName, scriptId, divisionId string, substitutions map[string]interface{}) (string, error) {
	return p.updateScriptAttr(ctx, filePath, scriptName, scriptId, divisionId, substitutions, p)
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

// publishScriptFn will publish the script after it has been successfully upload
func publishScriptFn(_ context.Context, p *scriptsProxy, scriptId string) (*platformclientv2.APIResponse, error) {
	publishScriptBody := &platformclientv2.Publishscriptrequestdata{
		ScriptId: &scriptId,
	}
	_, resp, err := p.scriptsApi.PostScriptsPublished("0", *publishScriptBody)
	return resp, err
}

// getAllPublishedScriptsFn returns all published scripts within a Genesys Cloud org
func getAllPublishedScriptsFn(_ context.Context, p *scriptsProxy) (*[]platformclientv2.Script, *platformclientv2.APIResponse, error) {
	var allPublishedScripts []platformclientv2.Script
	var resp *platformclientv2.APIResponse
	const pageSize = 100

	data, resp, err := p.scriptsApi.GetScriptsPublished(pageSize, 1, "", "", "", "", "", "")
	if err != nil {
		return nil, resp, err
	}
	if data.Entities == nil || len(*data.Entities) == 0 {
		return &allPublishedScripts, resp, nil
	}

	allPublishedScripts = append(allPublishedScripts, *data.Entities...)

	pageCount := *data.PageCount
	for pageNum := 2; pageNum <= pageCount; pageNum++ {
		data, resp, err = p.scriptsApi.GetScriptsPublished(pageSize, pageNum, "", "", "", "", "", "")
		if err != nil {
			return nil, resp, err
		}
		if data.Entities == nil || len(*data.Entities) == 0 {
			continue
		}
		allPublishedScripts = append(allPublishedScripts, *data.Entities...)
	}

	for _, script := range allPublishedScripts {
		rc.SetCache(p.scriptCache, *script.Id, script)
	}

	return &allPublishedScripts, resp, nil
}

// getScriptsByNameFn Retrieves all scripts instances that match the name passed in
func getScriptsByNameFn(_ context.Context, p *scriptsProxy, scriptName string) (scriptsThatMatchName []platformclientv2.Script, resp *platformclientv2.APIResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("getScriptsByNameFn: %w", err)
		}
	}()

	const pageSize = 100
	var (
		getScriptsRespBody *platformclientv2.Scriptentitylisting
		processedScriptIds []string
	)

	log.Printf("Retrieving scripts with name '%s'", scriptName)
	for pageNum := 1; ; pageNum++ {
		getScriptsRespBody, resp, err = p.scriptsApi.GetScripts(pageSize, pageNum, "", scriptName, "", "", "", "", "", "")
		if err != nil {
			return nil, resp, err
		}
		if getScriptsRespBody.Entities == nil || len(*getScriptsRespBody.Entities) == 0 {
			break
		}

		for _, script := range *getScriptsRespBody.Entities {
			if *script.Name == scriptName {
				scriptsThatMatchName = append(scriptsThatMatchName, script)
				processedScriptIds = append(processedScriptIds, *script.Id)
			}
		}
	}

	log.Printf("Retrieving published scripts with name '%s'", scriptName)
	for pageNum := 1; ; pageNum++ {
		getScriptsRespBody, resp, err = p.scriptsApi.GetScriptsPublished(pageSize, pageNum, "", scriptName, "", "", "", "")
		if err != nil {
			return nil, resp, err
		}
		if getScriptsRespBody.Entities == nil || len(*getScriptsRespBody.Entities) == 0 {
			break
		}
		for _, script := range *getScriptsRespBody.Entities {
			if *script.Name == scriptName && !util.StringExists(*script.Id, processedScriptIds) {
				scriptsThatMatchName = append(scriptsThatMatchName, script)
				processedScriptIds = append(processedScriptIds, *script.Id)
			}
		}
	}

	return scriptsThatMatchName, resp, err
}

// createScriptFormData creates the form data attributes to create a script in Genesys Cloud
func (p *scriptsProxy) createScriptFormData(filePath, scriptName, scriptId string) (map[string]io.Reader, error) {
	fileReader, _, err := files.DownloadOrOpenFile(context.Background(), filePath, S3Enabled)
	if err != nil {
		return nil, err
	}
	formData := make(map[string]io.Reader)
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
func getScriptIdByNameFn(ctx context.Context, p *scriptsProxy, name string) (_ string, _ bool, _ *platformclientv2.APIResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("getScriptIdByNameFn: %w", err)
		}
	}()

	sdkScripts, resp, err := p.getScriptsByName(ctx, name)
	if err != nil {
		return "", false, resp, err
	}
	if len(sdkScripts) > 1 {
		var extraErrorInfo string
		if isDefaultScriptByName(name) {
			extraErrorInfo = fmt.Sprintf("'%s' is the name of a reserved script in Genesys Cloud that cannot be deleted. Please select another name.", name)
		}
		return "", false, resp, fmt.Errorf("more than one script found with name '%s'. %s", name, extraErrorInfo)
	}
	if len(sdkScripts) == 0 {
		return "", true, resp, fmt.Errorf("no script found with name '%s'", name)
	}
	return *sdkScripts[0].Id, false, resp, nil
}

func isDefaultScriptByName(name string) bool {
	return name == constants.DefaultOutboundScriptName || name == constants.DefaultInboundScriptName || name == constants.DefaultCallbackScriptName
}

func isDefaultScriptById(id string) bool {
	return id == constants.DefaultCallbackScriptID || id == constants.DefaultOutboundScriptID || id == constants.DefaultInboundScriptID
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

// getUploadIdFromBody retrieves the upload Id from the json file being uploaded
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

	// Sets the VersionId on the request so that the Published Version of the script is exported and not the editable version
	// See DEVTOOLING-777
	scriptCache := rc.GetCacheItem(p.scriptCache, scriptId)
	body.VersionId = scriptCache.VersionId

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

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("Failed to delete script '%s' because it does not exist", scriptId)
		return nil
	}

	if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
		response := "nil"
		if resp != nil {
			response = resp.Status
		}
		return fmt.Errorf("failed to delete script %s. API response: %s. Error: %v", scriptId, response, err)
	}

	rc.DeleteCacheItem(p.scriptCache, scriptId)
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

// createScriptFn is an implementation function for creating a Genesys Cloud Script
func createScriptFn(ctx context.Context, filePath, scriptName, divisionId string, substitutions map[string]interface{}, p *scriptsProxy) (string, error) {
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

	setDivisionErr := setScriptDivision(scriptId, divisionId, p)
	if setDivisionErr != nil {
		return "", setDivisionErr
	}

	if resp, err := p.publishScript(ctx, scriptId); err != nil {
		// If the script cannot be published, clean up the script instance on the API before throwing an error
		// See DEVTOOLING-777
		log.Printf("Attempting to delete script '%s'", scriptId)
		deleteErr := p.deleteScript(ctx, scriptId)
		if deleteErr != nil {
			log.Printf("Error occurred while trying to delete script '%s': %s", scriptId, deleteErr.Error())
		}
		return "", fmt.Errorf("script '%s' (ID: %s) failed to publish and was deleted: %w (response: %v)", scriptName, scriptId, err, resp)
	}
	return scriptId, nil
}

// updateScriptFn is an implementation function for updating a Genesys Cloud Script
func updateScriptFn(ctx context.Context, filePath, scriptName, scriptId, divisionId string, substitutions map[string]interface{}, p *scriptsProxy) (string, error) {
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
	setDivisionErr := setScriptDivision(scriptId, divisionId, p)
	if setDivisionErr != nil {
		return "", setDivisionErr
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

func setScriptDivision(scriptId, divisionId string, p *scriptsProxy) error {
	if divisionId == "" {
		return nil
	}
	apiClient := &p.scriptsApi.Configuration.APIClient
	action := http.MethodPost
	fullPath := p.scriptsApi.Configuration.BasePath + "/api/v2/authorization/divisions/" + divisionId + "/objects/SCRIPT"
	body := []string{scriptId}

	headerParams := make(map[string]string)

	for key := range p.scriptsApi.Configuration.DefaultHeader {
		headerParams[key] = p.scriptsApi.Configuration.DefaultHeader[key]
	}
	headerParams["Authorization"] = "Bearer " + p.scriptsApi.Configuration.AccessToken
	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	response, err := apiClient.CallAPI(fullPath, action, body, headerParams, nil, nil, "", nil, "")

	if err != nil || response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to set divisionId script %s: status code %d due to %s", scriptId, response.StatusCode, response.ErrorMessage)
	}

	log.Printf("successfully set divisionId for script %s", scriptId)
	return nil
}
