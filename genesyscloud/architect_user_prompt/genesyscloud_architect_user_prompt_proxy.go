package architect_user_prompt

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/files"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *architectUserPromptProxy

type createArchitectUserPromptFunc func(ctx context.Context, p *architectUserPromptProxy, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error)
type getArchitectUserPromptFunc func(ctx context.Context, p *architectUserPromptProxy, id string, includeMediaUris bool, includeResources bool, language []string, checkCache bool) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error)
type getAllArchitectUserPromptsFilterByNameFunc func(ctx context.Context, p *architectUserPromptProxy, includeMediaUris bool, includeResources bool, name string) (*[]platformclientv2.Prompt, *platformclientv2.APIResponse, error)
type getArchitectUserPromptPageCountFunc func(ctx context.Context, p *architectUserPromptProxy, name string) (int, *platformclientv2.APIResponse, error)
type getAllArchitectUserPromptsFunc func(ctx context.Context, p *architectUserPromptProxy, includeMediaUris bool, includeResources bool, name string) (*[]platformclientv2.Prompt, *platformclientv2.APIResponse, error)
type updateArchitectUserPromptFunc func(ctx context.Context, p *architectUserPromptProxy, id string, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error)
type deleteArchitectUserPromptFunc func(ctx context.Context, p *architectUserPromptProxy, id string, allResources bool) (*platformclientv2.APIResponse, error)
type createArchitectUserPromptResourceFunc func(ctx context.Context, p *architectUserPromptProxy, id string, resource platformclientv2.Promptassetcreate) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error)
type createOrUpdateArchitectUserPromptResourcesFunc func(context.Context, *architectUserPromptProxy, *schema.ResourceData, string, bool) (*platformclientv2.APIResponse, error)
type updateArchitectUserPromptResourceFunc func(ctx context.Context, p *architectUserPromptProxy, id string, languageCode string, body platformclientv2.Promptasset) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error)
type getArchitectUserPromptIdByNameFunc func(ctx context.Context, p *architectUserPromptProxy, name string) (string, *platformclientv2.APIResponse, error, bool)
type uploadPromptFileFunc func(ctx context.Context, p *architectUserPromptProxy, uploadUri, filename string) error
type getArchitectUserPromptResourcesFunc func(ctx context.Context, p *architectUserPromptProxy, promptId string) (*[]platformclientv2.Promptasset, *platformclientv2.APIResponse, error)

// ArchitectUserPromptProxy - proxy for Architect User Prompts
type architectUserPromptProxy struct {
	clientConfig                                   *platformclientv2.Configuration
	architectApi                                   *platformclientv2.ArchitectApi
	createArchitectUserPromptAttr                  createArchitectUserPromptFunc
	getArchitectUserPromptAttr                     getArchitectUserPromptFunc
	getAllArchitectUserPromptsFilterByNameAttr     getAllArchitectUserPromptsFilterByNameFunc
	getArchitectUserPromptPageCountAttr            getArchitectUserPromptPageCountFunc
	getAllArchitectUserPromptsAttr                 getAllArchitectUserPromptsFunc
	updateArchitectUserPromptAttr                  updateArchitectUserPromptFunc
	deleteArchitectUserPromptAttr                  deleteArchitectUserPromptFunc
	createArchitectUserPromptResourceAttr          createArchitectUserPromptResourceFunc
	updateArchitectUserPromptResourceAttr          updateArchitectUserPromptResourceFunc
	createOrUpdateArchitectUserPromptResourcesAttr createOrUpdateArchitectUserPromptResourcesFunc
	getArchitectUserPromptIdByNameAttr             getArchitectUserPromptIdByNameFunc
	uploadPromptFileAttr                           uploadPromptFileFunc
	getArchitectUserPromptResourcesAttr            getArchitectUserPromptResourcesFunc
	promptCache                                    rc.CacheInterface[platformclientv2.Prompt]
}

