package outbound_digitalruleset

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	obContactList "terraform-provider-genesyscloud/genesyscloud/outbound_contact_list"
)

/*
The resource_genesyscloud_outbound_digitalruleset_test.go contains all of the test cases for running the resource
tests for outbound_digitalruleset.
*/

func TestAccResourceOutboundDigitalruleset(t *testing.T) {
	t.Parallel()
	var (
		name1             = "Terraform Test Digital RuleSet1"
		resourceId        = "digital-rule-set"
		version           = "0"
		ruleName          = "RuleWork"
		ruleOrder         = "0"
		ruleCategory      = "PreContact"
		contactColumnName = "Work"
		columnOperator    = "Equals"
		columnValue       = "\"XYZ\""
		columnValueType   = "String"

		updatePropertiesWork = "Work"
		updateOption         = "Set"

		contactListResourceId1    = "contact-list-1"
		contactListName1          = "Test Contact List " + uuid.NewString()
		previewModeColumnName     = ""
		previewModeAcceptedValues = []string{}
		columnNames               = []string{strconv.Quote("Cell"), strconv.Quote("Work")}
		automaticTimeZoneMapping  = util.FalseValue
	)

	contactListResourceGenerate := obContactList.GenerateOutboundContactList(
		contactListResourceId1,
		contactListName1,
		util.NullValue,
		strconv.Quote(previewModeColumnName),
		previewModeAcceptedValues,
		columnNames,
		automaticTimeZoneMapping,
		util.NullValue,
		util.NullValue,
		obContactList.GeneratePhoneColumnsBlock(
			"Cell",
			"cell",
			strconv.Quote("Cell"),
		),
		obContactList.GenerateEmailColumnsBlock(
			"Work",
			"Work",
			strconv.Quote("Work"),
		),
		obContactList.GeneratePhoneColumnsDataTypeSpecBlock(
			strconv.Quote("Cell"),
			strconv.Quote("TEXT"),
			"1",
			"11",
			"10",
		),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{

				Config: contactListResourceGenerate +
					GenerateOutboundDigitalRuleSetResource(
						resourceId,
						name1,
						version,
						"genesyscloud_outbound_contact_list."+contactListResourceId1+".id",
						GenerateDigitalRules(
							ruleName,
							ruleOrder,
							ruleCategory,
							GenerateDigitalRuleSetConditions(
								util.TrueValue,
								GenerateContactColumnConditionSettings(
									contactColumnName,
									columnOperator,
									columnValue,
									columnValueType,
								),
							),
							GenerateDigitalRuleSetActions(
								GenerateUpdateContactColumnActionSettings(
									//util.GenerateJsonEncodedProperties(util.GenerateJsonProperty(updatePropertiesWork, updatePropertiesWork)),
									updateOption,
								),
								GenerateDoNotSendActionSettings(),
							),
						),
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "name", name1),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_digitalruleset."+resourceId, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceId1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "version", version),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.order", ruleOrder),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.category", ruleCategory),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.0.inverted", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.0.contact_column_condition_settings.0.column_name", contactColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.0.contact_column_condition_settings.0.operator", columnOperator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.0.contact_column_condition_settings.0.value", columnValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.0.contact_column_condition_settings.0.value_type", columnValueType),
					util.ValidateValueInJsonAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.actions.0.update_contact_column_action_settings.0.properties", updatePropertiesWork, updatePropertiesWork),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.actions.0.update_contact_column_action_settings.0.update_option", updateOption),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_outbound_digitalruleset." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundDigitalrulesetDestroyed,
	})
}

// func GenerateSetSmsPhoneNumberActionSettings() string {
// 	return fmt.Sprintf(`
// 		set_sms_phone_number_action_settings {
// 			sender_sms_phone_number = %s
// 		}
// 	`)
// }

// func GenerateSetContentTemplateActionSettings() string {
// 	return fmt.Sprintf(`
// 		set_content_template_action_settings {
// 			sms_content_template_id = "%s"
// 			email_content_template_id = "%s"
// 		}
// 	`)
// }

// func GenerateMarkContactAddressUncontactableActionSettings() string {
// 	return fmt.Sprintf(`
// 		mark_contact_address_uncontactable_action_settings = %s
// 	`)
// }

// func GenerateMarkContactUncontactableActionSettings() string {
// 	return fmt.Sprintf(`
// 		mark_contact_uncontactable_action_settings {
// 			media_types = [%s]
// 		}
// 	`)
// }

// func GenerateAppendToDncActionSettings() string {
// 	return fmt.Sprintf(`
// 		append_to_dnc_action_settings {
// 			expire = %s
// 			expiration_duration = %s
// 			list_type = %s
// 		}
// 	`)
// }

