package architect_flow

import (
	"context"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

// The validate job API calls live in their own file so the apidocs audit attributes
// them only to the genesyscloud_flow_validation data source.

// CreateFlowValidateJob registers an architect validate job for a single flow.
// Implementation function: createFlowValidateJobFn
func (a *architectFlowProxy) CreateFlowValidateJob(ctx context.Context, flow platformclientv2.Architectflowreference) (*platformclientv2.Registerarchitectvalidatejobresponse, *platformclientv2.APIResponse, error) {
	return a.createFlowValidateJobAttr(ctx, a, flow)
}

// GetFlowValidateJob fetches the state of an architect validate job, including its messages.
// Implementation function: getFlowValidateJobFn
func (a *architectFlowProxy) GetFlowValidateJob(ctx context.Context, jobId string) (*platformclientv2.Architectvalidatejobstateresponse, *platformclientv2.APIResponse, error) {
	return a.getFlowValidateJobAttr(ctx, a, jobId)
}

func createFlowValidateJobFn(ctx context.Context, p *architectFlowProxy, flow platformclientv2.Architectflowreference) (*platformclientv2.Registerarchitectvalidatejobresponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ValidationDataSourceType)
	body := platformclientv2.Registerarchitectvalidatejob{
		Flows: &[]platformclientv2.Validatedetails{{Flow: &flow}},
	}
	return p.api.PostFlowsValidateJobs(body)
}

func getFlowValidateJobFn(ctx context.Context, p *architectFlowProxy, jobId string) (*platformclientv2.Architectvalidatejobstateresponse, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ValidationDataSourceType)
	return p.api.GetFlowsValidateJob(jobId, []string{"messages"})
}