func newArchitectUserPromptProxy(clientConfig *platformclientv2.Configuration) *architectUserPromptProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	promptCache := rc.NewResourceCache[platformclientv2.Prompt]()
	return &architectUserPromptProxy{
		clientConfig:                                   clientConfig,
		architectApi:                                   api,
		createArchitectUserPromptAttr:                  createArchitectUserPromptFn,
		getArchitectUserPromptAttr:                     getArchitectUserPromptFn,
		getAllArchitectUserPromptsFilterByNameAttr:     getAllArchitectUserPromptsFilterByNameFn,
		getArchitectUserPromptPageCountAttr:            getArchitectUserPromptPageCountFn,
		getAllArchitectUserPromptsAttr:                 getAllArchitectUserPromptsFn,
		updateArchitectUserPromptAttr:                  updateArchitectUserPromptFn,
		deleteArchitectUserPromptAttr:                  deleteArchitectUserPromptFn,
		createArchitectUserPromptResourceAttr:          createArchitectUserPromptResourceFn,
		updateArchitectUserPromptResourceAttr:          updateArchitectUserPromptResourceFn,
		createOrUpdateArchitectUserPromptResourcesAttr: createOrUpdateArchitectUserPromptResourcesFn,
		getArchitectUserPromptIdByNameAttr:             getArchitectUserPromptIdByNameFn,
		uploadPromptFileAttr:                           uploadPromptFileFn,
		getArchitectUserPromptResourcesAttr:            getArchitectUserPromptResourcesFn,
		promptCache:                                    promptCache,
	}
}

func getArchitectUserPromptProxy(clientConfig *platformclientv2.Configuration) *architectUserPromptProxy {
	if internalProxy == nil {
		internalProxy = newArchitectUserPromptProxy(clientConfig)
	}

	return internalProxy
}

// createArchitectUserPrompt creates a new user prompt
func (p *architectUserPromptProxy) createArchitectUserPrompt(ctx context.Context, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	return p.createArchitectUserPromptAttr(ctx, p, body)
}

// getArchitectUserPrompt retrieves a user prompt
func (p *architectUserPromptProxy) getArchitectUserPrompt(ctx context.Context, id string, includeMediaUris, includeResources bool, languages []string, checkCache bool) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	return p.getArchitectUserPromptAttr(ctx, p, id, includeMediaUris, includeResources, languages, checkCache)
}

func (p *architectUserPromptProxy) getAllArchitectUserPromptsFilterByName(ctx context.Context, includeMediaUris, includeResources bool, name string) (*[]platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	return p.getAllArchitectUserPromptsFilterByNameAttr(ctx, p, includeMediaUris, includeResources, name)
}

func (p *architectUserPromptProxy) getArchitectUserPromptPageCount(ctx context.Context, name string) (int, *platformclientv2.APIResponse, error) {
	return p.getArchitectUserPromptPageCountAttr(ctx, p, name)
}

// getAllArchitectUserPrompts retrieves a list of user prompts
func (p *architectUserPromptProxy) getAllArchitectUserPrompts(ctx context.Context, includeMediaUris, includeResources bool, name string) (*[]platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	return p.getAllArchitectUserPromptsAttr(ctx, p, includeMediaUris, includeResources, name)
}

// updateArchitectUserPrompt updates a user prompt
func (p *architectUserPromptProxy) updateArchitectUserPrompt(ctx context.Context, id string, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	return p.updateArchitectUserPromptAttr(ctx, p, id, body)
}

// deleteArchitectUserPrompt deletes a user prompt
func (p *architectUserPromptProxy) deleteArchitectUserPrompt(ctx context.Context, id string, allResources bool) (*platformclientv2.APIResponse, error) {
	return p.deleteArchitectUserPromptAttr(ctx, p, id, allResources)
}

// createArchitectUserPromptResource creates a new user prompt resource
func (p *architectUserPromptProxy) createArchitectUserPromptResource(ctx context.Context, id string, resource platformclientv2.Promptassetcreate) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	return p.createArchitectUserPromptResourceAttr(ctx, p, id, resource)
}

func (p *architectUserPromptProxy) createOrUpdateArchitectUserPromptResources(ctx context.Context, d *schema.ResourceData, promptId string, create bool) (*platformclientv2.APIResponse, error) {
	return p.createOrUpdateArchitectUserPromptResourcesAttr(ctx, p, d, promptId, create)
}

// updateArchitectUserPromptResource updates a user prompt resource
func (p *architectUserPromptProxy) updateArchitectUserPromptResource(ctx context.Context, id, languageCode string, body platformclientv2.Promptasset) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	return p.updateArchitectUserPromptResourceAttr(ctx, p, id, languageCode, body)
}

// getArchitectUserPromptIdByName retrieves a user prompt by name
func (p *architectUserPromptProxy) getArchitectUserPromptIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, error, bool) {
	return p.getArchitectUserPromptIdByNameAttr(ctx, p, name)
}

func (p *architectUserPromptProxy) getArchitectUserPromptResources(ctx context.Context, promptId string) (*[]platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	return p.getArchitectUserPromptResourcesAttr(ctx, p, promptId)
}

