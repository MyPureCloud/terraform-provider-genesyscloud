package speechandtextanalytics_topic

import (
	"context"
	"fmt"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
)

var internalProxy *sttTopicProxy

type (
	createTopicFunc   func(ctx context.Context, p *sttTopicProxy, body *platformclientv2.Topicrequest) (*platformclientv2.Topic, *platformclientv2.APIResponse, error)
	getTopicFunc      func(ctx context.Context, p *sttTopicProxy, id string) (*platformclientv2.Topic, *platformclientv2.APIResponse, error)
	updateTopicFunc   func(ctx context.Context, p *sttTopicProxy, id string, body *platformclientv2.Topicrequest) (*platformclientv2.Topic, *platformclientv2.APIResponse, error)
	deleteTopicFunc   func(ctx context.Context, p *sttTopicProxy, id string) (*platformclientv2.APIResponse, error)
	listTopicsFunc    func(ctx context.Context, p *sttTopicProxy, pageSize int, pageNumber int) (*platformclientv2.Topicsentitylisting, *platformclientv2.APIResponse, error)
	publishTopicsFunc func(ctx context.Context, p *sttTopicProxy, topicIds []string) (*platformclientv2.Topicjob, *platformclientv2.APIResponse, error)
	getPublishJobFunc func(ctx context.Context, p *sttTopicProxy, jobId string) (*platformclientv2.Topicjob, *platformclientv2.APIResponse, error)
)

type sttTopicProxy struct {
	clientConfig      *platformclientv2.Configuration
	sttApi            *platformclientv2.SpeechTextAnalyticsApi
	createTopicAttr   createTopicFunc
	getTopicAttr      getTopicFunc
	updateTopicAttr   updateTopicFunc
	deleteTopicAttr   deleteTopicFunc
	listTopicsAttr    listTopicsFunc
	publishTopicsAttr publishTopicsFunc
	getPublishJobAttr getPublishJobFunc
}

func newSttTopicProxy(clientConfig *platformclientv2.Configuration) *sttTopicProxy {
	api := platformclientv2.NewSpeechTextAnalyticsApiWithConfig(clientConfig)
	return &sttTopicProxy{
		clientConfig:      clientConfig,
		sttApi:            api,
		createTopicAttr:   createTopicFn,
		getTopicAttr:      getTopicFn,
		updateTopicAttr:   updateTopicFn,
		deleteTopicAttr:   deleteTopicFn,
		listTopicsAttr:    listTopicsFn,
		publishTopicsAttr: publishTopicsFn,
		getPublishJobAttr: getPublishJobFn,
	}
}

func getSttTopicProxy(clientConfig *platformclientv2.Configuration) *sttTopicProxy {
	if internalProxy == nil {
		internalProxy = newSttTopicProxy(clientConfig)
	}
	return internalProxy
}

func (p *sttTopicProxy) createTopic(ctx context.Context, body *platformclientv2.Topicrequest) (*platformclientv2.Topic, *platformclientv2.APIResponse, error) {
	return p.createTopicAttr(ctx, p, body)
}

func (p *sttTopicProxy) getTopic(ctx context.Context, id string) (*platformclientv2.Topic, *platformclientv2.APIResponse, error) {
	return p.getTopicAttr(ctx, p, id)
}

func (p *sttTopicProxy) updateTopic(ctx context.Context, id string, body *platformclientv2.Topicrequest) (*platformclientv2.Topic, *platformclientv2.APIResponse, error) {
	return p.updateTopicAttr(ctx, p, id, body)
}

func (p *sttTopicProxy) deleteTopic(ctx context.Context, id string) (*platformclientv2.APIResponse, error) {
	return p.deleteTopicAttr(ctx, p, id)
}

func (p *sttTopicProxy) listTopics(ctx context.Context, pageSize int, pageNumber int) (*platformclientv2.Topicsentitylisting, *platformclientv2.APIResponse, error) {
	return p.listTopicsAttr(ctx, p, pageSize, pageNumber)
}

func (p *sttTopicProxy) publishTopics(ctx context.Context, topicIds []string) (*platformclientv2.Topicjob, *platformclientv2.APIResponse, error) {
	return p.publishTopicsAttr(ctx, p, topicIds)
}

func (p *sttTopicProxy) getPublishJob(ctx context.Context, jobId string) (*platformclientv2.Topicjob, *platformclientv2.APIResponse, error) {
	return p.getPublishJobAttr(ctx, p, jobId)
}

func createTopicFn(ctx context.Context, p *sttTopicProxy, body *platformclientv2.Topicrequest) (*platformclientv2.Topic, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	topic, resp, err := p.sttApi.PostSpeechandtextanalyticsTopics(*body) // POST /api/v2/speechandtextanalytics/topics
	if err != nil {
		return nil, resp, fmt.Errorf("failed to create speech and text analytics topic: %s", err)
	}
	return topic, resp, nil
}

func getTopicFn(ctx context.Context, p *sttTopicProxy, id string) (*platformclientv2.Topic, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	topic, resp, err := p.sttApi.GetSpeechandtextanalyticsTopic(id) // /api/v2/speechandtextanalytics/topics/{topicId}
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get speech and text analytics topic %s: %s", id, err)
	}
	return topic, resp, nil
}

func updateTopicFn(ctx context.Context, p *sttTopicProxy, id string, body *platformclientv2.Topicrequest) (*platformclientv2.Topic, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	topic, resp, err := p.sttApi.PutSpeechandtextanalyticsTopic(id, *body) // /api/v2/speechandtextanalytics/topics/{topicId}
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update speech and text analytics topic %s: %s", id, err)
	}
	return topic, resp, nil
}

func deleteTopicFn(ctx context.Context, p *sttTopicProxy, id string) (*platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	resp, err := p.sttApi.DeleteSpeechandtextanalyticsTopic(id) // /api/v2/speechandtextanalytics/topics/{topicId}
	if err != nil {
		return resp, fmt.Errorf("failed to delete speech and text analytics topic %s: %s", id, err)
	}
	return resp, nil
}

func listTopicsFn(ctx context.Context, p *sttTopicProxy, pageSize int, pageNumber int) (*platformclientv2.Topicsentitylisting, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	listing, resp, err := p.sttApi.GetSpeechandtextanalyticsTopics("", pageSize, pageNumber, "", "", nil, nil, "", "") // GET /api/v2/speechandtextanalytics/topics
	if err != nil {
		return nil, resp, fmt.Errorf("failed to list speech and text analytics topics: %s", err)
	}
	return listing, resp, nil
}

func publishTopicsFn(ctx context.Context, p *sttTopicProxy, topicIds []string) (*platformclientv2.Topicjob, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	job, resp, err := p.sttApi.PostSpeechandtextanalyticsTopicsPublishjobs(platformclientv2.Topicjobrequest{
		TopicIds: &topicIds,
	})
	if err != nil {
		return nil, resp, fmt.Errorf("failed to publish topics: %s", err)
	}
	return job, resp, nil
}

func getPublishJobFn(ctx context.Context, p *sttTopicProxy, jobId string) (*platformclientv2.Topicjob, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	job, resp, err := p.sttApi.GetSpeechandtextanalyticsTopicsPublishjob(jobId)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get topics publish job %s: %s", jobId, err)
	}
	return job, resp, nil
}
