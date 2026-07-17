package routing_queue_identity_resolution

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v193/platformclientv2"
)

func buildIdentityResolutionQueueConfig(d *schema.ResourceData) (platformclientv2.Identityresolutionqueueconfig, error) {
	inner, err := buildCallOnBehalfOfQueueConfig(d.Get("call_on_behalf_of_queue").([]interface{}))
	if err != nil {
		return platformclientv2.Identityresolutionqueueconfig{}, err
	}

	return platformclientv2.Identityresolutionqueueconfig{
		CallOnBehalfOfQueue: inner,
	}, nil
}

func buildDefaultIdentityResolutionQueueConfig() platformclientv2.Identityresolutionqueueconfig {
	resolveIdentities := true
	return platformclientv2.Identityresolutionqueueconfig{
		CallOnBehalfOfQueue: &platformclientv2.Outboundqueueidentityresolutionconfig{
			ResolveIdentities: &resolveIdentities,
		},
	}
}

func buildCallOnBehalfOfQueueConfig(blocks []interface{}) (*platformclientv2.Outboundqueueidentityresolutionconfig, error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("call_on_behalf_of_queue block is required")
	}

	block, ok := blocks[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("call_on_behalf_of_queue must be a block")
	}

	resolveIdentities := block["resolve_identities"].(bool)
	config := &platformclientv2.Outboundqueueidentityresolutionconfig{
		ResolveIdentities: &resolveIdentities,
	}

	if divisionId, ok := block["division_id"].(string); ok && !isUnassignedDivisionId(divisionId) {
		config.Division = &platformclientv2.Writablestarrabledivision{
			Id: platformclientv2.String(divisionId),
		}
	}

	return config, nil
}

func flattenCallOnBehalfOfQueue(config *platformclientv2.Outboundqueueidentityresolutionconfig) []interface{} {
	if config == nil {
		return []interface{}{
			map[string]interface{}{
				"resolve_identities": true,
			},
		}
	}

	result := map[string]interface{}{}
	if config.ResolveIdentities != nil {
		result["resolve_identities"] = *config.ResolveIdentities
	}
	if config.Division != nil && config.Division.Id != nil {
		divisionId := *config.Division.Id
		if !isUnassignedDivisionId(divisionId) {
			result["division_id"] = divisionId
		}
	}

	return []interface{}{result}
}

func isDefaultIdentityResolutionConfig(config *platformclientv2.Identityresolutionqueueconfig) bool {
	if config == nil || config.CallOnBehalfOfQueue == nil {
		return true
	}

	inner := config.CallOnBehalfOfQueue
	if inner.ResolveIdentities == nil || !*inner.ResolveIdentities {
		return false
	}

	if inner.Division != nil && inner.Division.Id != nil && !isUnassignedDivisionId(*inner.Division.Id) {
		return false
	}

	return true
}

// isUnassignedDivisionId reports whether division_id is the unassigned (STAR) division
// sentinel (omitted in config/state, or explicit "*"). In contacts-service, "*" serializes
// a null/STAR division id — not "all divisions" and not "parent resource's division".
func isUnassignedDivisionId(v string) bool {
	return v == "" || v == "*"
}

// suppressUnassignedDivisionIdDiff treats omitted division_id and "*" as equivalent
// (both mean the unassigned / STAR division). Read omits "*" from state, so explicit
// config "*" would otherwise show a perpetual plan diff.
func suppressUnassignedDivisionIdDiff(_, old, new string, _ *schema.ResourceData) bool {
	return isUnassignedDivisionId(old) && isUnassignedDivisionId(new)
}
