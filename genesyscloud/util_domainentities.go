package genesyscloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v67/platformclientv2"
)

func buildSdkDomainEntityRef(d *schema.ResourceData, idAttr string) *platformclientv2.Domainentityref {
	idVal := d.Get(idAttr).(string)
	if idVal == "" {
		return nil
	}
	return &platformclientv2.Domainentityref{Id: &idVal}
}

func buildSdkDomainEntityRefArr(d *schema.ResourceData, idAttr string) *[]platformclientv2.Domainentityref {
	if ids, ok := d.GetOk(idAttr); ok && ids != nil {
		strList := setToStringList(ids.(*schema.Set))
		if strList != nil {
			domainEntityRefs := make([]platformclientv2.Domainentityref, len(*strList))
			for i, id := range *strList {
				tempId := id
				domainEntityRefs[i] = platformclientv2.Domainentityref{Id: &tempId}
			}
			return &domainEntityRefs
		}
	}
	return nil
}

func sdkDomainEntityRefArrToSet(entityRefs []platformclientv2.Domainentityref) *schema.Set {
	interfaceList := make([]interface{}, len(entityRefs))
	for i, v := range entityRefs {
		interfaceList[i] = *v.Id
	}
	return schema.NewSet(schema.HashString, interfaceList)
}