func (p *architectUserPromptProxy) uploadPromptFile(ctx context.Context, uploadUri, filename string) error {
	return p.uploadPromptFileAttr(ctx, p, uploadUri, filename)
}

func createArchitectUserPromptFn(_ context.Context, p *architectUserPromptProxy, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	return p.architectApi.PostArchitectPrompts(body)
}

func getArchitectUserPromptFn(_ context.Context, p *architectUserPromptProxy, id string, includeMediaUris, includeResources bool, languages []string, checkCache bool) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	if prompt := rc.GetCacheItem(p.promptCache, id); prompt != nil && checkCache {
		return prompt, nil, nil
	}
	return p.architectApi.GetArchitectPrompt(id, includeMediaUris, includeResources, languages)
}

func updateArchitectUserPromptFn(_ context.Context, p *architectUserPromptProxy, id string, body platformclientv2.Prompt) (*platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	return p.architectApi.PutArchitectPrompt(id, body)
}

func deleteArchitectUserPromptFn(_ context.Context, p *architectUserPromptProxy, id string, allResources bool) (*platformclientv2.APIResponse, error) {
	resp, err := p.architectApi.DeleteArchitectPrompt(id, allResources)
	if err != nil {
		return resp, err
	}
	rc.DeleteCacheItem(p.promptCache, id)
	return nil, nil
}

func getAllArchitectUserPromptsFilterByNameFn(_ context.Context, p *architectUserPromptProxy, includeMediaUris, includeResources bool, exportNameFilter string) (*[]platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var allPrompts []platformclientv2.Prompt
	var response *platformclientv2.APIResponse

	for _, filter := range strings.Split(exportNameFilter, "") {
		userPrompts, response, err := p.architectApi.GetArchitectPrompts(1, pageSize, []string{filter + "*"}, "", "", "", "", includeMediaUris, includeResources, nil)
		if err != nil {
			return nil, response, err
		}

		allPrompts = append(allPrompts, *userPrompts.Entities...)

		pageCount := *userPrompts.PageCount
		if userPrompts.Entities != nil || len(*userPrompts.Entities) != 0 {
			for pageNum := 2; pageNum <= pageCount; pageNum++ {
				userPrompts, response, getErr := p.architectApi.GetArchitectPrompts(pageNum, pageSize, []string{filter + "*"}, "", "", "", "", includeMediaUris, includeResources, nil)
				if getErr != nil {
					return nil, response, getErr
				}

				if userPrompts == nil || userPrompts.Entities == nil || len(*userPrompts.Entities) == 0 {
					break
				}

				allPrompts = append(allPrompts, *userPrompts.Entities...)
			}
		}
	}

	for _, prompt := range allPrompts {
		rc.SetCache(p.promptCache, *prompt.Id, prompt)
	}

	return &allPrompts, response, nil
}

func getArchitectUserPromptPageCountFn(_ context.Context, p *architectUserPromptProxy, name string) (int, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	userPrompts, resp, err := p.architectApi.GetArchitectPrompts(1, pageSize, []string{name}, "", "", "", "", false, false, nil)
	if err != nil {
		return 0, resp, err
	}
	return *userPrompts.PageCount, nil, nil
}

func getAllArchitectUserPromptsFn(_ context.Context, p *architectUserPromptProxy, includeMediaUris, includeResources bool, name string) (*[]platformclientv2.Prompt, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var allPrompts []platformclientv2.Prompt

	userPrompts, response, err := p.architectApi.GetArchitectPrompts(1, pageSize, []string{name}, "", "", "", "", includeMediaUris, includeResources, nil)
	if err != nil {
		return nil, response, err
	}

	if userPrompts.Entities == nil || len(*userPrompts.Entities) == 0 {
		return &allPrompts, response, nil
	}

	allPrompts = append(allPrompts, *userPrompts.Entities...)

	pageCount := *userPrompts.PageCount
	for pageNum := 2; pageNum <= pageCount; pageNum++ {
		userPrompts, response, getErr := p.architectApi.GetArchitectPrompts(pageNum, pageSize, []string{name}, "", "", "", "", includeMediaUris, includeResources, nil)
		if getErr != nil {
			return nil, response, getErr
		}
		if userPrompts == nil || userPrompts.Entities == nil || len(*userPrompts.Entities) == 0 {
			break
		}
		allPrompts = append(allPrompts, *userPrompts.Entities...)
	}

	for _, prompt := range allPrompts {
		rc.SetCache(p.promptCache, *prompt.Id, prompt)
	}

	return &allPrompts, response, nil
}

