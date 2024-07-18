package routing_settings

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

var internalProxy *routingSettingsProxy

type getRoutingSettingsFunc func(ctx context.Context, p *routingSettingsProxy) (*platformclientv2.Routingsettings, *platformclientv2.APIResponse, error)
type updateRoutingSettingsFunc func(ctx context.Context, p *routingSettingsProxy, routingSettings *platformclientv2.Routingsettings) (*platformclientv2.Routingsettings, *platformclientv2.APIResponse, error)
type deleteRoutingSettingsFunc func(ctx context.Context, p *routingSettingsProxy) (*platformclientv2.APIResponse, error)

type getRoutingSettingsContactCenterFunc func(ctx context.Context, p *routingSettingsProxy) (*platformclientv2.Contactcentersettings, *platformclientv2.APIResponse, error)
type updateRoutingSettingsContactCenterFunc func(ctx context.Context, p *routingSettingsProxy, contactCenterSettings platformclientv2.Contactcentersettings) (*platformclientv2.APIResponse, error)

type getRoutingSettingsTranscriptionFunc func(ctx context.Context, p *routingSettingsProxy) (*platformclientv2.Transcriptionsettings, *platformclientv2.APIResponse, error)
type updateRoutingSettingsTranscriptionFunc func(ctx context.Context, p *routingSettingsProxy, transcriptionSettings platformclientv2.Transcriptionsettings) (*platformclientv2.Transcriptionsettings, *platformclientv2.APIResponse, error)

type routingSettingsProxy struct {
	clientConfig              *platformclientv2.Configuration
	routingSettingsApi        *platformclientv2.RoutingApi
	getRoutingSettingsAttr    getRoutingSettingsFunc
	updateRoutingSettingsAttr updateRoutingSettingsFunc
	deleteRoutingSettingsAttr deleteRoutingSettingsFunc

	getRoutingSettingsContactCenterAttr    getRoutingSettingsContactCenterFunc
	updateRoutingSettingsContactCenterAttr updateRoutingSettingsContactCenterFunc

	getRoutingSettingsTranscriptionAttr    getRoutingSettingsTranscriptionFunc
	updateRoutingSettingsTranscriptionAttr updateRoutingSettingsTranscriptionFunc
}

func newRoutingSettingsProxy(clientConfig *platformclientv2.Configuration) *routingSettingsProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingSettingsProxy{
		clientConfig:                           clientConfig,
		routingSettingsApi:                     api,
		getRoutingSettingsAttr:                 getRoutingSettingsFn,
		updateRoutingSettingsAttr:              updateRoutingSettingsFn,
		deleteRoutingSettingsAttr:              deleteRoutingSettingsFn,
		getRoutingSettingsContactCenterAttr:    getRoutingSettingsContactCenterFn,
		updateRoutingSettingsContactCenterAttr: updateRoutingSettingsContactCenterFn,
		getRoutingSettingsTranscriptionAttr:    getRoutingSettingsTranscriptionFn,
		updateRoutingSettingsTranscriptionAttr: updateRoutingSettingsTranscriptionFn,
	}
}

func getRoutingSettingsProxy(clientConfig *platformclientv2.Configuration) *routingSettingsProxy {
	if internalProxy == nil {
		internalProxy = newRoutingSettingsProxy(clientConfig)
	}
	return internalProxy
}

func (p *routingSettingsProxy) getRoutingSettings(ctx context.Context) (*platformclientv2.Routingsettings, *platformclientv2.APIResponse, error) {
	return p.getRoutingSettingsAttr(ctx, p)
}

func (p *routingSettingsProxy) updateRoutingSettings(ctx context.Context, routingSettings *platformclientv2.Routingsettings) (*platformclientv2.Routingsettings, *platformclientv2.APIResponse, error) {
	return p.updateRoutingSettingsAttr(ctx, p, routingSettings)
}

func (p *routingSettingsProxy) deleteRoutingSettings(ctx context.Context) (*platformclientv2.APIResponse, error) {
	return p.deleteRoutingSettingsAttr(ctx, p)
}

func (p *routingSettingsProxy) getRoutingSettingsContactCenter(ctx context.Context) (*platformclientv2.Contactcentersettings, *platformclientv2.APIResponse, error) {
	return p.getRoutingSettingsContactCenterAttr(ctx, p)
}

func (p *routingSettingsProxy) updateRoutingSettingsContactCenter(ctx context.Context, contactCenterSettings platformclientv2.Contactcentersettings) (*platformclientv2.APIResponse, error) {
	return p.updateRoutingSettingsContactCenterAttr(ctx, p, contactCenterSettings)
}

func (p *routingSettingsProxy) getRoutingSettingsTranscription(ctx context.Context) (*platformclientv2.Transcriptionsettings, *platformclientv2.APIResponse, error) {
	return p.getRoutingSettingsTranscriptionAttr(ctx, p)
}

func (p *routingSettingsProxy) updateRoutingSettingsTranscription(ctx context.Context, transcriptionSettings platformclientv2.Transcriptionsettings) (*platformclientv2.Transcriptionsettings, *platformclientv2.APIResponse, error) {
	return p.updateRoutingSettingsTranscriptionAttr(ctx, p, transcriptionSettings)
}

func getRoutingSettingsFn(ctx context.Context, p *routingSettingsProxy) (*platformclientv2.Routingsettings, *platformclientv2.APIResponse, error) {
	return p.routingSettingsApi.GetRoutingSettings()
}

func updateRoutingSettingsFn(ctx context.Context, p *routingSettingsProxy, routingSettings *platformclientv2.Routingsettings) (*platformclientv2.Routingsettings, *platformclientv2.APIResponse, error) {
	return p.routingSettingsApi.PutRoutingSettings(*routingSettings)
}

func deleteRoutingSettingsFn(ctx context.Context, p *routingSettingsProxy) (*platformclientv2.APIResponse, error) {
	return p.routingSettingsApi.DeleteRoutingSettings()
}

func getRoutingSettingsContactCenterFn(ctx context.Context, p *routingSettingsProxy) (*platformclientv2.Contactcentersettings, *platformclientv2.APIResponse, error) {
	return p.routingSettingsApi.GetRoutingSettingsContactcenter()
}

func updateRoutingSettingsContactCenterFn(ctx context.Context, p *routingSettingsProxy, contactCenterSettings platformclientv2.Contactcentersettings) (*platformclientv2.APIResponse, error) {
	return p.routingSettingsApi.PatchRoutingSettingsContactcenter(contactCenterSettings)
}

func getRoutingSettingsTranscriptionFn(ctx context.Context, p *routingSettingsProxy) (*platformclientv2.Transcriptionsettings, *platformclientv2.APIResponse, error) {
	return p.routingSettingsApi.GetRoutingSettingsTranscription()
}

func updateRoutingSettingsTranscriptionFn(ctx context.Context, p *routingSettingsProxy, transcriptionSettings platformclientv2.Transcriptionsettings) (*platformclientv2.Transcriptionsettings, *platformclientv2.APIResponse, error) {
	return p.routingSettingsApi.PutRoutingSettingsTranscription(transcriptionSettings)
}
