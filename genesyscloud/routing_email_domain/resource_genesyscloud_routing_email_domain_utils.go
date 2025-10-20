package routing_email_domain

import (
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v170/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"strings"
	"time"
)

func CleanupRoutingEmailDomains(prefix string) error {
	sdkConfig, _ := provider.AuthorizeSdk()
	routingAPI := platformclientv2.NewRoutingApiWithConfig(sdkConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		routingEmailDomains, _, getErr := routingAPI.GetRoutingEmailDomains(pageSize, pageNum, false, "", "")
		if getErr != nil {
			return fmt.Errorf("failed to get page %v of routing email domains: %v", pageNum, getErr)

		}

		if routingEmailDomains.Entities == nil || len(*routingEmailDomains.Entities) == 0 {
			break
		}

		for _, routingEmailDomain := range *routingEmailDomains.Entities {
			if routingEmailDomain.Id != nil && strings.HasPrefix(*routingEmailDomain.Id, prefix) {
				_, err := routingAPI.DeleteRoutingEmailDomain(*routingEmailDomain.Id)
				if err != nil {
					return fmt.Errorf("failed to delete routing email domain %s: %s", *routingEmailDomain.Id, err)
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
	return nil
}
