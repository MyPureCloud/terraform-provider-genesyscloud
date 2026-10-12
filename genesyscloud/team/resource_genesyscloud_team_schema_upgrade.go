package team

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
)

// resourceTeamSchemaV1 is the schema at version 1 where member_ids was TypeSet.
// Used only for the v1 → v2 state upgrader ImpliedType.
func resourceTeamSchemaV1() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"member_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

// upgradeTeamStateV1 converts member_ids from TypeSet to sorted TypeList.
// This addresses DEVTOOLING-1493 where the SDK gRPC layer was forcing planned TypeSet
// values into state on apply errors (e.g. UserGroupLimitExceeded), causing state corruption.
func upgradeTeamStateV1(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	rawState["member_ids"] = memberIdsValueToSortedList(rawState["member_ids"])
	return rawState, nil
}

func memberIdsValueToSortedList(value interface{}) []interface{} {
	switch memberIds := value.(type) {
	case *schema.Set:
		return sortedStringInterfaceList(lists.InterfaceListToStrings(memberIds.List()))
	case []interface{}:
		return sortedStringInterfaceList(lists.InterfaceListToStrings(memberIds))
	case []string:
		return sortedStringInterfaceList(memberIds)
	default:
		return []interface{}{}
	}
}

func sortedStringInterfaceList(values []string) []interface{} {
	sort.Strings(values)
	return lists.StringListToInterfaceList(values)
}