func createArchitectUserPromptResourceFn(_ context.Context, p *architectUserPromptProxy, id string, promptResource platformclientv2.Promptassetcreate) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	return p.architectApi.PostArchitectPromptResources(id, promptResource)
}

func updateArchitectUserPromptResourceFn(_ context.Context, p *architectUserPromptProxy, id, languageCode string, body platformclientv2.Promptasset) (*platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	return p.architectApi.PutArchitectPromptResource(id, languageCode, body)
}

func createOrUpdateArchitectUserPromptResourcesFn(ctx context.Context, p *architectUserPromptProxy, d *schema.ResourceData, promptId string, create bool) (*platformclientv2.APIResponse, error) {
	var allLanguages []string

	resourcesToCreate, resourcesToUpdate, resp, err := p.buildUserPromptResourcesForCreateAndUpdate(ctx, d, promptId, create)
	if err != nil {
		return resp, err
	}

	for _, r := range resourcesToCreate {
		log.Printf("Creating user prompt resource for language: %s", *r.Language)
		resource, resp, err := p.createArchitectUserPromptResource(ctx, promptId, r)
		if err != nil {
			return resp, fmt.Errorf("failed to create user prompt resource for language '%s': %v", *r.Language, err)
		}

		if err := p.retrieveFilenameAndUploadPromptAsset(ctx, resource); err != nil {
			return nil, err
		}

		allLanguages = append(allLanguages, *r.Language)
	}

	for _, r := range resourcesToUpdate {
		log.Printf("Updating user prompt resource for language: %s", *r.Language)
		resource, resp, err := p.updateArchitectUserPromptResource(ctx, d.Id(), *r.Language, r)
		if err != nil {
			return resp, fmt.Errorf("failed to update user prompt resource for language '%s': %v", *r.Language, err)
		}

		if err := p.retrieveFilenameAndUploadPromptAsset(ctx, resource); err != nil {
			return nil, err
		}

		allLanguages = append(allLanguages, *r.Language)
	}

	return p.verifyPromptResourceFilesAreTranscoded(ctx, promptId, allLanguages)
}

func getArchitectUserPromptResourcesFn(ctx context.Context, p *architectUserPromptProxy, promptId string) (*[]platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var allResources []platformclientv2.Promptasset

	data, resp, err := p.architectApi.GetArchitectPromptResources(promptId, 1, pageSize)
	if err != nil {
		return nil, resp, err
	}
	if data.Entities == nil || len(*data.Entities) == 0 {
		return nil, nil, nil
	}
	allResources = append(allResources, *data.Entities...)

	for pageNum := 2; pageNum <= *data.PageCount; pageNum++ {
		data, resp, err := p.architectApi.GetArchitectPromptResources(promptId, pageNum, pageSize)
		if err != nil {
			return nil, resp, err
		}
		if data.Entities == nil || len(*data.Entities) == 0 {
			break
		}
		allResources = append(allResources, *data.Entities...)
	}

	return &allResources, nil, nil
}

func (p *architectUserPromptProxy) verifyPromptResourceFilesAreTranscoded(ctx context.Context, promptId string, languages []string) (*platformclientv2.APIResponse, error) {
	var response *platformclientv2.APIResponse

	retryErr := util.WithRetries(ctx, 20*time.Second, func() *retry.RetryError {
		userPrompt, resp, err := p.getArchitectUserPrompt(ctx, promptId, true, true, languages, false)
		if err != nil {
			response = resp
			return retry.NonRetryableError(fmt.Errorf("failed to read user prompt '%s': %v", promptId, err))
		}

		if userPrompt == nil || userPrompt.Resources == nil {
			log.Printf("WARNING: User prompt or userPrompt.Resources is nil in the call from getArchitectUserPrompt().  StatusCode returned by the call %d", resp.StatusCode)
			return nil
		}

		for _, APIResource := range *userPrompt.Resources {
			if APIResource.Tags == nil {
				continue
			}
			filenameTag, ok := (*APIResource.Tags)["filename"]
			if !ok {
				continue
			}
			if len(filenameTag) == 0 {
				continue
			}
			if APIResource.UploadStatus != nil && *APIResource.UploadStatus != "transcoded" {
				return retry.RetryableError(fmt.Errorf("prompt file not transcoded. User prompt ID: '%s'. Filename: '%s'", promptId, filenameTag[0]))
			}
		}
		return nil
	})

	if retryErr != nil {
		return response, fmt.Errorf("%v", retryErr)
	}
	return response, nil
}

