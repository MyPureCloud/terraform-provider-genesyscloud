package routing_queue_conditional_group_activation

import (
	"context"
	"fmt"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	rc "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_cache"
	routingQueue "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/routing_queue"

	"github.com/mypurecloud/platform-client-sdk-go/v179/platformclientv2"
)

var internalProxy *routingQueueConditionalGroupActivationProxy

type getRoutingQueueConditionActivationFunc func(ctx context.Context, p *routingQueueConditionalGroupActivationProxy, queueId string) (*platformclientv2.Conditionalgroupactivation, *platformclientv2.APIResponse, error)
type updateRoutingQueueConditionActivationFunc func(ctx context.Context, p *routingQueueConditionalGroupActivationProxy, queueId string, cga *platformclientv2.Conditionalgroupactivation) (*platformclientv2.Conditionalgroupactivation, *platformclientv2.APIResponse, error)
type getRoutingQueueByIdFunc func(ctx context.Context, p *routingQueueConditionalGroupActivationProxy, id string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error)

type routingQueueConditionalGroupActivationProxy struct {
	clientConfig                              *platformclientv2.Configuration
	routingApi                                *platformclientv2.RoutingApi
	getRoutingQueueConditionActivationAttr    getRoutingQueueConditionActivationFunc
	getRoutingQueueByIdAttr                   getRoutingQueueByIdFunc
	updateRoutingQueueConditionActivationAttr updateRoutingQueueConditionActivationFunc
	routingQueueProxy                         *routingQueue.RoutingQueueProxy
}

func newRoutingQueueConditionalGroupActivationProxy(clientConfig *platformclientv2.Configuration) *routingQueueConditionalGroupActivationProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	routingQueueProxy := routingQueue.GetRoutingQueueProxy(clientConfig)

	return &routingQueueConditionalGroupActivationProxy{
		clientConfig:                              clientConfig,
		routingApi:                                api,
		getRoutingQueueConditionActivationAttr:    getRoutingQueueConditionActivationFn,
		updateRoutingQueueConditionActivationAttr: updateRoutingQueueConditionActivationFn,
		getRoutingQueueByIdAttr:                   getRoutingQueueByIdFn,
		routingQueueProxy:                         routingQueueProxy,
	}
}

func getRoutingQueueConditionalGroupActivationProxy(clientConfig *platformclientv2.Configuration) *routingQueueConditionalGroupActivationProxy {
	if internalProxy == nil {
		internalProxy = newRoutingQueueConditionalGroupActivationProxy(clientConfig)
	}
	return internalProxy
}

func (p *routingQueueConditionalGroupActivationProxy) getRoutingQueueById(ctx context.Context, id string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueByIdAttr(ctx, p, id)
}

func (p *routingQueueConditionalGroupActivationProxy) getRoutingQueueConditionActivation(ctx context.Context, queueId string) (*platformclientv2.Conditionalgroupactivation, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueConditionActivationAttr(ctx, p, queueId)
}

func (p *routingQueueConditionalGroupActivationProxy) updateRoutingQueueConditionActivation(ctx context.Context, queueId string, cga *platformclientv2.Conditionalgroupactivation) (*platformclientv2.Conditionalgroupactivation, *platformclientv2.APIResponse, error) {
	return p.updateRoutingQueueConditionActivationAttr(ctx, p, queueId, cga)
}

func getRoutingQueueConditionActivationFn(ctx context.Context, p *routingQueueConditionalGroupActivationProxy, queueId string) (*platformclientv2.Conditionalgroupactivation, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	var (
		queue *platformclientv2.Queue
		resp  *platformclientv2.APIResponse
		err   error
	)

	queue = rc.GetCacheItem(p.routingQueueProxy.RoutingQueueCache, queueId)
	if queue == nil {
		queue, resp, err = p.getRoutingQueueById(ctx, queueId)
		if err != nil {
			return nil, resp, fmt.Errorf("error when reading queue %s: %s", queueId, err)
		}
	}

	if queue.ConditionalGroupActivation != nil {
		return queue.ConditionalGroupActivation, resp, nil
	}

	return nil, resp, nil
}

func updateRoutingQueueConditionActivationFn(ctx context.Context, p *routingQueueConditionalGroupActivationProxy, queueId string, cga *platformclientv2.Conditionalgroupactivation) (*platformclientv2.Conditionalgroupactivation, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)

	queue, resp, err := p.getRoutingQueueById(ctx, queueId)
	if err != nil {
		return nil, resp, fmt.Errorf("error when reading queue %s: %s", queueId, err)
	}

	updateQueue := platformclientv2.Queuerequest{
		Name:                         queue.Name,
		Description:                  queue.Description,
		MemberCount:                  queue.MemberCount,
		UserMemberCount:              queue.UserMemberCount,
		JoinedMemberCount:            queue.JoinedMemberCount,
		MediaSettings:                queue.MediaSettings,
		RoutingRules:                 queue.RoutingRules,
		ConditionalGroupRouting:      queue.ConditionalGroupRouting,
		ConditionalGroupActivation:   cga,
		Bullseye:                     queue.Bullseye,
		ScoringMethod:                queue.ScoringMethod,
		AcwSettings:                  queue.AcwSettings,
		SkillEvaluationMethod:        queue.SkillEvaluationMethod,
		MemberGroups:                 queue.MemberGroups,
		QueueFlow:                    queue.QueueFlow,
		EmailInQueueFlow:             queue.EmailInQueueFlow,
		MessageInQueueFlow:           queue.MessageInQueueFlow,
		WhisperPrompt:                queue.WhisperPrompt,
		OnHoldPrompt:                 queue.OnHoldPrompt,
		AutoAnswerOnly:               queue.AutoAnswerOnly,
		EnableTranscription:          queue.EnableTranscription,
		EnableAudioMonitoring:        queue.EnableAudioMonitoring,
		EnableManualAssignment:       queue.EnableManualAssignment,
		AgentOwnedRouting:            queue.AgentOwnedRouting,
		DirectRouting:                queue.DirectRouting,
		CallingPartyName:             queue.CallingPartyName,
		CallingPartyNumber:           queue.CallingPartyNumber,
		DefaultScripts:               queue.DefaultScripts,
		OutboundMessagingAddresses:   queue.OutboundMessagingAddresses,
		PeerId:                       queue.PeerId,
		SuppressInQueueCallRecording: queue.SuppressInQueueCallRecording,
	}

	// OutboundEmailAddress returned by GetRoutingQueue is a pointer to a pointer
	if queue.OutboundEmailAddress != nil && *queue.OutboundEmailAddress != nil {
		updateQueue.OutboundEmailAddress = *queue.OutboundEmailAddress
	}

	// Remove queue_id from first CGR rule to avoid API error
	if updateQueue.ConditionalGroupRouting != nil && updateQueue.ConditionalGroupRouting.Rules != nil && len(*updateQueue.ConditionalGroupRouting.Rules) > 0 {
		(*updateQueue.ConditionalGroupRouting.Rules)[0].Queue = nil
	}

	queue, resp, err = p.routingApi.PutRoutingQueue(queueId, updateQueue)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update conditional group activation for routing queue %s: %s", queueId, err)
	}

	if queue.ConditionalGroupActivation != nil {
		return queue.ConditionalGroupActivation, resp, nil
	}

	return nil, resp, nil
}

func getRoutingQueueByIdFn(ctx context.Context, p *routingQueueConditionalGroupActivationProxy, id string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	ctx = provider.EnsureResourceContext(ctx, ResourceType)
	return p.routingApi.GetRoutingQueue(id, nil)
}
