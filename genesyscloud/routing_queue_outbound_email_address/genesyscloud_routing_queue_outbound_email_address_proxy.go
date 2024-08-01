package routing_queue_outbound_email_address

import (
	"context"
	"fmt"
	rc "terraform-provider-genesyscloud/genesyscloud/resource_cache"
	routingQueue "terraform-provider-genesyscloud/genesyscloud/routing_queue"

	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingQueueOutboundEmailAddressProxy

type getRoutingQueueOutboundEmailAddressFunc func(ctx context.Context, p *routingQueueOutboundEmailAddressProxy, queueId string) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error)
type updateRoutingQueueOutboundEmailAddressFunc func(ctx context.Context, p *routingQueueOutboundEmailAddressProxy, queueId string, address *platformclientv2.Queueemailaddress) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error)

// routingQueueOutboundEmailAddressProxy contains all of the methods that call genesys cloud APIs.
type routingQueueOutboundEmailAddressProxy struct {
	clientConfig                               *platformclientv2.Configuration
	routingApi                                 *platformclientv2.RoutingApi
	getRoutingQueueOutboundEmailAddressAttr    getRoutingQueueOutboundEmailAddressFunc
	updateRoutingQueueOutboundEmailAddressAttr updateRoutingQueueOutboundEmailAddressFunc
	routingQueueProxy                          *routingQueue.RoutingQueueProxy
}

// newRoutingQueueOutboundEmailAddressProxy initializes the Routing queue outbound email address proxy with the data needed to communicate with Genesys Cloud
func newRoutingQueueOutboundEmailAddressProxy(clientConfig *platformclientv2.Configuration) *routingQueueOutboundEmailAddressProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	routingQueueProxy := routingQueue.GetRoutingQueueProxy(clientConfig)

	return &routingQueueOutboundEmailAddressProxy{
		clientConfig:                            clientConfig,
		routingApi:                              api,
		getRoutingQueueOutboundEmailAddressAttr: getRoutingQueueOutboundEmailAddressFn,
		updateRoutingQueueOutboundEmailAddressAttr: updateRoutingQueueOutboundEmailAddressFn,
		routingQueueProxy:                          routingQueueProxy,
	}
}

func getRoutingQueueOutboundEmailAddressProxy(clientConfig *platformclientv2.Configuration) *routingQueueOutboundEmailAddressProxy {
	if internalProxy == nil {
		internalProxy = newRoutingQueueOutboundEmailAddressProxy(clientConfig)
	}

	return internalProxy
}

// getRoutingQueueOutboundEmailAddress gets the Outbound Email Address for a queue
func (p *routingQueueOutboundEmailAddressProxy) getRoutingQueueOutboundEmailAddress(ctx context.Context, queueId string) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error) {
	return p.getRoutingQueueOutboundEmailAddressAttr(ctx, p, queueId)
}

// updateRoutingQueueOutboundEmailAddress updates the Outbound Email Address for a queue
func (p *routingQueueOutboundEmailAddressProxy) updateRoutingQueueOutboundEmailAddress(ctx context.Context, queueId string, rules *platformclientv2.Queueemailaddress) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error) {
	return p.updateRoutingQueueOutboundEmailAddressAttr(ctx, p, queueId, rules)
}

// getRoutingQueueOutboundEmailAddressFn is an implementation function for getting the outbound email address for a queue
func getRoutingQueueOutboundEmailAddressFn(ctx context.Context, p *routingQueueOutboundEmailAddressProxy, queueId string) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error) {
	var (
		queue *platformclientv2.Queue
		resp  *platformclientv2.APIResponse
		err   error
	)

	queue = rc.GetCacheItem(p.routingQueueProxy.RoutingQueueCache, queueId)
	if queue == nil {
		queue, resp, err = p.routingApi.GetRoutingQueue(queueId)
		if err != nil {
			return nil, resp, fmt.Errorf("error when reading queue %s: %s", queueId, err)
		}
	}

	queue, resp, err = p.routingApi.GetRoutingQueue(queueId)
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
