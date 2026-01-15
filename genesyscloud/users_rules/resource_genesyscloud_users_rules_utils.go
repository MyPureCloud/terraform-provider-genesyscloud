package users_rules

import (
	"fmt"
	"strconv"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

type UsersRulesValueStruct struct {
	ContextId string
	Ids       []string
}

type UsersRulesGroupItemStruct struct {
	Id        string
	Operator  string
	Container string
	Values    []UsersRulesValueStruct
}

type UsersRulesCriteriaStruct struct {
	Id       string
	Operator string
	Group    []UsersRulesGroupItemStruct
}

func buildSdkUsersRulesValue(resourceUsersRulesValueList []interface{}) *[]platformclientv2.Usersrulesvalue {
	usersRulesValues := make([]platformclientv2.Usersrulesvalue, 0)

	for _, usersRulesValue := range resourceUsersRulesValueList {
		usersRulesValueMap := usersRulesValue.(map[string]interface{})

		var sdkUsersRulesValue platformclientv2.Usersrulesvalue
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUsersRulesValue.ContextId, usersRulesValueMap, "context_id")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkUsersRulesValue.Ids, usersRulesValueMap, "ids")

		usersRulesValues = append(usersRulesValues, sdkUsersRulesValue)
	}

	return &usersRulesValues
}

func buildSdkUsersRulesGroupItem(resourceUsersRulesGroupItemList []interface{}) *[]platformclientv2.Usersrulesgroupitem {
	usersRulesGroupItems := make([]platformclientv2.Usersrulesgroupitem, 0)

	for _, usersRulesGroupItem := range resourceUsersRulesGroupItemList {
		usersRulesGroupItemMap := usersRulesGroupItem.(map[string]interface{})

		var sdkUsersRulesGroupItem platformclientv2.Usersrulesgroupitem
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUsersRulesGroupItem.Id, usersRulesGroupItemMap, "id")
		sdkUsersRulesGroupItem.Operator = platformclientv2.String(usersRulesGroupItemMap["operator"].(string))
		sdkUsersRulesGroupItem.Container = platformclientv2.String(usersRulesGroupItemMap["container"].(string))
		sdkUsersRulesGroupItem.Values = buildSdkUsersRulesValue(usersRulesGroupItemMap["values"].([]interface{}))

		usersRulesGroupItems = append(usersRulesGroupItems, sdkUsersRulesGroupItem)
	}

	return &usersRulesGroupItems
}

func buildSdkUsersRulesCriteria(resourceUsersRulesCriteriaList []interface{}) *[]platformclientv2.Usersrulescriteria {
	if resourceUsersRulesCriteriaList == nil {
		return nil
	}

	usersRulesCriteriaList := make([]platformclientv2.Usersrulescriteria, 0)

	for _, usersRulesCriteria := range resourceUsersRulesCriteriaList {
		usersRulesCriteriaMap := usersRulesCriteria.(map[string]interface{})

		var sdkUsersRulesCriteria platformclientv2.Usersrulescriteria
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUsersRulesCriteria.Id, usersRulesCriteriaMap, "id")
		sdkUsersRulesCriteria.Operator = platformclientv2.String(usersRulesCriteriaMap["operator"].(string))
		sdkUsersRulesCriteria.Group = buildSdkUsersRulesGroupItem(usersRulesCriteriaMap["group"].([]interface{}))

		usersRulesCriteriaList = append(usersRulesCriteriaList, sdkUsersRulesCriteria)
	}

	return &usersRulesCriteriaList
}

func flattenUsersRulesValue(usersRulesValues *[]platformclientv2.Usersrulesvalue) []interface{} {
	if usersRulesValues == nil {
		return nil
	}

	usersRulesValuesList := make([]interface{}, 0)
	for _, usersRulesValue := range *usersRulesValues {
		usersRulesValueMap := make(map[string]interface{})

		if usersRulesValue.ContextId != nil {
			usersRulesValueMap["context_id"] = *usersRulesValue.ContextId
		}
		if usersRulesValue.Ids != nil {
			usersRulesValueMap["ids"] = *usersRulesValue.Ids
		}

		usersRulesValuesList = append(usersRulesValuesList, usersRulesValueMap)
	}

	return usersRulesValuesList
}

