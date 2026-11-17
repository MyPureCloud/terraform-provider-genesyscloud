package analytics_dataretention_settings

import (
	"context"
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
)

/*
The genesyscloud_analytics_dataretention_settings_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *analyticsDataretentionSettingsProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAnalyticsDataretentionSettingsFunc func(ctx context.Context, p *analyticsDataretentionSettingsProxy) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error)
type updateAnalyticsDataretentionSettingsFunc func(ctx context.Context, p *analyticsDataretentionSettingsProxy, request *platformclientv2.Updateanalyticsdataretentionrequest) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error)

// analyticsDataretentionSettingsProxy contains all of the methods that call genesys cloud APIs.
type analyticsDataretentionSettingsProxy struct {
	clientConfig                             *platformclientv2.Configuration
	analyticsApi                             *platformclientv2.AnalyticsApi
	getAnalyticsDataretentionSettingsAttr    getAnalyticsDataretentionSettingsFunc
	updateAnalyticsDataretentionSettingsAttr updateAnalyticsDataretentionSettingsFunc
}

// newAnalyticsDataretentionSettingsProxy initializes the analytics data retention settings proxy with all of the data needed to communicate with Genesys Cloud
func newAnalyticsDataretentionSettingsProxy(clientConfig *platformclientv2.Configuration) *analyticsDataretentionSettingsProxy {
	api := platformclientv2.NewAnalyticsApiWithConfig(clientConfig)
	return &analyticsDataretentionSettingsProxy{
		clientConfig:                             clientConfig,
		analyticsApi:                             api,
		getAnalyticsDataretentionSettingsAttr:    getAnalyticsDataretentionSettingsFn,
		updateAnalyticsDataretentionSettingsAttr: updateAnalyticsDataretentionSettingsFn,
	}
}

// getAnalyticsDataretentionSettingsProxy acts as a singleton for the internalProxy. It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getAnalyticsDataretentionSettingsProxy(clientConfig *platformclientv2.Configuration) *analyticsDataretentionSettingsProxy {
	if internalProxy == nil {
		internalProxy = newAnalyticsDataretentionSettingsProxy(clientConfig)
	}
	return internalProxy
}

// getAnalyticsDataretentionSettings retrieves the Genesys Cloud organization analytics data retention setting
func (p *analyticsDataretentionSettingsProxy) getAnalyticsDataretentionSettings(ctx context.Context) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
	return p.getAnalyticsDataretentionSettingsAttr(ctx, p)
}

// updateAnalyticsDataretentionSettings updates the Genesys Cloud organization analytics data retention setting
func (p *analyticsDataretentionSettingsProxy) updateAnalyticsDataretentionSettings(ctx context.Context, request *platformclientv2.Updateanalyticsdataretentionrequest) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
	return p.updateAnalyticsDataretentionSettingsAttr(ctx, p, request)
}

// getAnalyticsDataretentionSettingsFn is an implementation of the function to get the Genesys Cloud organization analytics data retention setting
func getAnalyticsDataretentionSettingsFn(ctx context.Context, p *analyticsDataretentionSettingsProxy) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	settings, resp, err := p.analyticsApi.GetAnalyticsDataretentionSettings()
	if err != nil {
		return nil, resp, fmt.Errorf("failed to retrieve analytics data retention settings: %s", err)
	}
	return settings, resp, nil
}

// updateAnalyticsDataretentionSettingsFn is an implementation of the function to update the Genesys Cloud organization analytics data retention setting
func updateAnalyticsDataretentionSettingsFn(ctx context.Context, p *analyticsDataretentionSettingsProxy, request *platformclientv2.Updateanalyticsdataretentionrequest) (*platformclientv2.Analyticsdataretentionresponse, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	updatedSettings, resp, err := p.analyticsApi.PutAnalyticsDataretentionSettings(*request)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update analytics data retention settings: %s", err)
	}
	return updatedSettings, resp, nil
}
