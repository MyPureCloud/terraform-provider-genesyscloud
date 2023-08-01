package outbound_wrapupcode_mappings

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

var internalProxy *outboundWrapupCodeMappingsProxy

type getAllOutboundWrapupCodeMappingsFunc func(ctx context.Context, p *outboundWrapupCodeMappingsProxy) (wrapupcodeMappings *platformclientv2.Wrapupcodemapping, resp *platformclientv2.APIResponse, err error)
type updateOutboundWrapUpCodeMappingsFunc func(ctx context.Context, p *outboundWrapupCodeMappingsProxy, outBoundWrappingCodes *platformclientv2.Wrapupcodemapping) (updatedWrapupCodeMappings *platformclientv2.Wrapupcodemapping, resp *platformclientv2.APIResponse, err error)

type outboundWrapupCodeMappingsProxy struct {
	clientConfig                         *platformclientv2.Configuration
	outboundApi                          *platformclientv2.OutboundApi
	getAllOutboundWrapupCodeMappingsAttr getAllOutboundWrapupCodeMappingsFunc
	updateOutboundWrapUpCodeMappingsAttr updateOutboundWrapUpCodeMappingsFunc
}

func newOutboundWrapupCodeMappingsProxy(clientConfig *platformclientv2.Configuration) *outboundWrapupCodeMappingsProxy {
	api := platformclientv2.NewOutboundApiWithConfig(clientConfig)
	return &outboundWrapupCodeMappingsProxy{
		clientConfig:                         clientConfig,
		outboundApi:                          api,
		getAllOutboundWrapupCodeMappingsAttr: getAllOutboundWrapupCodeMappingsFn,
		updateOutboundWrapUpCodeMappingsAttr: updateOutboundWrapUpCodeMappingsFn,
		// createExternalContactAttr:        createExternalContactFn,
		// getExternalContactByIdAttr:       getExternalContactByIdFn,
		// deleteExternalContactByIdAttr:    deleteExternalContactsFn,
		// getExternalContactIdBySearchAttr: getExternalContactIdBySearchFn,
		// updateExternalContactAttr:        updateExternalContactFn,
	}
}

func getOutboundWrapupCodeMappingsProxy(clientConfig *platformclientv2.Configuration) *outboundWrapupCodeMappingsProxy {
	if internalProxy == nil {
		internalProxy = newOutboundWrapupCodeMappingsProxy(clientConfig)
	}

	return internalProxy
}

func (p *outboundWrapupCodeMappingsProxy) getAllOutboundWrapupCodeMappings(ctx context.Context) (wrapupcodeMappings *platformclientv2.Wrapupcodemapping, resp *platformclientv2.APIResponse, err error) {
	return p.getAllOutboundWrapupCodeMappingsAttr(ctx, p)
}
func (p *outboundWrapupCodeMappingsProxy) updateOutboundWrapUpCodeMappings(ctx context.Context, outBoundWrapupCodes platformclientv2.Wrapupcodemapping) (updatedWrapupCodeMappings *platformclientv2.Wrapupcodemapping, response *platformclientv2.APIResponse, err error) {
	return p.updateOutboundWrapUpCodeMappingsAttr(ctx, p, &outBoundWrapupCodes)
}

func getAllOutboundWrapupCodeMappingsFn(ctx context.Context, p *outboundWrapupCodeMappingsProxy) (wrapupcodeMappings *platformclientv2.Wrapupcodemapping, resp *platformclientv2.APIResponse, err error) {
	wrapupcodemappings, resp, err := p.outboundApi.GetOutboundWrapupcodemappings()
	return wrapupcodemappings, resp, err
}

func updateOutboundWrapUpCodeMappingsFn(ctx context.Context, p *outboundWrapupCodeMappingsProxy, outBoundWrapupCodes *platformclientv2.Wrapupcodemapping) (updatedWrapupCodeMappings *platformclientv2.Wrapupcodemapping, resp *platformclientv2.APIResponse, err error) {
	w, resp, err := p.outboundApi.PutOutboundWrapupcodemappings(*outBoundWrapupCodes)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update wrap-up code mappings: %s", err)
	}
	return w, resp, nil
}
