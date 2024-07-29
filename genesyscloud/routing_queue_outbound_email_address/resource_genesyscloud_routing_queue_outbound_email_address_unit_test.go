package routing_queue_outbound_email_address

import (
	"context"
	"net/http"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	featureToggles "terraform-provider-genesyscloud/genesyscloud/util/feature_toggles"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceRoutingQueueOutboundEmailAddressUpdate(t *testing.T) {
	tQueueId := uuid.NewString()
	tDomainId := uuid.NewString()
	tRouteId := uuid.NewString()

	tId := tQueueId

	if !featureToggles.OEAToggleExists() {
		t.Skipf("Skipping because env variable %s is not set", featureToggles.OEAToggleName())
	}

	groupRoutingProxy := &routingQueueOutboundEmailAddressProxy{}
	groupRoutingProxy.updateRoutingQueueOutboundEmailAddressAttr = func(ctx context.Context, p *routingQueueOutboundEmailAddressProxy, queueId string, address *platformclientv2.Queueemailaddress) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error) {
		assert.Equal(t, tDomainId, *address.Domain.Id)
		assert.Equal(t, tRouteId, *(*address.Route).Id)

		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}
		return address, &apiResponse, nil
	}

	groupRoutingProxy.getRoutingQueueOutboundEmailAddressAttr = func(ctx context.Context, p *routingQueueOutboundEmailAddressProxy, queueId string) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error) {
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}

		route := &platformclientv2.Inboundroute{
			Id: &tRouteId,
		}
		address := platformclientv2.Queueemailaddress{
			Domain: &platformclientv2.Domainentityref{Id: &tDomainId},
			Route:  &route,
		}

		return &address, &apiResponse, nil
	}

	internalProxy = groupRoutingProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceRoutingQueueOutboundEmailAddress().Schema

	//Setup a map of values
	resourceDataMap := buildOutboundEmailAddressResourceMap(tQueueId, tDomainId, tRouteId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := updateRoutingQueueOutboundEmailAddress(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError(), diag)
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tQueueId, d.Get("queue_id").(string))
	assert.Equal(t, tRouteId, d.Get("route_id").(string))
	assert.Equal(t, tDomainId, d.Get("domain_id").(string))
}

func TestUnitResourceRoutingQueueOutboundEmailAddressRead(t *testing.T) {
	tQueueId := uuid.NewString()
	tDomainId := uuid.NewString()
	tRouteId := uuid.NewString()
	tId := tQueueId

	if !featureToggles.OEAToggleExists() {
		t.Skipf("Skipping because env variable %s is not set", featureToggles.OEAToggleName())
	}

	groupRoutingProxy := &routingQueueOutboundEmailAddressProxy{}

	groupRoutingProxy.getRoutingQueueOutboundEmailAddressAttr = func(ctx context.Context, p *routingQueueOutboundEmailAddressProxy, queueId string) (*platformclientv2.Queueemailaddress, *platformclientv2.APIResponse, error) {
		apiResponse := platformclientv2.APIResponse{StatusCode: http.StatusOK}

		route := &platformclientv2.Inboundroute{
			Id: &tRouteId,
		}
		address := platformclientv2.Queueemailaddress{
			Domain: &platformclientv2.Domainentityref{Id: &tDomainId},
			Route:  &route,
		}

		return &address, &apiResponse, nil
	}

	internalProxy = groupRoutingProxy
	defer func() { internalProxy = nil }()

	ctx := context.Background()
	gcloud := &provider.ProviderMeta{ClientConfig: &platformclientv2.Configuration{}}

	//Grab our defined schema
	resourceSchema := ResourceRoutingQueueOutboundEmailAddress().Schema

	//Setup a map of values
	resourceDataMap := buildOutboundEmailAddressResourceMap(tQueueId, tDomainId, tRouteId)

	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId(tId)

	diag := readRoutingQueueOutboundEmailAddress(ctx, d, gcloud)
	assert.Equal(t, false, diag.HasError(), diag)
	assert.Equal(t, tId, d.Id())
	assert.Equal(t, tQueueId, d.Get("queue_id").(string))
	assert.Equal(t, tDomainId, d.Get("domain_id").(string))
	assert.Equal(t, tRouteId, d.Get("route_id").(string))
}

func buildOutboundEmailAddressResourceMap(queueId string, domainId string, routeId string) map[string]interface{} {
	resourceDataMap := map[string]interface{}{
		"queue_id":  queueId,
		"domain_id": domainId,
		"route_id":  routeId,
	}

	return resourceDataMap
}
