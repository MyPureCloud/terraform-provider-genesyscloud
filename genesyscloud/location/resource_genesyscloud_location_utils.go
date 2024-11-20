package location

import (
	"fmt"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	lists "terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
	"github.com/nyaruka/phonenumbers"
)

func buildSdkLocationPath(d *schema.ResourceData) *[]string {
	path := []string{}
	if pathConfig, ok := d.GetOk("path"); ok {
		path = lists.InterfaceListToStrings(pathConfig.([]interface{}))
	}
	return &path
}

func buildSdkLocationEmergencyNumber(d *schema.ResourceData) *platformclientv2.Locationemergencynumber {
	if numberConfig := d.Get("emergency_number"); numberConfig != nil {
		if numberList := numberConfig.([]interface{}); len(numberList) > 0 {
			settingsMap := numberList[0].(map[string]interface{})

			number := settingsMap["number"].(string)
			typeStr := settingsMap["type"].(string)
			return &platformclientv2.Locationemergencynumber{
				Number:  &number,
				VarType: &typeStr,
			}
		}
	}
	return &platformclientv2.Locationemergencynumber{}
}

func buildSdkLocationAddress(d *schema.ResourceData) *platformclientv2.Locationaddress {
	if addressConfig := d.Get("address"); addressConfig != nil {
		if addrList := addressConfig.([]interface{}); len(addrList) > 0 {
			addrMap := addrList[0].(map[string]interface{})

			city := addrMap["city"].(string)
			country := addrMap["country"].(string)
			zip := addrMap["zip_code"].(string)
			street1 := addrMap["street1"].(string)
			address := platformclientv2.Locationaddress{
				City:    &city,
				Country: &country,
				Zipcode: &zip,
				Street1: &street1,
			}
			// Optional values
			if state, ok := addrMap["state"]; ok {
				stateStr := state.(string)
				address.State = &stateStr
			}
			if street2, ok := addrMap["street2"]; ok {
				street2Str := street2.(string)
				address.Street2 = &street2Str
			}
			return &address
		}
	}
	return &platformclientv2.Locationaddress{}
}

func flattenLocationEmergencyNumber(numberConfig *platformclientv2.Locationemergencynumber) []interface{} {
	if numberConfig == nil {
		return nil
	}
	numberSettings := make(map[string]interface{})
	if numberConfig.Number != nil {
		utilE164 := util.NewUtilE164Service()
		numberSettings["number"] = utilE164.FormatAsCalculatedE164Number(*numberConfig.Number)
	}
	if numberConfig.VarType != nil {
		numberSettings["type"] = *numberConfig.VarType
	}
	return []interface{}{numberSettings}
}

func flattenLocationAddress(addrConfig *platformclientv2.Locationaddress) []interface{} {
	if addrConfig == nil {
		return nil
	}
	addrSettings := make(map[string]interface{})
	if addrConfig.City != nil {
		addrSettings["city"] = *addrConfig.City
	}
	if addrConfig.Country != nil {
		addrSettings["country"] = *addrConfig.Country
	}
	if addrConfig.State != nil {
		addrSettings["state"] = *addrConfig.State
	}
	if addrConfig.Street1 != nil {
		addrSettings["street1"] = *addrConfig.Street1
	}
	if addrConfig.Street2 != nil {
		addrSettings["street2"] = *addrConfig.Street2
	}
	if addrConfig.Zipcode != nil {
		addrSettings["zip_code"] = *addrConfig.Zipcode
	}
	return []interface{}{addrSettings}
}

func comparePhoneNumbers(_, old, new string, _ *schema.ResourceData) bool {
	oldNum, err := phonenumbers.Parse(old, "US")
	if err != nil {
		return old == new
	}

	newNum, err := phonenumbers.Parse(new, "US")
	if err != nil {
		return old == new
	}
	return phonenumbers.IsNumberMatchWithNumbers(oldNum, newNum) == phonenumbers.EXACT_MATCH
}

func GenerateLocationResourceBasic(
	resourceLabel,
	name string,
	nestedBlocks ...string) string {
	return GenerateLocationResource(resourceLabel, name, "", []string{})
}

func GenerateLocationResource(
	resourceLabel,
	name,
	notes string,
	paths []string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_location" "%s" {
		name = "%s"
        notes = "%s"
        path = [%s]
        %s
	}
	`, resourceLabel, name, notes, strings.Join(paths, ","), strings.Join(nestedBlocks, "\n"))
}

func GenerateLocationEmergencyNum(number, typeStr string) string {
	return fmt.Sprintf(`emergency_number {
		number = "%s"
        type = %s
	}
	`, number, typeStr)
}

func GenerateLocationAddress(street1, city, state, country, zip string) string {
	return fmt.Sprintf(`address {
		street1  = "%s"
		city     = "%s"
		state    = "%s"
		country  = "%s"
		zip_code = "%s"
	}
	`, street1, city, state, country, zip)
}
