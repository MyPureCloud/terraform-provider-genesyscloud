package routing_queue_outbound_email_address

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingQueueOutboundEmailAddressProxy

type getAllRoutingQueuesFunc func(ctx context.Context, p *routingQueueOutboundEmailAddressProxy) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error)
type getRoutingQueueOutboundEmailAddressFunc func(ctx context.Context, p *routingQueueOutboundEmailAddressProxy, queueId string) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error)
type updateRoutingQueueOutboundEmailAddressFunc func(ctx context.Context, p *routingQueueOutboundEmailAddressProxy, queueId string, address *platformclientv2.Queueemailaddress) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error)

// routingQueueOutboundEmailAddressProxy contains all of the methods that call genesys cloud APIs.
type routingQueueOutboundEmailAddressProxy struct {
	clientConfig                               *platformclientv2.Configuration
	routingApi                                 *platformclientv2.RoutingApi
	getAllRoutingQueuesAttr                    getAllRoutingQueuesFunc
	getRoutingQueueOutboundEmailAddressAttr    getRoutingQueueOutboundEmailAddressFunc
	updateRoutingQueueOutboundEmailAddressAttr updateRoutingQueueOutboundEmailAddressFunc
}

// newRoutingQueueOutboundEmailAddressProxy initializes the Routing queue outbound email address proxy with the data needed to communicate with Genesys Cloud
func newRoutingQueueOutboundEmailAddressProxy(clientConfig *platformclientv2.Configuration) *routingQueueOutboundEmailAddressProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingQueueOutboundEmailAddressProxy{
		clientConfig:                               clientConfig,
		routingApi:                                 api,
		getAllRoutingQueuesAttr:                    getAllRoutingQueuesFn,
		getRoutingQueueOutboundEmailAddressAttr:    getRoutingQueueOutboundEmailAddressFn,
		updateRoutingQueueOutboundEmailAddressAttr: updateRoutingQueueOutboundEmailAddressFn,
	}
}

func getRoutingQueueOutboundEmailAddressProxy(clientConfig *platformclientv2.Configuration) *routingQueueOutboundEmailAddressProxy {
	if internalProxy == nil {
		internalProxy = newRoutingQueueOutboundEmailAddressProxy(clientConfig)
	}

	return internalProxy
}

// getAllRoutingQueues gets all routing queues in an org
func (p *routingQueueOutboundEmailAddressProxy) getAllRoutingQueues(ctx context.Context) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return p.getAllRoutingQueuesAttr(ctx, p)
}

// getRoutingQueueOutboundEmailAddress gets the Outbound Email Address for a queue
func (p *routingQueueOutboundEmailAddressProxy) getRoutingQueueOutboundEmailAddress(ctx context.Context, queueId string) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueOutboundEmailAddressAttr(ctx, p, queueId)
}

// updateRoutingQueueOutboundEmailAddress updates the Outbound Email Address for a queue
func (p *routingQueueOutboundEmailAddressProxy) updateRoutingQueueOutboundEmailAddress(ctx context.Context, queueId string, rules *platformclientv2.Queueemailaddress) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error) {
	return p.updateRoutingQueueOutboundEmailAddressAttr(ctx, p, queueId, rules)
}

// getAllRoutingQueuesFn is an implementation function for getting all queues in an org
func getAllRoutingQueuesFn(ctx context.Context, p *routingQueueOutboundEmailAddressProxy) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
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

// getRoutingQueueOutboundEmailAddressFn is an implementation function for getting the outbound email address for a queue
func getRoutingQueueOutboundEmailAddressFn(ctx context.Context, p *routingQueueOutboundEmailAddressProxy, queueId string) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error) {
	queue, resp, err := p.routingApi.GetRoutingQueue(queueId)
	if err != nil {
		return nil, resp, fmt.Errorf("error when reading queue %s: %s", queueId, err)
	}

	// For some reason outbound email address is a double pointer
	if queue.OutboundEmailAddress != nil && *queue.OutboundEmailAddress != nil {
		return *queue.OutboundEmailAddress, resp, nil
	}

	return nil, resp, fmt.Errorf("no outbound email address for queue %s", queueId)
}

// updateRoutingQueueOutboundEmailAddressFn is an implementation function for updating the outbound email address for a queue
func updateRoutingQueueOutboundEmailAddressFn(ctx context.Context, p *routingQueueOutboundEmailAddressProxy, queueId string, address *platformclientv2.Queueemailaddress) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error) {
	// Get the routing queue the rules belong to
	queue, resp, err := p.routingApi.GetRoutingQueue(queueId)
	if err != nil {
		return nil, resp, fmt.Errorf("error when reading queue %s: %s", queueId, err)
	}

	// Copy over all the values from the original object to the new object
	updateQueue := platformclientv2.Queuerequest{
		Name:                         queue.Name,
		Description:                  queue.Description,
		MemberCount:                  queue.MemberCount,
		UserMemberCount:              queue.UserMemberCount,
		JoinedMemberCount:            queue.JoinedMemberCount,
		MediaSettings:                queue.MediaSettings,
		RoutingRules:                 queue.RoutingRules,
		ConditionalGroupRouting:      queue.ConditionalGroupRouting,
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
		OutboundEmailAddress:         address, // Add the new address
		PeerId:                       queue.PeerId,
		SuppressInQueueCallRecording: queue.SuppressInQueueCallRecording,
	}

	// Update the queue
	queue, resp, err = p.routingApi.PutRoutingQueue(queueId, updateQueue)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to update outbound email address for routing queue %s: %s", queueId, err)
	}

	if queue.OutboundEmailAddress != nil && *queue.OutboundEmailAddress != nil {
		return *queue.OutboundEmailAddress, resp, nil
	}

	return nil, resp, fmt.Errorf("error updating outbound email address for routing queue %s: %s", queueId, err)
}
