package telephony_providers_edges_site_outbound_route

import (
	"log"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func buildOutboundRoutes(outboundRoutes *schema.Set) *[]platformclientv2.Outboundroutebase {
	outboundRoutesList := outboundRoutes.List()

	outboundRoutesSdk := make([]platformclientv2.Outboundroutebase, 0)
	for _, outboundRoute := range outboundRoutesList {
		outboundRoutesMap := outboundRoute.(map[string]interface{})
		outboundRouteSdk := platformclientv2.Outboundroutebase{}

		resourcedata.BuildSDKStringValueIfNotNil(&outboundRouteSdk.Name, outboundRoutesMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&outboundRouteSdk.Description, outboundRoutesMap, "description")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&outboundRouteSdk.ClassificationTypes, outboundRoutesMap, "classification_types")
		if enabled, ok := outboundRoutesMap["enabled"].(bool); ok {
			outboundRouteSdk.Enabled = &enabled
		}
		resourcedata.BuildSDKStringValueIfNotNil(&outboundRouteSdk.Distribution, outboundRoutesMap, "distribution")

		if externalTrunkBaseIds, ok := outboundRoutesMap["external_trunk_base_ids"].([]interface{}); ok && len(externalTrunkBaseIds) > 0 {
			ids := make([]platformclientv2.Domainentityref, 0)
			for _, externalTrunkBaseId := range externalTrunkBaseIds {
				externalTrunkBaseIdStr := externalTrunkBaseId.(string)
				ids = append(ids, platformclientv2.Domainentityref{Id: &externalTrunkBaseIdStr})
			}
			outboundRouteSdk.ExternalTrunkBases = &ids
		}

		outboundRoutesSdk = append(outboundRoutesSdk, outboundRouteSdk)
	}

	return &outboundRoutesSdk
}

func nameInOutboundRoutes(name string, outboundRoutes []platformclientv2.Outboundroutebase) (*platformclientv2.Outboundroutebase, bool) {
	for _, outboundRoute := range outboundRoutes {
		if name == *outboundRoute.Name {
			return &outboundRoute, true
		}
	}

	return nil, false
}

func checkExistingRoutes(definedRoutes, apiRoutes *[]platformclientv2.Outboundroutebase, siteId string) (newRoutes []platformclientv2.Outboundroutebase) {
	for _, definedRoute := range *definedRoutes {
		if _, present := nameInOutboundRoutes(*definedRoute.Name, *apiRoutes); present {
			log.Printf("Route %s associated with site %s already exists. Creating only non-existing routes", *definedRoute.Name, siteId)
		} else {
			newRoutes = append(newRoutes, definedRoute)
		}
	}
	return newRoutes
}
