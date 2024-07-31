package outbound_contact_list_contact

import (
	"fmt"
	"strings"
	utillists "terraform-provider-genesyscloud/genesyscloud/util/lists"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// buildWritableContactFromResourceData used to build the request body for contact creation
func buildWritableContactFromResourceData(d *schema.ResourceData) platformclientv2.Writabledialercontact {
	contactListId := d.Get("contact_list_id").(string)
	callable := d.Get("callable").(bool)

	var contactRequest = platformclientv2.Writabledialercontact{
		ContactListId: &contactListId,
		Callable:      &callable,
	}

	if dataMap, ok := d.Get("data").(map[string]any); ok {
		stringMap := utillists.ConvertMapStringAnyToMapStringString(dataMap)
		contactRequest.Data = &stringMap
	}

	contactRequest.PhoneNumberStatus = buildPhoneNumberStatus(d)
	contactRequest.ContactableStatus = buildContactableStatus(d)
	return contactRequest
}

// buildDialerContactFromResourceData used to build the request body for contact updates
func buildDialerContactFromResourceData(d *schema.ResourceData) platformclientv2.Dialercontact {
	contactListId := d.Get("contact_list_id").(string)
	callable := d.Get("callable").(bool)
	var contactRequest = platformclientv2.Dialercontact{
		ContactListId: &contactListId,
		Callable:      &callable,
	}
	if dataMap, ok := d.Get("data").(map[string]any); ok {
		stringMap := utillists.ConvertMapStringAnyToMapStringString(dataMap)
		contactRequest.Data = &stringMap
	}
	contactRequest.PhoneNumberStatus = buildPhoneNumberStatus(d)
	contactRequest.ContactableStatus = buildContactableStatus(d)
	return contactRequest
}

func buildContactableStatus(d *schema.ResourceData) *map[string]platformclientv2.Contactablestatus {
	contactableStatus, ok := d.Get("contactable_status").(*schema.Set)
	if !ok {
		return nil
	}

	contactableStatusMap := make(map[string]platformclientv2.Contactablestatus)

	contactableStatusList := contactableStatus.List()
	for _, status := range contactableStatusList {
		currentStatusMap := status.(map[string]any)
		mediaType := currentStatusMap["media_type"].(string)
		contactable := currentStatusMap["contactable"].(bool)

		columnStatusMap := make(map[string]platformclientv2.Columnstatus)
		if columnStatus, ok := currentStatusMap["column_status"].(*schema.Set); ok {
			columnStatusList := columnStatus.List()
			for _, status := range columnStatusList {
				currentColumnStatusMap := status.(map[string]any)
				column := currentColumnStatusMap["column"].(string)
				columnContactable := currentColumnStatusMap["contactable"].(bool)
				columnStatusMap[column] = platformclientv2.Columnstatus{
					Contactable: &columnContactable,
				}
			}
		}
		contactableStatusMap[mediaType] = platformclientv2.Contactablestatus{
			Contactable:  &contactable,
			ColumnStatus: &columnStatusMap,
		}
	}

	return &contactableStatusMap
}

func buildPhoneNumberStatus(d *schema.ResourceData) *map[string]platformclientv2.Phonenumberstatus {
	phoneNumberStatus, ok := d.Get("phone_number_status").(*schema.Set)
	if !ok {
		return nil
	}

	phoneNumberStatusMap := make(map[string]platformclientv2.Phonenumberstatus)

	phoneNumberStatusList := phoneNumberStatus.List()
	for _, status := range phoneNumberStatusList {
		statusMap := status.(map[string]any)
		key := statusMap["key"].(string)
		callable, _ := statusMap["callable"].(bool)
		phoneNumberStatusMap[key] = platformclientv2.Phonenumberstatus{
			Callable: &callable,
		}
	}

	return &phoneNumberStatusMap
}

func flattenPhoneNumberStatus(phoneNumberStatus *map[string]platformclientv2.Phonenumberstatus) *schema.Set {
	pnsSet := schema.NewSet(schema.HashResource(phoneNumberStatusResource), []interface{}{})
	for k, v := range *phoneNumberStatus {
		pns := make(map[string]any)
		pns["key"] = k
		resourcedata.SetMapValueIfNotNil(pns, "callable", v.Callable)
		pnsSet.Add(pns)
	}
	return pnsSet
}

func flattenContactableStatus(contactableStatus *map[string]platformclientv2.Contactablestatus) *schema.Set {
	csSet := schema.NewSet(schema.HashResource(contactableStatusResource), []interface{}{})
	for k, v := range *contactableStatus {
		cs := make(map[string]any)
		cs["media_type"] = k
		cs["contactable"] = *v.Contactable
		if v.ColumnStatus != nil {
			cs["column_status"] = flattenColumnStatus(v.ColumnStatus)
		}
		csSet.Add(cs)
	}
	return csSet
}

func flattenColumnStatus(columnStatus *map[string]platformclientv2.Columnstatus) *schema.Set {
	if columnStatus == nil {
		return nil
	}
	csSet := schema.NewSet(schema.HashResource(columnStatusResource), []interface{}{})
	for k, v := range *columnStatus {
		cs := make(map[string]any)
		cs["column"] = k
		cs["contactable"] = *v.Contactable
		csSet.Add(cs)
	}
	return csSet
}

func GenerateOutboundContactListContact(
	resourceId,
	contactListId,
	callable,
	data string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`resource "%s" "%s" {
    contact_list_id = %s
    callable        = %s
    %s
    %s
}`, resourceName, resourceId, contactListId, callable, data, strings.Join(nestedBlocks, "\n"))
}

func GeneratePhoneNumberStatus(key, callable string) string {
	return fmt.Sprintf(`
	phone_number_status {
		key      = "%s"
        callable = %s
	}`, key, callable)
}

func GenerateContactableStatus(mediaType, contactable string, nestedBlocks ...string) string {
	return fmt.Sprintf(`
	contactable_status {
		media_type  = "%s"
		contactable = %s
		%s
	}`, mediaType, contactable, strings.Join(nestedBlocks, "\n"))
}

func GenerateColumnStatus(column, contactable string) string {
	return fmt.Sprintf(`
		column_status {
			column      = "%s"
			contactable = %s
		}`, column, contactable)
}
