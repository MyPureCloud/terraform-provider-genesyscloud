package outbound_ruleset

import (
	"github.com/mypurecloud/platform-client-sdk-go/v102/platformclientv2"
)

func postOutboundRulesets(sdkruleset platformclientv2.Ruleset, outboundApi *platformclientv2.OutboundApi) (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error) {
	return outboundApi.PostOutboundRulesets(sdkruleset)	 
}

func putOutboundRuleset(sdkruleset platformclientv2.Ruleset, outboundApi *platformclientv2.OutboundApi, id string) (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error) {
	return outboundApi.PutOutboundRuleset(id, sdkruleset)
}

func getOutboundRuleset(id string, outboundApi *platformclientv2.OutboundApi) (*platformclientv2.Ruleset, *platformclientv2.APIResponse, error) {
	return outboundApi.GetOutboundRuleset(id)
}

func removeOutboundRuleset(id string, outboundApi *platformclientv2.OutboundApi) (*platformclientv2.APIResponse, error) {
	return outboundApi.DeleteOutboundRuleset(id)
}

func getOutboundRulesets(pageSize int, pageNum int, allowEmptyResult bool, outboundApi *platformclientv2.OutboundApi) (*platformclientv2.Rulesetentitylisting, *platformclientv2.APIResponse, error) {
	return outboundApi.GetOutboundRulesets(pageSize, pageNum, allowEmptyResult, "", "", "", "")
}
