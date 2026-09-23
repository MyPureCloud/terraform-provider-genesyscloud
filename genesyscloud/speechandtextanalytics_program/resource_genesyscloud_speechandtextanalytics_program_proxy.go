package speechandtextanalytics_program

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v199/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

var internalProxy *sttProgramProxy

type (
	createProgramFunc func(ctx context.Context, p *sttProgramProxy, body *platformclientv2.Programrequest) (*platformclientv2.Program, *platformclientv2.APIResponse, error)
	getProgramFunc    func(ctx context.Context, p *sttProgramProxy, id string) (*platformclientv2.Program, *platformclientv2.APIResponse, error)
	updateProgramFunc func(ctx context.Context, p *sttProgramProxy, id string, body *platformclientv2.Programrequest) (*platformclientv2.Program, *platformclientv2.APIResponse, error)
	deleteProgramFunc func(ctx context.Context, p *sttProgramProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error)
	listProgramsFunc  func(ctx context.Context, p *sttProgramProxy, nextPage string, pageSize int) (*platformclientv2.Programsentitylisting, *platformclientv2.APIResponse, error)
)

type sttProgramProxy struct {
	clientConfig      *platformclientv2.Configuration
	sttApi            *platformclientv2.SpeechTextAnalyticsApi
	createProgramAttr createProgramFunc
	getProgramAttr    getProgramFunc
	updateProgramAttr updateProgramFunc
	deleteProgramAttr deleteProgramFunc
	listProgramsAttr  listProgramsFunc
}

func newSttProgramProxy(clientConfig *platformclientv2.Configuration) *sttProgramProxy {
	api := platformclientv2.NewSpeechTextAnalyticsApiWithConfig(clientConfig)
	return &sttProgramProxy{
		clientConfig:      clientConfig,
		sttApi:            api,
		createProgramAttr: createProgramFn,
		getProgramAttr:    getProgramFn,
		updateProgramAttr: updateProgramFn,
		deleteProgramAttr: deleteProgramFn,
		listProgramsAttr:  listProgramsFn,
	}
}

func getSttProgramProxy(clientConfig *platformclientv2.Configuration) *sttProgramProxy {
	if internalProxy == nil {
		internalProxy = newSttProgramProxy(clientConfig)
	}
	return internalProxy
}

func (p *sttProgramProxy) createProgram(ctx context.Context, body *platformclientv2.Programrequest) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
	return p.createProgramAttr(ctx, p, body)
}

func (p *sttProgramProxy) getProgram(ctx context.Context, id string) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
	return p.getProgramAttr(ctx, p, id)
}

func (p *sttProgramProxy) updateProgram(ctx context.Context, id string, body *platformclientv2.Programrequest) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
	return p.updateProgramAttr(ctx, p, id, body)
}

func (p *sttProgramProxy) deleteProgram(ctx context.Context, id string, forceDelete bool) (*platformclientv2.APIResponse, error) {
	return p.deleteProgramAttr(ctx, p, id, forceDelete)
}

func (p *sttProgramProxy) listPrograms(ctx context.Context, nextPage string, pageSize int) (*platformclientv2.Programsentitylisting, *platformclientv2.APIResponse, error) {
	return p.listProgramsAttr(ctx, p, nextPage, pageSize)
}

func createProgramFn(ctx context.Context, p *sttProgramProxy, body *platformclientv2.Programrequest) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	program, resp, err := p.sttApi.PostSpeechandtextanalyticsPrograms(*body) // POST /api/v2/speechandtextanalytics/programs
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create speech and text analytics program: %s", err)
	}
	return program, resp, nil
}

func getProgramFn(ctx context.Context, p *sttProgramProxy, id string) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	program, resp, err := p.sttApi.GetSpeechandtextanalyticsProgram(id) // GET /api/v2/speechandtextanalytics/programs/{programId}
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get speech and text analytics program %s: %s", id, err)
	}
	return program, resp, nil
}

func updateProgramFn(ctx context.Context, p *sttProgramProxy, id string, body *platformclientv2.Programrequest) (*platformclientv2.Program, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	program, resp, err := p.sttApi.PutSpeechandtextanalyticsProgram(id, *body) // PUT /api/v2/speechandtextanalytics/programs/{programId}
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update speech and text analytics program %s: %s", id, err)
	}
	return program, resp, nil
}

func deleteProgramFn(ctx context.Context, p *sttProgramProxy, id string, forceDelete bool) (*platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	// forceDelete=true so destroy is not blocked when the program is still referenced.
	_, resp, err := p.sttApi.DeleteSpeechandtextanalyticsProgram(id, forceDelete) // DELETE /api/v2/speechandtextanalytics/programs/{programId}
	if err != nil {
		return resp, fmt.Errorf("failed to delete speech and text analytics program %s: %s", id, err)
	}
	return resp, nil
}

func listProgramsFn(ctx context.Context, p *sttProgramProxy, nextPage string, pageSize int) (*platformclientv2.Programsentitylisting, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	listing, resp, err := p.sttApi.GetSpeechandtextanalyticsPrograms(nextPage, pageSize, "", "", nil, "", "") // GET /api/v2/speechandtextanalytics/programs
	if err != nil {
		return nil, resp, fmt.Errorf("failed to list speech and text analytics programs: %s", err)
	}
	return listing, resp, nil
}
