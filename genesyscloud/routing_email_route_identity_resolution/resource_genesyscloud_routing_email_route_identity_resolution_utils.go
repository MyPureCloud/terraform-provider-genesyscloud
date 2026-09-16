package routing_email_route_identity_resolution

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

func buildIdentityResolutionRoutingEmailRouteConfig(d *schema.ResourceData) (platformclientv2.Routeidentityresolutionconfig, error) {
	resolveIdentities := d.Get("resolve_identities").(bool)
	config := platformclientv2.Routeidentityresolutionconfig{
		ResolveIdentities: &resolveIdentities,
	}

	if divisionId, ok := d.Get("division_id").(string); ok && !isUnassignedDivisionId(divisionId) {
		config.Division = &platformclientv2.Writablestarrabledivision{
			Id: platformclientv2.String(divisionId),
		}
	}

	return config, nil
}


func buildDefaultIdentityResolutionRoutingEmailRouteConfig() platformclientv2.Routeidentityresolutionconfig {
	resolveIdentities := true
	return platformclientv2.Routeidentityresolutionconfig{
		ResolveIdentities: &resolveIdentities,
	}
}

func isDefaultIdentityResolutionConfig(config *platformclientv2.Routeidentityresolutionconfig) bool {
	if config == nil {
		return true
	}

	if config.ResolveIdentities == nil || !*config.ResolveIdentities {
		return false
	}

	if config.Division != nil && config.Division.Id != nil && !isUnassignedDivisionId(*config.Division.Id) {
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
