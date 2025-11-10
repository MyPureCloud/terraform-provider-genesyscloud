package ai_studio_summary_setting

import (
	"context"
	"fmt"

	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"

	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

/*
The genesyscloud_ai_studio_summary_setting_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *aiStudioSummarySettingProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createAiStudioSummarySettingFunc func(ctx context.Context, p *aiStudioSummarySettingProxy, summarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error)
type getAllAiStudioSummarySettingFunc func(ctx context.Context, p *aiStudioSummarySettingProxy, name string) (*[]platformclientv2.Summarysetting, *platformclientv2.APIResponse, error)
type getAiStudioSummarySettingIdByNameFunc func(ctx context.Context, p *aiStudioSummarySettingProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getAiStudioSummarySettingByIdFunc func(ctx context.Context, p *aiStudioSummarySettingProxy, id string) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error)
type updateAiStudioSummarySettingFunc func(ctx context.Context, p *aiStudioSummarySettingProxy, id string, summarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error)
type deleteAiStudioSummarySettingFunc func(ctx context.Context, p *aiStudioSummarySettingProxy, id string) (*platformclientv2.APIResponse, error)

// aiStudioSummarySettingProxy contains all of the methods that call genesys cloud APIs.
type aiStudioSummarySettingProxy struct {
	clientConfig                          *platformclientv2.Configuration
	aIStudioApi                           *platformclientv2.AIStudioApi
	createAiStudioSummarySettingAttr      createAiStudioSummarySettingFunc
	getAllAiStudioSummarySettingAttr      getAllAiStudioSummarySettingFunc
	getAiStudioSummarySettingIdByNameAttr getAiStudioSummarySettingIdByNameFunc
	getAiStudioSummarySettingByIdAttr     getAiStudioSummarySettingByIdFunc
	updateAiStudioSummarySettingAttr      updateAiStudioSummarySettingFunc
	deleteAiStudioSummarySettingAttr      deleteAiStudioSummarySettingFunc
	summaryCache                          rc.CacheInterface[platformclientv2.Summarysetting]
}

// newAiStudioSummarySettingProxy initializes the ai studio summary setting proxy with all of the data needed to communicate with Genesys Cloud
func newAiStudioSummarySettingProxy(clientConfig *platformclientv2.Configuration) *aiStudioSummarySettingProxy {
	api := platformclientv2.NewAIStudioApiWithConfig(clientConfig)
	summaryCache := rc.NewResourceCache[platformclientv2.Summarysetting]()
	return &aiStudioSummarySettingProxy{
		clientConfig:                          clientConfig,
		aIStudioApi:                           api,
		summaryCache:                          summaryCache,
		createAiStudioSummarySettingAttr:      createAiStudioSummarySettingFn,
		getAllAiStudioSummarySettingAttr:      getAllAiStudioSummarySettingFn,
		getAiStudioSummarySettingIdByNameAttr: getAiStudioSummarySettingIdByNameFn,
		getAiStudioSummarySettingByIdAttr:     getAiStudioSummarySettingByIdFn,
		updateAiStudioSummarySettingAttr:      updateAiStudioSummarySettingFn,
		deleteAiStudioSummarySettingAttr:      deleteAiStudioSummarySettingFn,
	}
}

// getAiStudioSummarySettingProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getAiStudioSummarySettingProxy(clientConfig *platformclientv2.Configuration) *aiStudioSummarySettingProxy {
	if internalProxy == nil {
		internalProxy = newAiStudioSummarySettingProxy(clientConfig)
	}

	return internalProxy
}

// createAiStudioSummarySetting creates a Genesys Cloud ai studio summary setting
func (p *aiStudioSummarySettingProxy) createAiStudioSummarySetting(ctx context.Context, aiStudioSummarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.createAiStudioSummarySettingAttr(ctx, p, aiStudioSummarySetting)
}

// getAiStudioSummarySetting retrieves all Genesys Cloud ai studio summary setting
func (p *aiStudioSummarySettingProxy) getAllAiStudioSummarySetting(ctx context.Context, name string) (*[]platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.getAllAiStudioSummarySettingAttr(ctx, p, name)
}

// getAiStudioSummarySettingIdByName returns a single Genesys Cloud ai studio summary setting by a name
func (p *aiStudioSummarySettingProxy) getAiStudioSummarySettingIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getAiStudioSummarySettingIdByNameAttr(ctx, p, name)
}

// getAiStudioSummarySettingById returns a single Genesys Cloud ai studio summary setting by Id
func (p *aiStudioSummarySettingProxy) getAiStudioSummarySettingById(ctx context.Context, id string) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	if summaries := rc.GetCacheItem(p.summaryCache, id); summaries != nil { // Get  the summary form teh cache, if not found then call the API
		return summaries, nil, nil
	}
	return p.getAiStudioSummarySettingByIdAttr(ctx, p, id)
}

// updateAiStudioSummarySetting updates a Genesys Cloud ai studio summary setting
func (p *aiStudioSummarySettingProxy) updateAiStudioSummarySetting(ctx context.Context, id string, aiStudioSummarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.updateAiStudioSummarySettingAttr(ctx, p, id, aiStudioSummarySetting)
}

// deleteAiStudioSummarySetting deletes a Genesys Cloud ai studio summary setting by Id
func (p *aiStudioSummarySettingProxy) deleteAiStudioSummarySetting(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteAiStudioSummarySettingAttr(ctx, p, id)
}

// createAiStudioSummarySettingFn is an implementation function for creating a Genesys Cloud ai studio summary setting
func createAiStudioSummarySettingFn(ctx context.Context, p *aiStudioSummarySettingProxy, aiStudioSummarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.aIStudioApi.PostConversationsSummariesSettings(*aiStudioSummarySetting)
}

// getAllAiStudioSummarySettingFn is the implementation for retrieving all ai studio summary setting in Genesys Cloud
func getAllAiStudioSummarySettingFn(ctx context.Context, p *aiStudioSummarySettingProxy, name string) (*[]platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return sdkGetAllSummarySettingsFn(ctx, p, name)
}

func sdkGetAllSummarySettingsFn(ctx context.Context, p *aiStudioSummarySettingProxy, name string) (*[]platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	var allSummarySettings []platformclientv2.Summarysetting
	const pageSize = 100

	summarySettings, resp, err := p.aIStudioApi.GetConversationsSummariesSettings(name, "", "", "", 1, pageSize)
	if err != nil {
		return nil, resp, err
	}
	if summarySettings.Entities == nil || len(*summarySettings.Entities) == 0 {
		return &allSummarySettings, resp, nil
	}
	for _, summarySetting := range *summarySettings.Entities {
		allSummarySettings = append(allSummarySettings, summarySetting)
	}

	for pageNum := 2; pageNum <= *summarySettings.PageCount; pageNum++ {
		summarySettings, _, err := p.aIStudioApi.GetConversationsSummariesSettings(name, "", "", "", pageNum, pageSize)
		if err != nil {
			return nil, resp, err
		}

		if summarySettings.Entities == nil || len(*summarySettings.Entities) == 0 {
			break
		}

		for _, summarySetting := range *summarySettings.Entities {
			allSummarySettings = append(allSummarySettings, summarySetting)
		}
	}

	return &allSummarySettings, resp, nil
}

// getAiStudioSummarySettingIdByNameFn is an implementation of the function to get a Genesys Cloud ai studio summary setting by name
func getAiStudioSummarySettingIdByNameFn(ctx context.Context, p *aiStudioSummarySettingProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	summarySettings, resp, err := getAllAiStudioSummarySettingFn(ctx, p, name)
	if err != nil {
		return "", resp, false, err
	}

	if summarySettings == nil || len(*summarySettings) == 0 {
		return "", resp, true, fmt.Errorf("no summary setting found with name: %s", name)
	}

	for _, summary := range *summarySettings {
		if summary.Name != nil && *summary.Name == name {
			if summary.Id != nil {
				return *summary.Id, resp, false, nil
			}
			return "", resp, false, fmt.Errorf("summary setting found but has nil ID: %s", name)
		}
	}

	return "", resp, false, fmt.Errorf("unable to find summary setting with name %s", name)
}

// getAiStudioSummarySettingByIdFn is an implementation of the function to get a Genesys Cloud ai studio summary setting by Id
func getAiStudioSummarySettingByIdFn(ctx context.Context, p *aiStudioSummarySettingProxy, id string) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.aIStudioApi.GetConversationsSummariesSetting(id)
}

// updateAiStudioSummarySettingFn is an implementation of the function to update a Genesys Cloud ai studio summary setting
func updateAiStudioSummarySettingFn(ctx context.Context, p *aiStudioSummarySettingProxy, id string, aiStudioSummarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.aIStudioApi.PutConversationsSummariesSetting(id, *aiStudioSummarySetting)
}

// deleteAiStudioSummarySettingFn is an implementation function for deleting a Genesys Cloud ai studio summary setting
func deleteAiStudioSummarySettingFn(ctx context.Context, p *aiStudioSummarySettingProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.aIStudioApi.DeleteConversationsSummariesSetting(id)
}
