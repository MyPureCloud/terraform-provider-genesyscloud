package aistudio_summary_setting

import (
	"context"
	"fmt"
	"log"

	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

/*
The genesyscloud_aistudio_summary_setting_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *aistudioSummarySettingProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createAistudioSummarySettingFunc func(ctx context.Context, p *aistudioSummarySettingProxy, summarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error)
type getAllAistudioSummarySettingFunc func(ctx context.Context, p *aistudioSummarySettingProxy) (*[]platformclientv2.Summarysetting, *platformclientv2.APIResponse, error)
type getAistudioSummarySettingIdByNameFunc func(ctx context.Context, p *aistudioSummarySettingProxy, name string) (string, *platformclientv2.APIResponse, bool, error)
type getAistudioSummarySettingByIdFunc func(ctx context.Context, p *aistudioSummarySettingProxy, id string) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error)
type updateAistudioSummarySettingFunc func(ctx context.Context, p *aistudioSummarySettingProxy, id string, summarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error)
type deleteAistudioSummarySettingFunc func(ctx context.Context, p *aistudioSummarySettingProxy, id string) (*platformclientv2.APIResponse, error)

// aistudioSummarySettingProxy contains all of the methods that call genesys cloud APIs.
type aistudioSummarySettingProxy struct {
	clientConfig                          *platformclientv2.Configuration
	aIStudioApi                           *platformclientv2.AIStudioApi
	createAistudioSummarySettingAttr      createAistudioSummarySettingFunc
	getAllAistudioSummarySettingAttr      getAllAistudioSummarySettingFunc
	getAistudioSummarySettingIdByNameAttr getAistudioSummarySettingIdByNameFunc
	getAistudioSummarySettingByIdAttr     getAistudioSummarySettingByIdFunc
	updateAistudioSummarySettingAttr      updateAistudioSummarySettingFunc
	deleteAistudioSummarySettingAttr      deleteAistudioSummarySettingFunc
}

// newAistudioSummarySettingProxy initializes the aistudio summary setting proxy with all of the data needed to communicate with Genesys Cloud
func newAistudioSummarySettingProxy(clientConfig *platformclientv2.Configuration) *aistudioSummarySettingProxy {
	api := platformclientv2.NewAIStudioApiWithConfig(clientConfig)
	return &aistudioSummarySettingProxy{
		clientConfig:                          clientConfig,
		aIStudioApi:                           api,
		createAistudioSummarySettingAttr:      createAistudioSummarySettingFn,
		getAllAistudioSummarySettingAttr:      getAllAistudioSummarySettingFn,
		getAistudioSummarySettingIdByNameAttr: getAistudioSummarySettingIdByNameFn,
		getAistudioSummarySettingByIdAttr:     getAistudioSummarySettingByIdFn,
		updateAistudioSummarySettingAttr:      updateAistudioSummarySettingFn,
		deleteAistudioSummarySettingAttr:      deleteAistudioSummarySettingFn,
	}
}

// getAistudioSummarySettingProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getAistudioSummarySettingProxy(clientConfig *platformclientv2.Configuration) *aistudioSummarySettingProxy {
	if internalProxy == nil {
		internalProxy = newAistudioSummarySettingProxy(clientConfig)
	}

	return internalProxy
}

// createAistudioSummarySetting creates a Genesys Cloud aistudio summary setting
func (p *aistudioSummarySettingProxy) createAistudioSummarySetting(ctx context.Context, aistudioSummarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.createAistudioSummarySettingAttr(ctx, p, aistudioSummarySetting)
}

// getAistudioSummarySetting retrieves all Genesys Cloud aistudio summary setting
func (p *aistudioSummarySettingProxy) getAllAistudioSummarySetting(ctx context.Context) (*[]platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.getAllAistudioSummarySettingAttr(ctx, p)
}

// getAistudioSummarySettingIdByName returns a single Genesys Cloud aistudio summary setting by a name
func (p *aistudioSummarySettingProxy) getAistudioSummarySettingIdByName(ctx context.Context, name string) (string, *platformclientv2.APIResponse, bool, error) {
	return p.getAistudioSummarySettingIdByNameAttr(ctx, p, name)
}

// getAistudioSummarySettingById returns a single Genesys Cloud aistudio summary setting by Id
func (p *aistudioSummarySettingProxy) getAistudioSummarySettingById(ctx context.Context, id string) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.getAistudioSummarySettingByIdAttr(ctx, p, id)
}

// updateAistudioSummarySetting updates a Genesys Cloud aistudio summary setting
func (p *aistudioSummarySettingProxy) updateAistudioSummarySetting(ctx context.Context, id string, aistudioSummarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.updateAistudioSummarySettingAttr(ctx, p, id, aistudioSummarySetting)
}

// deleteAistudioSummarySetting deletes a Genesys Cloud aistudio summary setting by Id
func (p *aistudioSummarySettingProxy) deleteAistudioSummarySetting(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteAistudioSummarySettingAttr(ctx, p, id)
}

// createAistudioSummarySettingFn is an implementation function for creating a Genesys Cloud aistudio summary setting
func createAistudioSummarySettingFn(ctx context.Context, p *aistudioSummarySettingProxy, aistudioSummarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.aIStudioApi.PostConversationsSummariesSettings(*aistudioSummarySetting)
}

// getAllAistudioSummarySettingFn is the implementation for retrieving all aistudio summary setting in Genesys Cloud
func getAllAistudioSummarySettingFn(ctx context.Context, p *aistudioSummarySettingProxy) (*[]platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	var allSummarySettings []platformclientv2.Summarysetting
	const pageSize = 100

	summarySettings, resp, err := p.aIStudioApi.GetConversationsSummariesSettings("", "", "", "", 1, pageSize)
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
		summarySettings, _, err := p.aIStudioApi.GetConversationsSummariesSettings("", "", "", "", pageNum, pageSize)
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

// getAistudioSummarySettingIdByNameFn is an implementation of the function to get a Genesys Cloud aistudio summary setting by name
func getAistudioSummarySettingIdByNameFn(ctx context.Context, p *aistudioSummarySettingProxy, name string) (string, *platformclientv2.APIResponse, bool, error) {
	summarySettings, resp, err := p.aIStudioApi.GetConversationsSummariesSettings("", name, "", "", 1, 100)
	if err != nil {
		return "", resp, false, err
	}

	if summarySettings.Entities == nil || len(*summarySettings.Entities) == 0 {
		return "", resp, true, err
	}

	for _, summarySetting := range *summarySettings.Entities {
		if *summarySetting.Name == name {
			log.Printf("Retrieved the aistudio summary setting id %s by name %s", *summarySetting.Id, name)
			return *summarySetting.Id, resp, false, nil
		}
	}

	return "", resp, true, fmt.Errorf("Unable to find aistudio summary setting with name %s", name)
}

// getAistudioSummarySettingByIdFn is an implementation of the function to get a Genesys Cloud aistudio summary setting by Id
func getAistudioSummarySettingByIdFn(ctx context.Context, p *aistudioSummarySettingProxy, id string) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.aIStudioApi.GetConversationsSummariesSetting(id)
}

// updateAistudioSummarySettingFn is an implementation of the function to update a Genesys Cloud aistudio summary setting
func updateAistudioSummarySettingFn(ctx context.Context, p *aistudioSummarySettingProxy, id string, aistudioSummarySetting *platformclientv2.Summarysetting) (*platformclientv2.Summarysetting, *platformclientv2.APIResponse, error) {
	return p.aIStudioApi.PutConversationsSummariesSetting(id, *aistudioSummarySetting)
}

// deleteAistudioSummarySettingFn is an implementation function for deleting a Genesys Cloud aistudio summary setting
func deleteAistudioSummarySettingFn(ctx context.Context, p *aistudioSummarySettingProxy, id string) (*platformclientv2.APIResponse, error) {
	return p.aIStudioApi.DeleteConversationsSummariesSetting(id)
}
