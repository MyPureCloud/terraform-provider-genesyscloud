package telephony_providers_edges_site_outbound_route

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v149/platformclientv2"
)

func buildOutboundRoutes(d *schema.ResourceData) *platformclientv2.Outboundroutebase {

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)
	distribution := d.Get("distribution").(string)
	outboundRouteSdk := platformclientv2.Outboundroutebase{
		Name:         &name,
		Description:  &description,
		Enabled:      &enabled,
		Distribution: &distribution,
	}

	if classificationTypes, ok := d.Get("classification_types").([]interface{}); ok && len(classificationTypes) > 0 {
		cts := make([]string, 0)
		for _, classificationType := range classificationTypes {
			cts = append(cts, classificationType.(string))
		}
		outboundRouteSdk.ClassificationTypes = &cts
	}

	if externalTrunkBaseIdsRaw, ok := d.GetOk("external_trunk_base_ids"); ok {
		if externalTrunkBaseIds, ok := externalTrunkBaseIdsRaw.([]interface{}); ok && len(externalTrunkBaseIds) > 0 {
			ids := make([]platformclientv2.Domainentityref, 0)
			for _, externalTrunkBaseId := range externalTrunkBaseIds {
				externalTrunkBaseIdStr := externalTrunkBaseId.(string)
				ids = append(ids, platformclientv2.Domainentityref{Id: &externalTrunkBaseIdStr})
			}
			outboundRouteSdk.ExternalTrunkBases = &ids
		}
	}

	return &outboundRouteSdk
}

func buildSiteAndOutboundRouteId(siteId string, outboundRouteId string) string {
	fullOutboundRouteId := fmt.Sprintf("%s:%s", siteId, outboundRouteId)
	return fullOutboundRouteId
}

func splitSiteAndOutboundRoute(dId string) (string, string) {
	split := strings.SplitN(dId, ":", 2)
	if len(split) == 2 {
		return split[0], split[1]
	}
	return "", ""
}