func flattenUsersRulesGroupItem(usersRulesGroupItem *[]platformclientv2.Usersrulesgroupitem) []interface{} {
	if usersRulesGroupItem == nil {
		return nil
	}

	usersRulesGroupItemList := make([]interface{}, 0)
	for _, usersRulesGroupItem := range *usersRulesGroupItem {
		usersRulesGroupItemMap := make(map[string]interface{})

		if usersRulesGroupItem.Id != nil {
			usersRulesGroupItemMap["id"] = *usersRulesGroupItem.Id
		}
		if usersRulesGroupItem.Operator != nil {
			usersRulesGroupItemMap["operator"] = *usersRulesGroupItem.Operator
		}
		if usersRulesGroupItem.Container != nil {
			usersRulesGroupItemMap["container"] = *usersRulesGroupItem.Container
		}
		if usersRulesGroupItem.Values != nil {
			usersRulesGroupItemMap["values"] = flattenUsersRulesValue(usersRulesGroupItem.Values)
		}

		usersRulesGroupItemList = append(usersRulesGroupItemList, usersRulesGroupItemMap)
	}

	return usersRulesGroupItemList
}

func flattenUsersRulesCriteria(usersRulesCriteria *[]platformclientv2.Usersrulescriteria) []interface{} {
	if usersRulesCriteria == nil {
		return nil
	}

	usersRulesCriteriaList := make([]interface{}, 0)
	for _, usersRulesCriteria := range *usersRulesCriteria {
		usersRulesCriteriaMap := make(map[string]interface{})

		if usersRulesCriteria.Id != nil {
			usersRulesCriteriaMap["id"] = *usersRulesCriteria.Id
		}
		if usersRulesCriteria.Operator != nil {
			usersRulesCriteriaMap["operator"] = *usersRulesCriteria.Operator
		}
		if usersRulesCriteria.Group != nil {
			usersRulesCriteriaMap["group"] = flattenUsersRulesGroupItem(usersRulesCriteria.Group)
		}

		usersRulesCriteriaList = append(usersRulesCriteriaList, usersRulesCriteriaMap)
	}

	return usersRulesCriteriaList
}

func GenerateUsersRulesResource(
	resourceLabel string,
	name string,
	description string,
	ruleType string,
	criteria []UsersRulesCriteriaStruct,
) string {
	return fmt.Sprintf(`resource "genesyscloud_users_rules" "%s" {
		name = "%s"
		description = "%s"
		type = "%s"
		%s
	}
	`, resourceLabel, name, description, ruleType, generateUsersRulesCriteria(criteria))
}

func generateUsersRulesCriteria(usersRulesCriteriaList []UsersRulesCriteriaStruct) string {
	if usersRulesCriteriaList == nil {
		return ""
	}

	var usersRulesCriteriaString string
	for _, usersRulesCriteria := range usersRulesCriteriaList {
		usersRulesCriteriaString += fmt.Sprintf(`
			criteria {
				id = "%s"
				operator = "%s"
				%s
			}
		`, usersRulesCriteria.Id, usersRulesCriteria.Operator, generateUsersRulesGroupItem(usersRulesCriteria.Group))
	}

	return usersRulesCriteriaString
}

func generateUsersRulesGroupItem(usersRulesGroupItems []UsersRulesGroupItemStruct) string {
	if usersRulesGroupItems == nil {
		return ""
	}

	var usersRulesGroupItemString string
	for _, usersRulesGroupItem := range usersRulesGroupItems {
		usersRulesGroupItemString += fmt.Sprintf(`
			group {
				id = "%s"
				operator = "%s"
				container = "%s"
				%s
			}
		`, usersRulesGroupItem.Id, usersRulesGroupItem.Operator, usersRulesGroupItem.Container, generateUsersRulesValue(usersRulesGroupItem.Values))
	}

	return usersRulesGroupItemString
}

func generateUsersRulesValue(usersRulesValues []UsersRulesValueStruct) string {
	if usersRulesValues == nil {
		return ""
	}

	var usersRulesValueString string
	for _, usersRulesValue := range usersRulesValues {
		idsString := ""

		for i, id := range usersRulesValue.Ids {
			if i > 0 {
				idsString += ", "
			}

			idsString += strconv.Quote(id)
		}

		usersRulesValueString += fmt.Sprintf(`
		values {
			context_id = "%s"
			ids = [%s]
		}
		`, usersRulesValue.ContextId, idsString)
	}

	return usersRulesValueString
}
