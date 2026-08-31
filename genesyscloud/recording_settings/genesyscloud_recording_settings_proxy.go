package recording_settings

import (
	"context"
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
The genesyscloud_recording_settings_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *recordingSettingsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getRecordingSettingsFunc func(ctx context.Context, p *recordingSettingsProxy) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error)
type updateRecordingSettingsFunc func(ctx context.Context, p *recordingSettingsProxy, settings *platformclientv2.Recordingsettings) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error)

// recordingSettingsProxy contains all of the methods that call genesys cloud APIs.
type recordingSettingsProxy struct {
	clientConfig                *platformclientv2.Configuration
	recordingApi                *platformclientv2.RecordingApi
	getRecordingSettingsAttr    getRecordingSettingsFunc
	updateRecordingSettingsAttr updateRecordingSettingsFunc
}

// newRecordingSettingsProxy initializes the recording settings proxy with all of the data needed to communicate with Genesys Cloud
func newRecordingSettingsProxy(clientConfig *platformclientv2.Configuration) *recordingSettingsProxy {
	api := platformclientv2.NewRecordingApiWithConfig(clientConfig)
	return &recordingSettingsProxy{
		clientConfig:                clientConfig,
		recordingApi:                api,
		getRecordingSettingsAttr:    getRecordingSettingsFn,
		updateRecordingSettingsAttr: updateRecordingSettingsFn,
	}
}

// getRecordingSettingsProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getRecordingSettingsProxy(clientConfig *platformclientv2.Configuration) *recordingSettingsProxy {
	if internalProxy == nil {
		internalProxy = newRecordingSettingsProxy(clientConfig)
	}
	return internalProxy
}

// getRecordingSettings retrieves the Genesys Cloud organization recording settings
func (p *recordingSettingsProxy) getRecordingSettings(ctx context.Context) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
	return p.getRecordingSettingsAttr(ctx, p)
}

// updateRecordingSettings updates the Genesys Cloud organization recording settings
func (p *recordingSettingsProxy) updateRecordingSettings(ctx context.Context, settings *platformclientv2.Recordingsettings) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
	return p.updateRecordingSettingsAttr(ctx, p, settings)
}

// getRecordingSettingsFn is an implementation of the function to get the Genesys Cloud organization recording settings
func getRecordingSettingsFn(ctx context.Context, p *recordingSettingsProxy) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	// createDefault=true ensures a default settings object is returned if none exists yet
	settings, resp, err := p.recordingApi.GetRecordingSettings(true)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve recording settings: %s", err)
	}
	return settings, resp, nil
}

// updateRecordingSettingsFn is an implementation of the function to update the Genesys Cloud organization recording settings
func updateRecordingSettingsFn(ctx context.Context, p *recordingSettingsProxy, settings *platformclientv2.Recordingsettings) (*platformclientv2.Recordingsettings, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	updatedSettings, resp, err := p.recordingApi.PutRecordingSettings(*settings)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update recording settings: %s", err)
	}
	return updatedSettings, resp, nil
}
