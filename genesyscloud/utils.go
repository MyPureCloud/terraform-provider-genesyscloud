package genesyscloud

import (
	"github.com/MyPureCloud/platform-client-sdk-go/platformclientv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getHomeDivisionID() (string, diag.Diagnostics) {
	authAPI := platformclientv2.NewAuthorizationApi()
	homeDiv, _, err := authAPI.GetAuthorizationDivisionsHome()
	if err != nil {
		return "", diag.Errorf("Failed to query home division: %s", err)
	}
	return *homeDiv.Id, nil
}

func buildSdkDomainEntityRef(d *schema.ResourceData, idAttr string) *platformclientv2.Domainentityref {
	idVal := d.Get(idAttr).(string)
	if idVal == "" {
		return nil
	}
	return &platformclientv2.Domainentityref{Id: &idVal}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// difference returns the elements in a that aren't in b
func sliceDifference(a, b []string) []string {
	var diff []string
	if len(a) == 0 {
		return diff
	}
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func stringListToSet(list []string) *schema.Set {
	interfaceList := make([]interface{}, len(list))
	for i, v := range list {
		interfaceList[i] = v
	}
	return schema.NewSet(schema.HashString, interfaceList)
}

func setToStringList(strSet *schema.Set) *[]string {
	interfaceList := strSet.List()
	strList := make([]string, len(interfaceList))
	for i, s := range interfaceList {
		strList[i] = s.(string)
	}
	return &strList
}
