package routing_queue_conditional_group_routing

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingQueueConditionalGroupRoutingProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type getAllRoutingQueuesFunc func(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error)
type getRoutingQueueConditionRoutingFunc func(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error)
type updateRoutingQueueConditionRoutingFunc func(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string, rules *[]platformclientv2.Conditionalgrouproutingrule) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error)

// routingQueueConditionalGroupRoutingProxy contains all of the methods that call genesys cloud APIs.
type routingQueueConditionalGroupRoutingProxy struct {
	clientConfig                           *platformclientv2.Configuration
	routingApi                             *platformclientv2.RoutingApi
	getAllRoutingQueueConditionRoutingAttr getAllRoutingQueuesFunc
	getRoutingQueueConditionRoutingAttr    getRoutingQueueConditionRoutingFunc
	updateRoutingQueueConditionRoutingAttr updateRoutingQueueConditionRoutingFunc
}

// newRoutingQueueConditionalGroupRoutingProxy initializes the Routing queue conditional group routing proxy with all of the data needed to communicate with Genesys Cloud
func newRoutingQueueConditionalGroupRoutingProxy(clientConfig *platformclientv2.Configuration) *routingQueueConditionalGroupRoutingProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingQueueConditionalGroupRoutingProxy{
		clientConfig:                           clientConfig,
		routingApi:                             api,
		getAllRoutingQueueConditionRoutingAttr: getAllRoutingQueuesFn,
		getRoutingQueueConditionRoutingAttr:    getRoutingQueueConditionRoutingFn,
		updateRoutingQueueConditionRoutingAttr: updateRoutingQueueConditionRoutingFn,
	}
}

// getRoutingQueueConditionalGroupRoutingProxy retrieves all Genesys Cloud Routing queue conditional group routing
func getRoutingQueueConditionalGroupRoutingProxy(clientConfig *platformclientv2.Configuration) *routingQueueConditionalGroupRoutingProxy {
	if internalProxy == nil {
		internalProxy = newRoutingQueueConditionalGroupRoutingProxy(clientConfig)
	}

	return internalProxy
}

// getAllRoutingQueues gets all routing queues in an org
func (p *routingQueueConditionalGroupRoutingProxy) getAllRoutingQueues(ctx context.Context) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingQueueConditionRoutingAttr(ctx, p)
}

// getRoutingQueueConditionRouting gets the conditional group routing rules for a queue
func (p *routingQueueConditionalGroupRoutingProxy) getRoutingQueueConditionRouting(ctx context.Context, queueId string) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueConditionRoutingAttr(ctx, p, queueId)
}

// updateRoutingQueueConditionRouting updates the conditional group routing rules for a queue
func (p *routingQueueConditionalGroupRoutingProxy) updateRoutingQueueConditionRouting(ctx context.Context, queueId string, rules *[]platformclientv2.Conditionalgrouproutingrule) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
	return p.updateRoutingQueueConditionRoutingAttr(ctx, p, queueId, rules)
}

// getAllRoutingQueuesFn is an implementation function for getting all queues in an org
func getAllRoutingQueuesFn(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	var allQueues []platformclientv2.Queue
	const pageSize = 100

	queues, resp, err := p.routingApi.GetRoutingQueues(1, pageSize, "", "", nil, nil, nil, false)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to get routing queues: %s", err)
	}

	if queues.Entities == nil || len(*queues.Entities) == 0 {
		return &allQueues, resp, nil
	}

	for _, queue := range *queues.Entities {
		allQueues = append(allQueues, queue)
	}

	for pageNum := 2; pageNum <= *queues.PageCount; pageNum++ {
		queues, resp, err := p.routingApi.GetRoutingQueues(pageNum, pageSize, "", "", nil, nil, nil, false)
		if err != nil {
			return nil, resp, fmt.Errorf("failed to get routing queues: %s", err)
		}

		if queues.Entities == nil || len(*queues.Entities) == 0 {
			break
		}

		for _, queue := range *queues.Entities {
			allQueues = append(allQueues, queue)
		}
	}

	return &allQueues, nil, nil
}

// getRoutingQueueConditionRoutingFn is an implementation function for getting the conditional group routing rules for a queue
func getRoutingQueueConditionRoutingFn(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
	queue, resp, err := p.routingApi.GetRoutingQueue(queueId)
	if err != nil {
		return nil, resp, fmt.Errorf("routing queue %s not found: %s", queueId, err)
	}

	if queue.ConditionalGroupRouting != nil && queue.ConditionalGroupRouting.Rules != nil {
		return queue.ConditionalGroupRouting.Rules, resp, nil
	}

	return nil, resp, fmt.Errorf("no conditional group routing rules found for queue %s", queueId)
}

// updateRoutingQueueConditionRoutingFn is an implementation function for updating the conditional group routing rules for a queue
func updateRoutingQueueConditionRoutingFn(ctx context.Context, p *routingQueueConditionalGroupRoutingProxy, queueId string, rules *[]platformclientv2.Conditionalgrouproutingrule) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
	// Get the routing queue the rules belong to
	queue, resp, err := p.routingApi.GetRoutingQueue(queueId)
	if err != nil {
		return nil, resp, fmt.Errorf("routing queue %s not found: %s", queueId, err)
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
		return nil, resp, fmt.Errorf("failed to update routing queue %s conditional group routing rules: %s", queueId, err)
	}

	if queue.ConditionalGroupRouting != nil && queue.ConditionalGroupRouting.Rules != nil {
		return queue.ConditionalGroupRouting.Rules, resp, nil
	}

	return nil, resp, fmt.Errorf("no conditional group routing rules found for queue %s", queueId)
}
