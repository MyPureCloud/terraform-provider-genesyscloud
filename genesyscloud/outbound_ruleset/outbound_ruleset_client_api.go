package outbound_ruleset

import (
	"fmt"
	"context"
	"time"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	resource_exporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)



type OutboundAPIProxy struct {
	oAPI *platformclientv2.OutboundApi

	CreateOutboundRulesets   postOutboundRulesetsFunc
	ReadOutboundRuleset    getOutboundRulesetFunc
	UpdateOutboundRuleset    putOutboundRulesetFunc
	DeleteOutboundRuleset removeOutboundRulesetFunc
	ReadAllOutboundRuleset   getAllOutboundRulesetFunc
	ReadOutboundRulesetsData getOutboundRulesetsDataFunc
}

func NewOutboundAPIProxy() *OutboundAPIProxy {
	var outboundApi *platformclientv2.OutboundApi
	return &OutboundAPIProxy{
		oAPI: outboundApi,

		CreateOutboundRulesets:   postOutboundRulesets,
		ReadOutboundRuleset:    getOutboundRuleset,
		UpdateOutboundRuleset:    putOutboundRuleset,
		DeleteOutboundRuleset: removeOutboundRuleset,
		ReadAllOutboundRuleset:   getOutboundRulesets,
		ReadOutboundRulesetsData: getOutboundRulesetsData,
	}
}

type postOutboundRulesetsFunc func(*OutboundAPIProxy, platformclientv2.Ruleset)  (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error)

type getOutboundRulesetFunc func(*OutboundAPIProxy, string)  (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error)

type putOutboundRulesetFunc func(*OutboundAPIProxy, platformclientv2.Ruleset, string)  (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error)

type removeOutboundRulesetFunc func(*OutboundAPIProxy, string) (*platformclientv2.APIResponse, error)

type getAllOutboundRulesetFunc func(*OutboundAPIProxy)  (resource_exporter.ResourceIDMetaMap, diag.Diagnostics)

type getOutboundRulesetsDataFunc func(context.Context, *OutboundAPIProxy, *schema.ResourceData, string) (diag.Diagnostics)


func (a *OutboundAPIProxy) ConfigureProxyApiInstance(c *platformclientv2.Configuration) {
	a.oAPI = platformclientv2.NewOutboundApiWithConfig(c)
}

func postOutboundRulesets( a *OutboundAPIProxy, sdkruleset platformclientv2.Ruleset) (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error) {
	return a.oAPI.PostOutboundRulesets(sdkruleset)	 
}

func putOutboundRuleset(a *OutboundAPIProxy,sdkruleset platformclientv2.Ruleset, id string) (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error) {
	return a.oAPI.PutOutboundRuleset(id, sdkruleset)
}

func getOutboundRuleset(a *OutboundAPIProxy,id string) (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error) {
	return a.oAPI.GetOutboundRuleset(id)
}

func removeOutboundRuleset(a *OutboundAPIProxy,id string) (*platformclientv2.APIResponse, error) {
	return a.oAPI.DeleteOutboundRuleset(id)
}

func getOutboundRulesets(a *OutboundAPIProxy) (resource_exporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resource_exporter.ResourceIDMetaMap)
	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		sdkrulesetentitylisting, _, getErr := a.oAPI.GetOutboundRulesets(pageSize, pageNum, true, "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Error requesting page of Outbound Ruleset: %s", getErr)
		}

		if sdkrulesetentitylisting.Entities == nil || len(*sdkrulesetentitylisting.Entities) == 0 {
			break
		}

		for _, entity := range *sdkrulesetentitylisting.Entities {
			resources[*entity.Id] = &resource_exporter.ResourceMeta{Name: *entity.Name}
		}
	}

	return resources, nil
}

func getOutboundRulesetsData(ctx context.Context, a *OutboundAPIProxy, d *schema.ResourceData, name string) (diag.Diagnostics) {

	return gcloud.WithRetries(ctx, 15*time.Second, func() *resource.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			sdkrulesetentitylisting, _, getErr := a.oAPI.GetOutboundRulesets(pageSize, pageNum, false, "", "", "", "")
			if getErr != nil {
				return resource.NonRetryableError(fmt.Errorf("Error requesting Outbound Ruleset %s: %s", name, getErr))
			}

			if sdkrulesetentitylisting.Entities == nil || len(*sdkrulesetentitylisting.Entities) == 0 {
				return resource.RetryableError(fmt.Errorf("No Outbound Ruleset found with name %s", name))
			}

			for _, entity := range *sdkrulesetentitylisting.Entities {
				if entity.Name != nil && *entity.Name == name {
					d.SetId(*entity.Id)
					return nil
				}
			}
		}
	})
}
