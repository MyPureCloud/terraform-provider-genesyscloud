package architect_ivr

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/validators"
)

// resourceArchitectIvrV1 is the schema at version 1, where dnis was a TypeSet.
// Used only for the v1 -> v2 state upgrader ImpliedType.
func resourceArchitectIvrV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dnis": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString, ValidateDiagFunc: validators.ValidatePhoneNumber},
			},
		},
	}
}

// stateUpgraderArchitectIvrV1ToV2 converts dnis from TypeSet to TypeList.
// This addresses the SDK gRPC layer forcing planned TypeSet values into state on apply errors
// (DEVTOOLING-1533 / DEVTOOLING-1881). A set is stored as a JSON array, so the numbers are kept as-is
// and only sorted so the resulting list order is deterministic.
func stateUpgraderArchitectIvrV1ToV2(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	rawDnis, ok := rawState["dnis"].([]interface{})
	if !ok {
		return rawState, nil
	}

	dnis := make([]string, 0, len(rawDnis))
	for _, v := range rawDnis {
		if number, isString := v.(string); isString {
			dnis = append(dnis, number)
		}
	}
	sort.Strings(dnis)

	upgraded := make([]interface{}, len(dnis))
	for i, number := range dnis {
		upgraded[i] = number
	}
	rawState["dnis"] = upgraded
	return rawState, nil
}
