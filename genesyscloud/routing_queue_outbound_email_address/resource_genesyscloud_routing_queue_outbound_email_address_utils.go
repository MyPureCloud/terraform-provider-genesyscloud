package routing_queue_outbound_email_address

import "github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

func isQueueEmailAddressEmpty(qea *platformclientv2.Queueemailaddress) bool {
	if qea == nil {
		return true
	}

	// Compare relevant fields of the struct
	if qea.Domain == nil || qea.Domain.Id == nil || *qea.Domain.Id == "" {
		return true
	}

	if qea.Route != nil && *qea.Route != nil {
		routeId := (*qea.Route).Id
		if routeId == nil || *routeId == "" {
			return true
		}
	} else {
		return true
	}

	return false
}
