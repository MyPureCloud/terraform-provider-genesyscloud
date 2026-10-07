package routing_queue

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

// resourceRoutingQueueV2 is the schema at version 2 where members and wrapup_codes were TypeSet.
// Used only for the v2 → v3 state upgrader ImpliedType.
func resourceRoutingQueueV2() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"members": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     queueMemberResource,
			},
			"wrapup_codes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// stateUpgraderRoutingQueueV2ToV3 converts members and wrapup_codes from TypeSet to TypeList.
// This addresses DEVTOOLING-1533 where the SDK gRPC layer was forcing planned TypeSet
// values into state on apply errors (e.g. member already assigned via group), causing
// state corruption.
func stateUpgraderRoutingQueueV2ToV3(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	rawState["wrapup_codes"] = setOrListToSortedStringList(rawState["wrapup_codes"])
	rawState["members"] = membersSetOrListToList(rawState["members"])
	return rawState, nil
}

func setOrListToSortedStringList(value interface{}) []interface{} {
	switch v := value.(type) {
	case *schema.Set:
		return sortedStringInterfaceList(lists.InterfaceListToStrings(v.List()))
	case []interface{}:
		return sortedStringInterfaceList(lists.InterfaceListToStrings(v))
	case []string:
		return sortedStringInterfaceList(v)
	default:
		return []interface{}{}
	}
}

func membersSetOrListToList(value interface{}) []interface{} {
	switch v := value.(type) {
	case *schema.Set:
		return v.List()
	case []interface{}:
		return v
	default:
		return []interface{}{}
	}
}

func sortedStringInterfaceList(values []string) []interface{} {
	sort.Strings(values)
	return lists.StringListToInterfaceList(values)
}
