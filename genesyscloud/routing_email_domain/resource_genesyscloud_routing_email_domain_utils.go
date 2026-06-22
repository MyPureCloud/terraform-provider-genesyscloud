package routing_email_domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v192/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
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

func flattenGraphApiSettings(settings *platformclientv2.Graphapisettings) []interface{} {
	if settings == nil {
		return nil
	}

	settingsMap := make(map[string]interface{})
	resourcedata.SetMapReferenceValueIfNotNil(settingsMap, "integration_id", settings.Integration)
	resourcedata.SetMapValueIfNotNil(settingsMap, "status", settings.Status)

	return []interface{}{settingsMap}
}

func flattenImapSettings(settings *platformclientv2.Imapsettings) []interface{} {
	if settings == nil {
		return nil
	}

	settingsMap := make(map[string]interface{})
	resourcedata.SetMapReferenceValueIfNotNil(settingsMap, "integration_id", settings.Integration)
	resourcedata.SetMapValueIfNotNil(settingsMap, "status", settings.Status)

	return []interface{}{settingsMap}
}

func expandGraphApiSettings(d *schema.ResourceData) *platformclientv2.Graphapisettings {
	raw, ok := d.GetOk("graph_api_settings")
	if !ok {
		return nil
	}

	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}

	settingsMap, ok := list[0].(map[string]interface{})
	if !ok {
		return nil
	}

	integrationID, ok := settingsMap["integration_id"].(string)
	if !ok || integrationID == "" {
		return nil
	}

	return &platformclientv2.Graphapisettings{
		Integration: &platformclientv2.Domainentityref{
			Id: &integrationID,
		},
	}
}

func expandImapSettings(d *schema.ResourceData) *platformclientv2.Imapsettings {
	raw, ok := d.GetOk("imap_settings")
	if !ok {
		return nil
	}

	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}

	settingsMap, ok := list[0].(map[string]interface{})
	if !ok {
		return nil
	}

	integrationID, ok := settingsMap["integration_id"].(string)
	if !ok || integrationID == "" {
		return nil
	}

	return &platformclientv2.Imapsettings{
		Integration: &platformclientv2.Domainentityref{
			Id: &integrationID,
		},
	}
}
