package architect_flow

import (
	"context"
	"fmt"
	"log"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var internalProxy *architectFlowProxy

type getArchitectFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.Flow, *platformclientv2.APIResponse, error)
type forceUnlockFlowFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.APIResponse, error)
type deleteArchitectFlowFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.APIResponse, error)
type createArchitectFlowJobsFunc func(context.Context, *architectFlowProxy) (*platformclientv2.Registerarchitectjobresponse, *platformclientv2.APIResponse, error)
type getArchitectFlowJobsFunc func(context.Context, *architectFlowProxy, string) (*platformclientv2.Architectjobstateresponse, *platformclientv2.APIResponse, error)
type getAllArchitectFlowsFunc func(context.Context, *architectFlowProxy, string, []string) (*[]platformclientv2.Flow, *platformclientv2.APIResponse, error)
type getFlowIdByNameAndTypeFunc func(ctx context.Context, a *architectFlowProxy, name string, varType string) (id string, resp *platformclientv2.APIResponse, retryable bool, err error)

type generateDownloadUrlFunc func(a *architectFlowProxy, flowId string) (string, error)
type createExportJobFunc func(a *architectFlowProxy, flowId string) (jobId string, err error)
type getExportJobFunc func(a *architectFlowProxy, jobId string) (*platformclientv2.Architectexportjobstateresponse, error)
type pollExportJobForDownloadUrlFunc func(a *architectFlowProxy, jobId string) (downloadUrl string, err error)

type architectFlowProxy struct {
	clientConfig *platformclientv2.Configuration
	api          *platformclientv2.ArchitectApi

	getArchitectFlowAttr            getArchitectFunc
	getAllArchitectFlowsAttr        getAllArchitectFlowsFunc
	forceUnlockFlowAttr             forceUnlockFlowFunc
	deleteArchitectFlowAttr         deleteArchitectFlowFunc
	createArchitectFlowJobsAttr     createArchitectFlowJobsFunc
	getArchitectFlowJobsAttr        getArchitectFlowJobsFunc
	getFlowIdByNameAndTypeAttr      getFlowIdByNameAndTypeFunc
	createExportJobAttr             createExportJobFunc
	getExportJobStatusAttr          getExportJobFunc
	pollExportJobForDownloadUrlAttr pollExportJobForDownloadUrlFunc
	generateDownloadUrlAttr         generateDownloadUrlFunc

	flowCache rc.CacheInterface[platformclientv2.Flow]
}

var flowCache = rc.NewResourceCache[platformclientv2.Flow]()

func newArchitectFlowProxy(clientConfig *platformclientv2.Configuration) *architectFlowProxy {
	api := platformclientv2.NewArchitectApiWithConfig(clientConfig)
	return &architectFlowProxy{
		clientConfig: clientConfig,
		api:          api,

		getArchitectFlowAttr:            getArchitectFlowFn,
		getAllArchitectFlowsAttr:        getAllArchitectFlowsFn,
		forceUnlockFlowAttr:             forceUnlockFlowFn,
		deleteArchitectFlowAttr:         deleteArchitectFlowFn,
		createArchitectFlowJobsAttr:     createArchitectFlowJobsFn,
		getArchitectFlowJobsAttr:        getArchitectFlowJobsFn,
		getFlowIdByNameAndTypeAttr:      getFlowIdByNameAndTypeFn,
		generateDownloadUrlAttr:         generateDownloadUrlFn,
		createExportJobAttr:             createExportJobFn,
		getExportJobStatusAttr:          getExportJobStatusFn,
		pollExportJobForDownloadUrlAttr: pollExportJobForDownloadUrlFn,
		flowCache:                       flowCache,
	}
}

func getArchitectFlowProxy(clientConfig *platformclientv2.Configuration) *architectFlowProxy {
	if internalProxy == nil {
		internalProxy = newArchitectFlowProxy(clientConfig)
	}
	return internalProxy
}

func (a *architectFlowProxy) GetFlow(ctx context.Context, id string) (*platformclientv2.Flow, *platformclientv2.APIResponse, error) {
	return a.getArchitectFlowAttr(ctx, a, id)
}

func (a *architectFlowProxy) ForceUnlockFlow(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return a.forceUnlockFlowAttr(ctx, a, id)
}

func (a *architectFlowProxy) DeleteFlow(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return a.deleteArchitectFlowAttr(ctx, a, id)
}

