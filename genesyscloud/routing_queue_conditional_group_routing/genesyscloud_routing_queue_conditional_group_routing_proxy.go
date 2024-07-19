package routing_queue_conditional_group_routing

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingQueueConditionalGroupRoutingProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getRoutingQueueConditionRoutingFunc func(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error)
type updateRoutingQueueConditionRoutingFunc func(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string, rules *[]platformclientv2.Conditionalgrouproutingrule) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error)
type getRoutingQueueByIdFunc func(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, id string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error)

// routingQueueConditionalGroupRoutingProxy contains all methods that call genesys cloud APIs.
type routingQueueConditionalGroupRoutingProxy struct {
	clientConfig                           *platformclientv2.Configuration
	routingApi                             *platformclientv2.RoutingApi
	getRoutingQueueConditionRoutingAttr    getRoutingQueueConditionRoutingFunc
	getRoutingQueueByIdAttr                getRoutingQueueByIdFunc
	updateRoutingQueueConditionRoutingAttr updateRoutingQueueConditionRoutingFunc
	routingQueueProxy                      *routingQueue.RoutingQueueProxy
}

// newRoutingQueueConditionalGroupRoutingProxy initializes the Routing queue conditional group routing proxy with all of the data needed to communicate with Genesys Cloud
func newRoutingQueueConditionalGroupRoutingProxy(clientConfig *platformclientv2.Configuration) *routingQueueConditionalGroupRoutingProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	routingQueueProxy := routingQueue.GetRoutingQueueProxy(clientConfig)

	return &routingQueueConditionalGroupRoutingProxy{
		clientConfig:                           clientConfig,
		routingApi:                             api,
		getRoutingQueueConditionRoutingAttr:    getRoutingQueueConditionRoutingFn,
		updateRoutingQueueConditionRoutingAttr: updateRoutingQueueConditionRoutingFn,
		getRoutingQueueByIdAttr:                getRoutingQueueByIdFn,
		routingQueueProxy:                      routingQueueProxy,
	}
}

// getRoutingQueueConditionalGroupRoutingProxy retrieves all Genesys Cloud Routing queue conditional group routing
func getRoutingQueueConditionalGroupRoutingProxy(clientConfig *platformclientv2.Configuration) *routingQueueConditionalGroupRoutingProxy {
	if internalProxy == nil {
		internalProxy = newRoutingQueueConditionalGroupRoutingProxy(clientConfig)
	}

	return internalProxy
}

// getRoutingQueueById get a queue by ID
func (p *routingQueueConditionalGroupRoutingProxy) getRoutingQueueById(ctx context.Context, id string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueByIdAttr(ctx, p, id)
}

// getRoutingQueueConditionRouting gets the conditional group routing rules for a queue
func (p *routingQueueConditionalGroupRoutingProxy) getRoutingQueueConditionRouting(ctx context.Context, queueId string) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueConditionRoutingAttr(ctx, p, queueId)
}

// updateRoutingQueueConditionRouting updates the conditional group routing rules for a queue
func (p *routingQueueConditionalGroupRoutingProxy) updateRoutingQueueConditionRouting(ctx context.Context, queueId string, rules *[]platformclientv2.Conditionalgrouproutingrule) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
	return p.updateRoutingQueueConditionRoutingAttr(ctx, p, queueId, rules)
}

// getRoutingQueueConditionRoutingFn is an implementation function for getting the conditional group routing rules for a queue
func getRoutingQueueConditionRoutingFn(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
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

	if queue.ConditionalGroupRouting != nil && queue.ConditionalGroupRouting.Rules != nil {
		return queue.ConditionalGroupRouting.Rules, resp, nil
	}

	return nil, resp, nil
}

// updateRoutingQueueConditionRoutingFn is an implementation function for updating the conditional group routing rules for a queue
func updateRoutingQueueConditionRoutingFn(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string, rules *[]platformclientv2.Conditionalgrouproutingrule) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
	// Get the routing queue the rules belong to
	queue, resp, err := p.getRoutingQueueById(ctx, queueId)
	if err != nil {
		return nil, resp, fmt.Errorf("error when reading queue %s: %s", queueId, err)
	}

	groupRoutingObj := platformclientv2.Conditionalgrouprouting{Rules: rules}

	// Copy over all the values from the original object to the new object
	updateQueue := platformclientv2.Queuerequest{
		Name:                         queue.Name,
		Description:                  queue.Description,
		MemberCount:                  queue.MemberCount,
		UserMemberCount:              queue.UserMemberCount,
		JoinedMemberCount:            queue.JoinedMemberCount,
		MediaSettings:                queue.MediaSettings,
		RoutingRules:                 queue.RoutingRules,
		ConditionalGroupRouting:      &groupRoutingObj, // Add the new rules
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

	// For some reason OutboundEmailAddress returned by GetRoutingQueue is a pointer to a pointer so I am handling it here
	if queue.OutboundEmailAddress != nil && *queue.OutboundEmailAddress != nil {
		updateQueue.OutboundEmailAddress = *queue.OutboundEmailAddress
	}

	// Update the queue with th new rules
	queue, resp, err = p.routingApi.PutRoutingQueue(queueId, updateQueue)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update conditional group routing rules for routing queue %s : %s", queueId, err)
	}

	if queue.ConditionalGroupRouting != nil && queue.ConditionalGroupRouting.Rules != nil {
		return queue.ConditionalGroupRouting.Rules, resp, nil
	}

	return nil, resp, nil
}

// getRoutingQueueByIdFn is an implementation function for getting a queue by ID
func getRoutingQueueByIdFn(_ context.Context, p *routingQueueConditionalGroupRoutingProxy, id string) (*platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.routingApi.GetRoutingQueue(id)
}