func GenerateDoNotSendActionSettings() string {
	return fmt.Sprintf(`
		do_not_send_action_settings = {}
	`)
}

func GenerateUpdateContactColumnActionSettings(
	//properties string,
	updateOption string,
) string {
	return fmt.Sprintf(`update_contact_column_action_settings {
		update_option = "%s"
		properties = {
			Cell	=	"Cell"
		}
	}
	`, updateOption)
}

func GenerateDigitalRuleSetActions(nestedBlocks ...string) string {
	return fmt.Sprintf(`
		actions {
			%s
		}
	`, strings.Join(nestedBlocks, "\n"))
}

// func GenerateDataActionContactColumnToDataActionFieldMappings() string {
// 	return fmt.Sprintf(`
// 		contact_column_to_data_action_field_mappings {
// 			contact_column_name = %s
// 			data_action_field = %s
// 		}
// 	`)
// }

// func GenerateDataActionConditionSettingsPredicates() string {
// 	return fmt.Sprintf(`
// 		predicates {
// 			output_field = %s
// 			output_operator = %s
// 			comparison_value = %s
// 			inverted = %s
// 			output_field_missing_resolution = %s
// 		}
// 	`)
// }

// func GenerateDataActionConditionSettings() string {
// 	return fmt.Sprintf(`
// 	data_action_condition_settings {
// 		data_action_id = %s
// 		contact_id_field = %s
// 		data_not_found_resolution = %s
// 		%s
// 	}
// 	`)
// }

// func GenerateLastResultOverallConditionSettings() string {
// 	return fmt.Sprintf(`
// 	last_result_overall_condition_settings {
// 		email_wrapup_codes = [%s]
// 		sms_wrapup_codes = [%s]
// 	}
// 	`)
// }

// func GenerateLastResultByColumnConditionSettings() string {
// 	return fmt.Sprintf(`
// 	last_result_by_column_condition_settings {
// 		email_column_name = %s
// 		email_wrapup_codes = [%s]
// 		sms_column_name = %s
// 		sms_wrapup_codes = [%s]
// 	}
// 	`)
// }

// func GenerateLastAttemptOverallConditionSettings() string {
// 	return fmt.Sprintf(`
// 	last_attempt_overall_condition_settings {
// 		media_types = [%s]
// 		operator = %s
// 		value = %s
// 	}
// 	`)
// }

// func GenerateLastAttemptByColumnConditionSettings() string {
// 	return fmt.Sprintf(`
// 	last_attempt_by_column_condition_settings {
// 		email_column_name = %s
// 		sms_column_name = %s
// 		operator = %s
// 		value = %s
// 	}
// 	`)
// }

// func GenerateContactAddressTypeConditionSettings() string {
// 	return fmt.Sprintf(`
// 	contact_address_type_condition_settings {
// 		operator = %s
// 		value = %s
// 	}
// 	`)
// }

// func GenerateContactAddressConditionSettings() string {
// 	return fmt.Sprintf(`
// 	contact_address_condition_settings {
// 		operator = %s
// 		value = %s
// 	}
// 	`)
// }

func GenerateContactColumnConditionSettings(
	columnName string,
	operator string,
	value string,
	valueType string,
) string {
	return fmt.Sprintf(`
	contact_column_condition_settings {
		column_name = "%s"
		operator = "%s"
		value = %s
		value_type = "%s"
	}
	`, columnName, operator, value, valueType)
}

func GenerateDigitalRuleSetConditions(
	inverted string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
		conditions {
			inverted = %s
			%s
		}
	`, inverted, strings.Join(nestedBlocks, "\n"))
}

func GenerateDigitalRules(
	name string,
	order string,
	category string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
		rules {
			name = "%s"
			order = %s
			category = "%s"
			%s
		}
	`, name, order, category, strings.Join(nestedBlocks, "\n"))
}

func GenerateOutboundDigitalRuleSetResource(
	resourceId string,
	name string,
	version string,
	contactListId string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
	resource "genesyscloud_outbound_digitalruleset" "%s" {
	name = "%s"
	version = %s
	contact_list_id = %s
	%s
	}
	`, resourceId, name, version, contactListId, strings.Join(nestedBlocks, "\n"))
}

func testVerifyOutboundDigitalrulesetDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_digitalruleset" {
			continue
		}
		ruleset, resp, err := outboundAPI.GetOutboundDigitalruleset(rs.Primary.ID)
		if ruleset != nil {
			return fmt.Errorf("digital ruleset (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// ruleset not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All rulesets destroyed
	return nil
}