func (a *architectFlowProxy) CreateFlowsDeployJob(ctx context.Context) (*platformclientv2.Registerarchitectjobresponse, *platformclientv2.APIResponse, error) {
	return a.createArchitectFlowJobsAttr(ctx, a)
}

func (a *architectFlowProxy) GetFlowsDeployJob(ctx context.Context, jobId string) (*platformclientv2.Architectjobstateresponse, *platformclientv2.APIResponse, error) {
	return a.getArchitectFlowJobsAttr(ctx, a, jobId)
}

func (a *architectFlowProxy) GetAllFlows(ctx context.Context, name string, varType []string) (*[]platformclientv2.Flow, *platformclientv2.APIResponse, error) {
	return a.getAllArchitectFlowsAttr(ctx, a, name, varType)
}

func (a *architectFlowProxy) getFlowIdByNameAndType(ctx context.Context, name, varType string) (string, *platformclientv2.APIResponse, bool, error) {
	return a.getFlowIdByNameAndTypeAttr(ctx, a, name, varType)
}

func (a *architectFlowProxy) generateDownloadUrl(flowId string) (string, error) {
	return a.generateDownloadUrlAttr(a, flowId)
}

func (a *architectFlowProxy) createExportJob(flowId string) (string, error) {
	return a.createExportJobAttr(a, flowId)
}

func (a *architectFlowProxy) getExportJob(jobId string) (*platformclientv2.Architectexportjobstateresponse, error) {
	return a.getExportJobStatusAttr(a, jobId)
}

func (a *architectFlowProxy) pollExportJobForDownloadUrl(jobId string) (downloadUrl string, err error) {
	return a.pollExportJobForDownloadUrlAttr(a, jobId)
}

func getFlowIdByNameAndTypeFn(ctx context.Context, a *architectFlowProxy, name, varType string) (string, *platformclientv2.APIResponse, bool, error) {
	var (
		matchedFlowIds []string
		typeDetails    string
		types          []string
	)

	if varType != "" {
		types = append(types, varType)
	}

	if varType != "" {
		typeDetails = fmt.Sprintf("type '%s'", varType)
	}
	noFlowsFoundErr := fmt.Errorf("no flows found with name '%s' %s", name, typeDetails)

	flows, resp, err := a.GetAllFlows(ctx, name, types)
	if err != nil {
		return "", resp, false, err
	}

	if flows == nil || len(*flows) == 0 {
		return "", nil, true, noFlowsFoundErr
	}

	for _, flow := range *flows {
		if *flow.Name == name {
			matchedFlowIds = append(matchedFlowIds, *flow.Id)
		}
		if len(matchedFlowIds) > 1 {
			return "", nil, true, fmt.Errorf("found more than one flow that matched the name '%s'", name)
		}
	}

	if len(matchedFlowIds) == 1 {
		return matchedFlowIds[0], nil, false, nil
	}

	return "", nil, true, noFlowsFoundErr
}

func getArchitectFlowFn(_ context.Context, p *architectFlowProxy, id string) (*platformclientv2.Flow, *platformclientv2.APIResponse, error) {
	flow := rc.GetCacheItem(p.flowCache, id)
	if flow != nil {
		return flow, nil, nil
	}
	return p.api.GetFlow(id, false)
}

func forceUnlockFlowFn(_ context.Context, p *architectFlowProxy, flowId string) (*platformclientv2.APIResponse, error) {
	log.Printf("Attempting to perform an unlock on flow: %s", flowId)
	_, resp, err := p.api.PostFlowsActionsUnlock(flowId)
	return resp, err
}

func deleteArchitectFlowFn(_ context.Context, p *architectFlowProxy, flowId string) (*platformclientv2.APIResponse, error) {
	return p.api.DeleteFlow(flowId)
}

func createArchitectFlowJobsFn(_ context.Context, p *architectFlowProxy) (*platformclientv2.Registerarchitectjobresponse, *platformclientv2.APIResponse, error) {
	return p.api.PostFlowsJobs()
}

func getArchitectFlowJobsFn(_ context.Context, p *architectFlowProxy, jobId string) (*platformclientv2.Architectjobstateresponse, *platformclientv2.APIResponse, error) {
	return p.api.GetFlowsJob(jobId, []string{"messages"})
}