// retrieveFilenameAndUploadPromptAsset takes a Promptasset struct, finds the file name in the tags, and uses the
// UploadUri to upload the file
func (p *architectUserPromptProxy) retrieveFilenameAndUploadPromptAsset(ctx context.Context, asset *platformclientv2.Promptasset) error {
	if asset.UploadUri == nil || asset.Tags == nil {
		return nil
	}
	filenameTagsArray, ok := (*asset.Tags)["filename"]
	if !ok || len(filenameTagsArray) == 0 || filenameTagsArray[0] == "" {
		return nil
	}
	filename := filenameTagsArray[0]

	if err := p.uploadPromptFile(ctx, *asset.UploadUri, filename); err != nil {
		return fmt.Errorf("failed to upload user prompt resource '%s' to %s", filename, *asset.UploadUri)
	}
	return nil
}

func (p *architectUserPromptProxy) buildUserPromptResourcesForCreateAndUpdate(ctx context.Context, d *schema.ResourceData, promptId string, create bool) ([]platformclientv2.Promptassetcreate, []platformclientv2.Promptasset, *platformclientv2.APIResponse, error) {
	var (
		toCreate []platformclientv2.Promptassetcreate
		toUpdate []platformclientv2.Promptasset

		existingResources *[]platformclientv2.Promptasset
	)

	resources, ok := d.Get("resources").(*schema.Set)
	if !ok || resources == nil {
		return toCreate, toUpdate, nil, nil
	}

	if !create {
		// Look up the existing resources for this prompt
		userPrompt, resp, err := p.getArchitectUserPrompt(ctx, d.Id(), true, true, nil, false)
		if err != nil {
			return toCreate, toUpdate, resp, fmt.Errorf("failed to lookup existing resources for prompt '%s': %v", d.Id(), err)
		}
		existingResources = userPrompt.Resources
	}

	for _, promptResource := range resources.List() {
		languageExists := false
		promptResourceMap, ok := promptResource.(map[string]interface{})
		if !ok {
			continue
		}

		resourceLanguage := promptResourceMap["language"].(string)

		if existingResources != nil {
			// Check if language resource already exists
			for _, r := range *existingResources {
				if *r.Language == resourceLanguage {
					languageExists = true
					break
				}
			}
		}

		if languageExists {
			updateResourceStruct := buildUserPromptResourceForUpdate(promptResourceMap)
			toUpdate = append(toUpdate, *updateResourceStruct)
		} else {
			createResourceStruct := buildUserPromptResourceForCreate(promptResourceMap)
			toCreate = append(toCreate, *createResourceStruct)
		}
	}

	return toCreate, toUpdate, nil, nil
}

// getArchitectUserPromptIdByNameFn will query user prompt by name and retry if search has not yet indexed the user prompt.
func getArchitectUserPromptIdByNameFn(ctx context.Context, p *architectUserPromptProxy, name string) (string, *platformclientv2.APIResponse, error, bool) {
	prompts, response, err := p.getAllArchitectUserPrompts(ctx, true, true, name)
	if err != nil {
		return "", response, err, false
	}
	if prompts == nil {
		return "", response, fmt.Errorf("no prompts found with name '%s'", name), true
	}
	for _, prompt := range *prompts {
		if name == *prompt.Name {
			log.Printf("found user prompt id %s by name %s", *prompt.Id, *prompt.Name)
			return *prompt.Id, response, nil, false
		}
	}
	return "", response, fmt.Errorf("no prompts found with name '%s'", name), true
}

func uploadPromptFileFn(_ context.Context, p *architectUserPromptProxy, uploadUri, filename string) error {
	reader, file, err := files.DownloadOrOpenFile(filename)
	if err != nil {
		return err
	}
	if file != nil {
		defer file.Close()
	}

	formData := make(map[string]io.Reader, 0)
	if file != nil {
		formData["file"] = file
	}

	headers := make(map[string]string)
	headers["Authorization"] = p.clientConfig.AccessToken

	s3Uploader := files.NewS3Uploader(nil, formData, nil, headers, http.MethodPost, uploadUri)

	// for hosted files
	if file == nil {
		part, err := s3Uploader.Writer.CreateFormFile("file", filepath.Base(filename))
		if err != nil {
			return err
		}
		_, err = io.Copy(part, reader)
		if err != nil {
			return err
		}
	}

	_, err = s3Uploader.Upload()
	return err
}
