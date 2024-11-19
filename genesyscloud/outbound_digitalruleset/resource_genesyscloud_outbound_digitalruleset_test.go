package outbound_digitalruleset

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"

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
		resourceId        = "digital-rule-set"
		ruleName          = "RuleWork"
		name2             = "DigitalRuleSet-" + uuid.NewString()
		ruleOrder         = "0"
		ruleCategory      = "PreContact"
		contactColumnName = "Work"
		columnOperator    = "Equals"
		columnValue       = "XYZ"
		columnValueType   = "String"

		updatePropertiesWork = "Work"
		updateOption         = "Set"

		contactListResourceId1    = "contact-list-1"
		contactListName1          = "Test Contact List " + uuid.NewString()
		previewModeColumnName     = ""
		previewModeAcceptedValues = []string{}
		columnNames               = []string{strconv.Quote("Cell"), strconv.Quote("Work")}
		automaticTimeZoneMapping  = util.FalseValue

		lastAttemptOverallOperator = "Before"
		lastAttemptOverallValue    = "P-1DT-1H-1M"
		outboundMessageSent        = "OUTBOUND-MESSAGE-SENT"
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
						name2,
						"genesyscloud_outbound_contact_list."+contactListResourceId1+".id",
						GenerateDigitalRules(
							ruleName,
							ruleOrder,
							ruleCategory,
							GenerateDigitalRuleSetConditions(
								GenerateInvertedConditionAttr(util.TrueValue),
								GenerateContactColumnConditionSettings(
									contactColumnName,
									columnOperator,
									columnValue,
									columnValueType,
								),
							),
							GenerateDigitalRuleSetConditions(
								GenerateLastAttemptOverallConditionSettings(
									[]string{strconv.Quote("Email")},
									lastAttemptOverallOperator,
									lastAttemptOverallValue,
								),
							),
							GenerateDigitalRuleSetConditions(
								GenerateLastResultByColumnConditionSettings(
									contactColumnName,
									[]string{strconv.Quote(outboundMessageSent)},
									"",
									[]string{},
								),
							),
							GenerateDigitalRuleSetConditions(
								GenerateLastResultOverallConditionSettings(
									[]string{strconv.Quote(outboundMessageSent)},
									[]string{},
								),
							),
							GenerateDigitalRuleSetActions(
								GenerateUpdateContactColumnActionSettings(
									updateOption,
									GeneratePropertiesForUpdateContactColumnSettings(updatePropertiesWork, updatePropertiesWork),
								),
							),
							GenerateDigitalRuleSetActions(
								GenerateDoNotSendActionSettings(),
							),
						),
					),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "name", name2),
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_digitalruleset."+resourceId, "contact_list_id", "genesyscloud_outbound_contact_list."+contactListResourceId1, "id"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.name", ruleName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.order", ruleOrder),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.category", ruleCategory),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.0.inverted", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.0.contact_column_condition_settings.0.column_name", contactColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.0.contact_column_condition_settings.0.operator", columnOperator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.0.contact_column_condition_settings.0.value", columnValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.0.contact_column_condition_settings.0.value_type", columnValueType),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.1.last_attempt_overall_condition_settings.0.media_types.0", "Email"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.1.last_attempt_overall_condition_settings.0.operator", lastAttemptOverallOperator),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.1.last_attempt_overall_condition_settings.0.value", lastAttemptOverallValue),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.2.last_result_by_column_condition_settings.0.email_column_name", contactColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.2.last_result_by_column_condition_settings.0.email_wrapup_codes.0", outboundMessageSent),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.2.last_result_by_column_condition_settings.0.sms_column_name", ""),
					resource.TestCheckResourceAttr("genesyscloud_outbound_digitalruleset."+resourceId, "rules.0.conditions.3.last_result_overall_condition_settings.0.email_wrapup_codes.0", outboundMessageSent),
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

func GenerateSetSmsPhoneNumberActionSettings(
	senderSmsPhone string,
) string {
	return fmt.Sprintf(`
		set_sms_phone_number_action_settings {
			sender_sms_phone_number = "%s"
		}
	`, senderSmsPhone)
}

func GenerateSetContentTemplateActionSettings(
	smsContentId string,
	emailContentId string,
) string {
	return fmt.Sprintf(`
		set_content_template_action_settings {
			sms_content_template_id = "%s"
			email_content_template_id = "%s"
		}
	`, smsContentId, emailContentId)
}

func GenerateMarkContactAddressUncontactableActionSettings() string {
	return fmt.Sprintf(`
		mark_contact_address_uncontactable_action_settings = jsonencode({})
	`)
}

func GenerateMarkContactUncontactableActionSettings(
	mediaTypes []string,
) string {
	return fmt.Sprintf(`
		mark_contact_uncontactable_action_settings {
			media_types = [%s]
		}
	`, strings.Join(mediaTypes, ","))
}

