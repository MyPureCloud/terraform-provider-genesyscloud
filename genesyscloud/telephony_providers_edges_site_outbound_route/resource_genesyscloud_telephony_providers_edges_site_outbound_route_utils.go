package telephony_providers_edges_site_outbound_route

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
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

// Any routes that are associated with managed site resources should be exported as data
func shouldExportRoutesWithManagedSitesAsData(ctx context.Context, sdkConfig *platformclientv2.Configuration, configMap map[string]string) (exportAsData bool, err error) {

	// Check if the site exists
	siteId := configMap["site_id"]
	if siteId == "" {
		return false, fmt.Errorf("site_id is not set")
	}

	proxy := getSiteOutboundRouteProxy(sdkConfig)
	site, _, err := proxy.siteProxy.GetSiteById(ctx, siteId)
	if err != nil {
		return false, err
	}

	return *site.Managed, nil
}
