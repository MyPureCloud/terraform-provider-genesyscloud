package util

import (
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func BuildSdkDomainEntityRef(d *schema.ResourceData, idAttr string) *platformclientv2.Domainentityref {
	idVal := d.Get(idAttr).(string)
	if idVal == "" {
		return nil
	}
	return &platformclientv2.Domainentityref{Id: &idVal}
}

func BuildSdkDomainEntityRefArr(d *schema.ResourceData, idAttr string) *[]platformclientv2.Domainentityref {
	if ids, ok := d.GetOk(idAttr); ok && ids != nil {
		if setIds, ok := ids.(*schema.Set); ok {
			strList := lists.SetToStringList(setIds)
			if setIds != nil {
				domainEntityRefs := make([]platformclientv2.Domainentityref, len(*strList))
				for i, id := range *strList {
					tempId := id
					domainEntityRefs[i] = platformclientv2.Domainentityref{Id: &tempId}
				}
				return &domainEntityRefs
			}
		} else {
			strList := lists.InterfaceListToStrings(ids.([]interface{}))
			if len(strList) > 0 {
				domainEntityRefs := make([]platformclientv2.Domainentityref, len(strList))
				for i, id := range strList {
					tempId := id
					domainEntityRefs[i] = platformclientv2.Domainentityref{Id: &tempId}
				}
				return &domainEntityRefs
			}
		}
	}
	return nil
}

func BuildSdkDomainEntityRefArrFromArr(ids []interface{}) *[]platformclientv2.Domainentityref {
	var domainEntityRefs []platformclientv2.Domainentityref
	for _, id := range ids {
		if idStr, ok := id.(string); ok {
			domainEntityRefs = append(domainEntityRefs, platformclientv2.Domainentityref{Id: &idStr})
		}
	}
	return &domainEntityRefs
}

func SdkDomainEntityRefArrToSet(entityRefs []platformclientv2.Domainentityref) *schema.Set {
	interfaceList := make([]interface{}, len(entityRefs))
	for i, v := range entityRefs {
		interfaceList[i] = *v.Id
	}
	return schema.NewSet(schema.HashString, interfaceList)
}

func SdkDomainEntityRefArrToList(entityRefs []platformclientv2.Domainentityref) []interface{} {
	interfaceList := make([]interface{}, len(entityRefs))
	for i, v := range entityRefs {
		interfaceList[i] = *v.Id
	}
	return interfaceList
}