func GenerateAppendToDncActionSettings(
	expire string,
	expirationDuration string,
	listType string,
) string {
	return fmt.Sprintf(`
		append_to_dnc_action_settings {
			expire = %s
			expiration_duration = "%s"
			list_type = "%s"
		}
	`, expire, expirationDuration, listType)
}

func GenerateDoNotSendActionSettings() string {
	return fmt.Sprintf(`
		do_not_send_action_settings = jsonencode({})
	`)
}

func GeneratePropertiesForUpdateContactColumnSettings(
	propType string,
	propValue string) string {
	return "properties = " + util.GenerateJsonEncodedProperties(util.GenerateJsonProperty(propType, strconv.Quote(propValue)))
}

func GenerateUpdateContactColumnActionSettings(
	updateOption string,
	properties ...string,
) string {
	return fmt.Sprintf(`update_contact_column_action_settings {
		update_option = "%s"
		%s
	}
	`, updateOption, strings.Join(properties, "\n"))
}

func GenerateDigitalRuleSetActions(nestedBlocks ...string) string {
	return fmt.Sprintf(`
		actions {
			%s
		}
	`, strings.Join(nestedBlocks, "\n"))
}

func GenerateDataActionContactColumnToDataActionFieldMappings(
	contactColumnName string,
	dataActionField string,
) string {
	return fmt.Sprintf(`
		contact_column_to_data_action_field_mappings {
			contact_column_name = "%s"
			data_action_field = "%s"
		}
	`, contactColumnName, dataActionField)
}

func GenerateDataActionConditionSettingsPredicates(
	outputField string,
	outputOperator string,
	comparisonValue string,
	inverted string,
	outputFieldMissingResolution string,
) string {
	return fmt.Sprintf(`
		predicates {
			output_field = "%s"
			output_operator = "%s"
			comparison_value = "%s"
			inverted = %s
			output_field_missing_resolution = "%s"
		}
	`, outputField, outputOperator, comparisonValue, inverted, outputFieldMissingResolution)
}

func GenerateDataActionConditionSettings(
	dataActionId string,
	contactIdField string,
	dataNotFound string,
	predicatesBlock ...string,
) string {
	return fmt.Sprintf(`
	data_action_condition_settings {
		data_action_id = "%s"
		contact_id_field = "%s"
		data_not_found_resolution = "%s"
		%s
	}
	`, dataActionId, contactIdField, dataNotFound, strings.Join(predicatesBlock, ","))
}

func GenerateLastResultOverallConditionSettings(
	emailCodes []string,
	smsCodes []string,
) string {
	return fmt.Sprintf(`
	last_result_overall_condition_settings {
		email_wrapup_codes = [%s]
		sms_wrapup_codes = [%s]
	}
	`, strings.Join(emailCodes, ","), strings.Join(smsCodes, ","))
}

func GenerateLastResultByColumnConditionSettings(
	emailColumnName string,
	emailCodes []string,
	smsColumnName string,
	smsCodes []string,
) string {
	return fmt.Sprintf(`
	last_result_by_column_condition_settings {
		email_column_name = "%s"
		email_wrapup_codes = [%s]
		sms_column_name = "%s"
		sms_wrapup_codes = [%s]
	}
	`, emailColumnName, strings.Join(emailCodes, ","), smsColumnName, strings.Join(smsCodes, ","))
}

func GenerateLastAttemptOverallConditionSettings(
	mediaTypes []string,
	operator string,
	value string,
) string {
	return fmt.Sprintf(`
	last_attempt_overall_condition_settings {
		media_types = [%s]
		operator = "%s"
		value = "%s"
	}
	`, strings.Join(mediaTypes, ","), operator, value)
}

func GenerateLastAttemptByColumnConditionSettings(
	emailColumnName string,
	smsColumnName string,
	operator string,
	value string,
) string {
	return fmt.Sprintf(`
	last_attempt_by_column_condition_settings {
		email_column_name = "%s"
		sms_column_name = "%s"
		operator = "%s"
		value = "%s"
	}
	`, emailColumnName, smsColumnName, operator, value)
}

func GenerateContactAddressTypeConditionSettings(
	operator string,
	value string,
) string {
	return fmt.Sprintf(`
	contact_address_type_condition_settings {
		operator = "%s"
		value = "%s"
	}
	`, operator, value)
}

func GenerateContactAddressConditionSettings(
	operator string,
	value string,
) string {
	return fmt.Sprintf(`
	contact_address_condition_settings {
		operator = "%s"
		value = "%s"
	}
	`, operator, value)
}

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
		value = "%s"
		value_type = "%s"
	}
	`, columnName, operator, value, valueType)
}

func GenerateDigitalRuleSetConditions(
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
		conditions {
			%s
		}
	`, strings.Join(nestedBlocks, "\n"))
}

func GenerateInvertedConditionAttr(
	inverted string,
) string {
	return fmt.Sprintf(`
		inverted = %s`, inverted)
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

func GenerateDigitalRuleSetVersion(
	version string,
) string {
	return fmt.Sprintf(`
		version = %s`, version)
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