func getAllArchitectFlowsFn(ctx context.Context, p *architectFlowProxy, name string, varType []string) (*[]platformclientv2.Flow, *platformclientv2.APIResponse, error) {
	const pageSize = 100
	var totalFlows []platformclientv2.Flow

	flows, resp, err := p.api.GetFlows(varType, 1, pageSize, "", "", nil, name, "", "", "", "", "", "", "", false, true, "", "", nil)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get page of flows: %v %v", err, resp)
	}
	if flows.Entities == nil || len(*flows.Entities) == 0 {
		return &totalFlows, nil, nil
	}

	totalFlows = append(totalFlows, *flows.Entities...)

	for pageNum := 2; pageNum <= *flows.PageCount; pageNum++ {
		flows, resp, err := p.api.GetFlows(varType, pageNum, pageSize, "", "", nil, name, "", "", "", "", "", "", "", false, true, "", "", nil)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get page %d of flows: %v", pageNum, err)
		}
		if flows.Entities == nil || len(*flows.Entities) == 0 {
			break
		}
		totalFlows = append(totalFlows, *flows.Entities...)
	}

	for _, flow := range totalFlows {
		rc.SetCache(p.flowCache, *flow.Id, flow)
	}

	return &totalFlows, nil, nil
}

func generateDownloadUrlFn(a *architectFlowProxy, flowId string) (string, error) {
	log.Printf("Creating export job for flow %s", flowId)
	jobId, err := a.createExportJob(flowId)
	if err != nil {
		return "", err
	}
	log.Printf("Successfully created export job '%s' for flow '%s'", jobId, flowId)

	log.Printf("Polling job '%s' for download url", jobId)
	downloadUrl, err := a.pollExportJobForDownloadUrl(jobId)
	if err != nil {
		return "", err
	}
	log.Printf("Successfully read download URL. Export job: %s", jobId)

	return downloadUrl, nil
}

func createExportJobFn(a *architectFlowProxy, flowId string) (string, error) {
	body := platformclientv2.Registerarchitectexportjob{
		Flows: &[]platformclientv2.Exportdetails{
			{
				Flow: &platformclientv2.Architectflowreference{
					Id: &flowId,
				},
			},
		},
	}

	createJob, _, err := a.api.PostFlowsExportJobs(body)
	if err != nil {
		return "", fmt.Errorf("failed to create export job for flow %s: %s", flowId, err.Error())
	}

	if createJob == nil || createJob.Id == nil {
		return "", fmt.Errorf("no export job flow ID returned for flow %s", flowId)
	}

	return *createJob.Id, nil
}

func getExportJobStatusFn(a *architectFlowProxy, jobId string) (*platformclientv2.Architectexportjobstateresponse, error) {
	jobStatus, resp, err := a.api.GetFlowsExportJob(jobId, []string{"messages"})
	if err != nil {
		return nil, fmt.Errorf("failed to get job status for job %s: %s", jobId, err.Error())
	}

	if jobStatus == nil || jobStatus.Status == nil {
		return nil, fmt.Errorf("export job %s response body did not provide a status. Response: %s", jobId, resp.String())
	}

	return jobStatus, nil
}

func pollExportJobForDownloadUrlFn(a *architectFlowProxy, jobId string) (string, error) {
	for {
		log.Printf("Sleeping for 3 seconds before polling job.")
		time.Sleep(3 * time.Second)

		exportJob, err := a.getExportJob(jobId)
		if err != nil {
			return "", err
		}

		status := *exportJob.Status

		if status == "Started" {
			continue
		} else if status == "Failure" {
			return "", fmt.Errorf("job %s failed. Messages: %s", jobId, parseMessagesFromExportJobStateResponse(*exportJob))
		} else if status != "Success" {
			return "", fmt.Errorf("unexpected job status %s for job %s", status, jobId)
		}

		if exportJob.DownloadUrl == nil {
			return "", fmt.Errorf("job %s was a success but no download ID was returned", jobId)
		}

		return *exportJob.DownloadUrl, nil
	}
}

func parseMessagesFromExportJobStateResponse(job platformclientv2.Architectexportjobstateresponse) string {
	if job.Messages == nil {
		return ""
	}
	var messages string
	for _, m := range *job.Messages {
		messages += fmt.Sprintf("\n%s", m.String())
	}
	return messages
}
