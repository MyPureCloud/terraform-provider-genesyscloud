package group

import (
	"fmt"
	"log"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// 'number' and 'extension' conflict with eachother. However, one must be set.
// This function validates that the user has satisfied these conditions
func validateAddressesMap(m map[string]interface{}) error {
	number, _ := m["number"].(string)
	extension, _ := m["extension"].(string)

	if (number != "" && extension != "") ||
		(number == "" && extension == "") {
		return fmt.Errorf("Either 'number' or 'extension' must be set inside addresses, but both cannot be set.")
	}

	return nil
}

func flattenGroupAddresses(d *schema.ResourceData, addresses *[]platformclientv2.Groupcontact) []interface{} {
	addressSlice := make([]interface{}, 0)
	for _, address := range *addresses {
		if address.MediaType != nil {
			if *address.MediaType == groupPhoneType {
				phoneNumber := make(map[string]interface{})

				// Strip off any parentheses from phone numbers
				if address.Address != nil {
					phoneNumber["number"], _ = util.FormatAsE164Number(strings.Trim(*address.Address, "()"))
				}

				resourcedata.SetMapValueIfNotNil(phoneNumber, "extension", address.Extension)
				resourcedata.SetMapValueIfNotNil(phoneNumber, "type", address.VarType)

				// Sometimes the number or extension is only returned in Display
				if address.Address == nil &&
					address.Extension == nil &&
					address.Display != nil {
					setExtensionOrNumberBasedOnDisplay(d, phoneNumber, &address)
				}

				addressSlice = append(addressSlice, phoneNumber)
			} else {
				log.Printf("Unknown address media type %s", *address.MediaType)
			}
		}
	}
	return addressSlice
}

/**
*  The api can sometimes return only the display which holds the value
*  that was stored in either `address` or `extension`
*  This function establishes which field was set in the schema data (`extension` or `address`)
*  and then sets that field in the map to the value that came back in `display`
 */
func setExtensionOrNumberBasedOnDisplay(d *schema.ResourceData, addressMap map[string]interface{}, address *platformclientv2.Groupcontact) {
	display := strings.Trim(*address.Display, "()")
	schemaAddresses := d.Get("addresses").([]interface{})
	for _, a := range schemaAddresses {
		currentAddress, ok := a.(map[string]interface{})
		if !ok {
			continue
		}
		addressType, _ := currentAddress["type"].(string)
		if addressType != *address.VarType {
			continue
		}
		if ext, _ := currentAddress["extension"].(string); ext != "" {
			addressMap["extension"] = display
		} else if number, _ := currentAddress["number"].(string); number != "" {
			addressMap["number"], _ = util.FormatAsE164Number(display)
		}
	}
}

func flattenGroupOwners(owners *[]platformclientv2.User) []interface{} {
	interfaceList := make([]interface{}, len(*owners))
	for i, v := range *owners {
		interfaceList[i] = *v.Id
	}
	return interfaceList
}

func buildSdkGroupAddresses(d *schema.ResourceData) (*[]platformclientv2.Groupcontact, error) {
	if addressSlice, ok := d.Get("addresses").([]interface{}); ok && len(addressSlice) > 0 {
		sdkContacts := make([]platformclientv2.Groupcontact, len(addressSlice))
		for i, configPhone := range addressSlice {
			phoneMap := configPhone.(map[string]interface{})
			phoneType := phoneMap["type"].(string)
			contact := platformclientv2.Groupcontact{
				VarType:   &phoneType,
				MediaType: &groupPhoneType, // Only option is PHONE
			}

			if err := validateAddressesMap(phoneMap); err != nil {
				return nil, err
			}

			if phoneNum, ok := phoneMap["number"].(string); ok && phoneNum != "" {
				contact.Address = &phoneNum
			}

			if phoneExt := phoneMap["extension"].(string); ok && phoneExt != "" {
				contact.Extension = &phoneExt
			}

			sdkContacts[i] = contact
		}
		return &sdkContacts, nil
	}
	return nil, nil
}

func GenerateBasicGroupResource(resourceID string, name string, nestedBlocks ...string) string {
	return GenerateGroupResource(resourceID, name, util.NullValue, util.NullValue, util.NullValue, util.TrueValue, nestedBlocks...)
}

func GenerateGroupResource(
	resourceID string,
	name string,
	desc string,
	groupType string,
	visibility string,
	rulesVisible string,
	nestedBlocks ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_group" "%s" {
		name = "%s"
		description = %s
		type = %s
		visibility = %s
		rules_visible = %s
        %s
	}
	`, resourceID, name, desc, groupType, visibility, rulesVisible, strings.Join(nestedBlocks, "\n"))
}

func generateGroupAddress(number string, phoneType string, extension string) string {
	return fmt.Sprintf(`addresses {
		number = %s
		type = "%s"
		extension = %s
	}
	`, number, phoneType, extension)
}

func GenerateGroupOwners(userIDs ...string) string {
	return fmt.Sprintf(`owner_ids = [%s]
	`, strings.Join(userIDs, ","))
}

func generateGroupMembers(userIDs ...string) string {
	return fmt.Sprintf(`member_ids = [%s]
	`, strings.Join(userIDs, ","))
}
