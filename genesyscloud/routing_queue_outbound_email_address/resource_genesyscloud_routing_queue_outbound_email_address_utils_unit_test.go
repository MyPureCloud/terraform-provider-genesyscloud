package routing_queue_outbound_email_address

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitResourceRoutingQueueOutboundEmailAddressEmpty(t *testing.T) {
	address := platformclientv2.Queueemailaddress{}
	result := isQueueEmailAddressEmpty(&address)
	assert.Equal(t, true, result)
}

func TestUnitResourceRoutingQueueOutboundEmailAddressNotEmpty(t *testing.T) {
	tDomainId := uuid.NewString()
	tRouteId := uuid.NewString()

	route := &platformclientv2.Inboundroute{
		Id: &tRouteId,
	}
	address := platformclientv2.Queueemailaddress{
		Domain: &platformclientv2.Domainentityref{Id: &tDomainId},
		Route:  &route,
	}
	result := isQueueEmailAddressEmpty(&address)
	assert.Equal(t, false, result)
}
