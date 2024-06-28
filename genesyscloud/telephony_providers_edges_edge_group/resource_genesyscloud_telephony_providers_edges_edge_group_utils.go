package telephony_providers_edges_edge_group

import (
	"fmt"
	"strings"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func buildSdkTrunkBases(d *schema.ResourceData) *[]platformclientv2.Trunkbase {
	returnValue := make([]platformclientv2.Trunkbase, 0)

	if ids, ok := d.GetOk("phone_trunk_base_ids"); ok {
		phoneTrunkBaseIds := lists.SetToStringList(ids.(*schema.Set))
		for _, trunkBaseId := range *phoneTrunkBaseIds {
			id := trunkBaseId
			returnValue = append(returnValue, platformclientv2.Trunkbase{
				Id: &id,
			})
		}
	}

	return &returnValue
}

func GenerateEdgeGroupResourceWithCustomAttrs(
	edgeGroupRes,
	name,
	description string,
	managed,
	hybrid bool,
	otherAttrs ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_telephony_providers_edges_edge_group" "%s" {
		name = "%s"
		description = "%s"
		managed = "%v"
		hybrid = "%v"
		%s
	}
	`, edgeGroupRes, name, description, managed, hybrid, strings.Join(otherAttrs, "\n"))
}

func GeneratePhoneTrunkBaseIds(userIDs ...string) string {
	return fmt.Sprintf(`phone_trunk_base_ids = [%s]
	`, strings.Join(userIDs, ","))
}

func flattenPhoneTrunkBases(trunkBases []platformclientv2.Trunkbase) *schema.Set {
	interfaceList := make([]interface{}, len(trunkBases))
	for i, v := range trunkBases {
		interfaceList[i] = *v.Id
	}
	return schema.NewSet(schema.HashString, interfaceList)
}
