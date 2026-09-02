package speechandtextanalytics_settings

import (
	"context"
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
The genesyscloud_speechandtextanalytics_settings_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *speechAndTextAnalyticsSettingsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getSpeechAndTextAnalyticsSettingsFunc func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error)
type updateSpeechAndTextAnalyticsSettingsFunc func(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy, settings *platformclientv2.Speechtextanalyticssettingsrequest) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error)

// speechAndTextAnalyticsSettingsProxy contains all of the methods that call genesys cloud APIs.
type speechAndTextAnalyticsSettingsProxy struct {
	clientConfig                             *platformclientv2.Configuration
	speechTextAnalyticsApi                   *platformclientv2.SpeechTextAnalyticsApi
	getSpeechAndTextAnalyticsSettingsAttr    getSpeechAndTextAnalyticsSettingsFunc
	updateSpeechAndTextAnalyticsSettingsAttr updateSpeechAndTextAnalyticsSettingsFunc
}

// newSpeechAndTextAnalyticsSettingsProxy initializes the proxy with all of the data needed to communicate with Genesys Cloud
func newSpeechAndTextAnalyticsSettingsProxy(clientConfig *platformclientv2.Configuration) *speechAndTextAnalyticsSettingsProxy {
	api := platformclientv2.NewSpeechTextAnalyticsApiWithConfig(clientConfig)
	return &speechAndTextAnalyticsSettingsProxy{
		clientConfig:                             clientConfig,
		speechTextAnalyticsApi:                   api,
		getSpeechAndTextAnalyticsSettingsAttr:    getSpeechAndTextAnalyticsSettingsFn,
		updateSpeechAndTextAnalyticsSettingsAttr: updateSpeechAndTextAnalyticsSettingsFn,
	}
}

// getSpeechAndTextAnalyticsSettingsProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getSpeechAndTextAnalyticsSettingsProxy(clientConfig *platformclientv2.Configuration) *speechAndTextAnalyticsSettingsProxy {
	if internalProxy == nil {
		internalProxy = newSpeechAndTextAnalyticsSettingsProxy(clientConfig)
	}
	return internalProxy
}

// getSpeechAndTextAnalyticsSettings returns the Genesys Cloud organization Speech & Text Analytics settings
func (p *speechAndTextAnalyticsSettingsProxy) getSpeechAndTextAnalyticsSettings(ctx context.Context) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
	return p.getSpeechAndTextAnalyticsSettingsAttr(ctx, p)
}

// updateSpeechAndTextAnalyticsSettings updates the Genesys Cloud organization Speech & Text Analytics settings
func (p *speechAndTextAnalyticsSettingsProxy) updateSpeechAndTextAnalyticsSettings(ctx context.Context, settings *platformclientv2.Speechtextanalyticssettingsrequest) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
	return p.updateSpeechAndTextAnalyticsSettingsAttr(ctx, p, settings)
}

// getSpeechAndTextAnalyticsSettingsFn is an implementation of the function to get the Genesys Cloud Speech & Text Analytics settings
func getSpeechAndTextAnalyticsSettingsFn(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	settings, resp, err := p.speechTextAnalyticsApi.GetSpeechandtextanalyticsSettings()
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve speech and text analytics settings: %s", err)
	}
	return settings, resp, nil
}

// updateSpeechAndTextAnalyticsSettingsFn is an implementation of the function to update the Genesys Cloud Speech & Text Analytics settings
func updateSpeechAndTextAnalyticsSettingsFn(ctx context.Context, p *speechAndTextAnalyticsSettingsProxy, settings *platformclientv2.Speechtextanalyticssettingsrequest) (*platformclientv2.Speechtextanalyticssettingsresponse, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	updated, resp, err := p.speechTextAnalyticsApi.PutSpeechandtextanalyticsSettings(*settings)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update speech and text analytics settings: %s", err)
	}
	return updated, resp, nil
}
