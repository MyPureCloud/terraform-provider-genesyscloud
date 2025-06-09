package guide_jobs

import (
	"context"
	"github.com/mypurecloud/platform-client-sdk-go/v157/platformclientv2"
)

var internalProxy *guideJobsProxy

type createGuideJobFunc func(ctx context.Context, p *guideJobsProxy, guideJob *GenerateGuideContentRequest) (*GuideJob, *platformclientv2.APIResponse, error)
type getGuideJobByIdFunc func(ctx context.Context, p *guideJobsProxy, id string) (guideJob *GuideJob, resp *platformclientv2.APIResponse, err error)
type deleteGuideJobFunc func(ctx context.Context, p *guideJobsProxy, id string) (resp *platformclientv2.APIResponse, err error)

type guideJobsProxy struct {
	clientConfig *platformclientv2.Configuration
	// TODO: implement api
	createGuideJobAttr  createGuideJobFunc
	getGuideJobByIdAttr getGuideJobByIdFunc
	deleteGuideJobAttr  deleteGuideJobFunc
}

func newGuideJobsProxy(clientConfig *platformclientv2.Configuration) *guideJobsProxy {
	// TODO: implement api
	return &guideJobsProxy{
		clientConfig:        clientConfig,
		createGuideJobAttr:  createGuideJobFn,
		getGuideJobByIdAttr: getGuideJobByIdFn,
		deleteGuideJobAttr:  deleteGuideJobFn,
	}
}

func getGuideJobsProxy(config *platformclientv2.Configuration) *guideJobsProxy {
	if internalProxy == nil {
		internalProxy = newGuideJobsProxy(config)
	}
	return internalProxy
}

func (p *guideJobsProxy) createGuideJob(ctx context.Context, guideJob *GenerateGuideContentRequest) (*GuideJob, *platformclientv2.APIResponse, error) {
	return p.createGuideJobAttr(ctx, p, guideJob)
}

func (p *guideJobsProxy) getGuideJobById(ctx context.Context, id string) (guideJob *GuideJob, resp *platformclientv2.APIResponse, err error) {
	return p.getGuideJobByIdAttr(ctx, p, id)
}

func (p *guideJobsProxy) deleteGuideJob(ctx context.Context, id string) (resp *platformclientv2.APIResponse, err error) {
	return p.deleteGuideJobAttr(ctx, p, id)
}

// Create Functions

func createGuideJobFn(ctx context.Context, p *guideJobsProxy, guideJob *GenerateGuideContentRequest) (*GuideJob, *platformclientv2.APIResponse, error) {
	return sdkCreateGuideJob(ctx, p, guideJob)
}

func sdkCreateGuideJob(ctx context.Context, p *guideJobsProxy, guideJob *GenerateGuideContentRequest) (*GuideJob, *platformclientv2.APIResponse, error) {

}

// Read Functions

func getGuideJobByIdFn(ctx context.Context, p *guideJobsProxy, id string) (guideJob *GuideJob, resp *platformclientv2.APIResponse, err error) {
	return sdkGetGuideJobById(ctx, p, id)
}

func sdkGetGuideJobById(ctx context.Context, p *guideJobsProxy, id string) (guideJob *GuideJob, resp *platformclientv2.APIResponse, err error) {

}

// Delete Functions

func deleteGuideJobFn(ctx context.Context, p *guideJobsProxy, id string) (resp *platformclientv2.APIResponse, err error) {
	return sdkDeleteGuideJob(ctx, p, id)
}

func sdkDeleteGuideJob(ctx context.Context, p *guideJobsProxy, id string) (resp *platformclientv2.APIResponse, err error) {

}
