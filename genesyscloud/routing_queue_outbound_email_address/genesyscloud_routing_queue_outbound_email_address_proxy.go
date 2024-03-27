package routing_queue_outbound_email_address

import (
	"context"
	"github.com/mypurecloud/platform-client-sdk-go/v125/platformclientv2"
)

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingQueueOutboundEmailAddressProxy

// routingQueueOutboundEmailAddressProxy contains all of the methods that call genesys cloud APIs.
type routingQueueOutboundEmailAddressProxy struct {
	clientConfig *platformclientv2.Configuration
	routingApi   *platformclientv2.RoutingApi
}

// newRoutingQueueConditionalGroupRoutingProxy initializes the Routing queue conditional group routing proxy with all of the data needed to communicate with Genesys Cloud
func newRoutingQueueOutboundEmailAddressProxy(clientConfig *platformclientv2.Configuration) *routingQueueOutboundEmailAddressProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingQueueOutboundEmailAddressProxy{
		clientConfig: clientConfig,
		routingApi:   api,
	}
}

// getRoutingQueueConditionalGroupRoutingProxy retrieves all Genesys Cloud Routing queue conditional group routing
func getRoutingQueueOutboundEmailAddressProxy(clientConfig *platformclientv2.Configuration) *routingQueueOutboundEmailAddressProxy {
	if internalProxy == nil {
		internalProxy = newRoutingQueueOutboundEmailAddressProxy(clientConfig)
	}

	return internalProxy
}

// getAllRoutingQueues gets all routing queues in an org
func (p *routingQueueOutboundEmailAddressProxy) getAllRoutingQueues(ctx context.Context) (*[]platformclientv2.Queue, *platformclientv2.APIResponse, error) {
	return nil, nil, nil
}

// getRoutingQueueConditionRouting gets the conditional group routing rules for a queue
func (p *routingQueueOutboundEmailAddressProxy) getRoutingQueueConditionRouting(ctx context.Context, queueId string) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
	return nil, nil, nil
}

// updateRoutingQueueConditionRouting updates the conditional group routing rules for a queue
func (p *routingQueueOutboundEmailAddressProxy) updateRoutingQueueConditionRouting(ctx context.Context, queueId string, rules *[]platformclientv2.Conditionalgrouproutingrule) (*[]platformclientv2.Conditionalgrouproutingrule, *platformclientv2.APIResponse, error) {
	return nil, nil, nil
}
