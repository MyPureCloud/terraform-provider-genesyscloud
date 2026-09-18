package architect_ivr_identity_resolution

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	architectIvr "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/architect_ivr"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

/*
The genesyscloud_architect_ivr_identity_resolution_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

var internalProxy *architectIvrIdentityResolutionProxy

type getArchitectIvrIdentityResolutionFunc func(ctx context.Context, p *architectIvrIdentityResolutionProxy, ivrId string) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error)
type putArchitectIvrIdentityResolutionFunc func(ctx context.Context, p *architectIvrIdentityResolutionProxy, ivrId string, config platformclientv2.Ivridentityresolutionconfig) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error)

type architectIvrIdentityResolutionProxy struct {
	clientConfig                          *platformclientv2.Configuration
	architectApi                          *platformclientv2.ArchitectApi
	getArchitectIvrIdentityResolutionAttr getArchitectIvrIdentityResolutionFunc
	putArchitectIvrIdentityResolutionAttr putArchitectIvrIdentityResolutionFunc
	architectIvrProxy                     *architectIvr.ArchitectIvrProxy
}

func newArchitectIvrIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *architectIvrIdentityResolutionProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	return &architectIvrIdentityResolutionProxy{
		clientConfig:                          clientConfig,
		architectApi:                          api,
		getArchitectIvrIdentityResolutionAttr: getArchitectIvrIdentityResolutionFn,
		putArchitectIvrIdentityResolutionAttr: putArchitectIvrIdentityResolutionFn,
		architectIvrProxy:                     architectIvr.GetArchitectIvrProxy(clientConfig),
	}
}

func getArchitectIvrIdentityResolutionProxy(clientConfig *platformclientv2.Configuration) *architectIvrIdentityResolutionProxy {
	if internalProxy == nil {
		internalProxy = newArchitectIvrIdentityResolutionProxy(clientConfig)
	}
	return internalProxy
}

func (p *architectIvrIdentityResolutionProxy) getArchitectIvrIdentityResolution(ctx context.Context, ivrId string) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error) {
	return p.getArchitectIvrIdentityResolutionAttr(ctx, p, ivrId)
}

func (p *architectIvrIdentityResolutionProxy) putArchitectIvrIdentityResolution(ctx context.Context, ivrId string, config platformclientv2.Ivridentityresolutionconfig) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error) {
	return p.putArchitectIvrIdentityResolutionAttr(ctx, p, ivrId, config)
}

func getArchitectIvrIdentityResolutionFn(ctx context.Context, p *architectIvrIdentityResolutionProxy, ivrId string) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.architectApi.GetArchitectIvrIdentityresolution(ivrId)
}

func putArchitectIvrIdentityResolutionFn(ctx context.Context, p *architectIvrIdentityResolutionProxy, ivrId string, config platformclientv2.Ivridentityresolutionconfig) (*platformclientv2.Ivridentityresolutionconfig, *platformclientv2.APIResponse, error) {
	// Set resource context for SDK debug logging
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.architectApi.PutArchitectIvrIdentityresolution(ivrId, config)
}