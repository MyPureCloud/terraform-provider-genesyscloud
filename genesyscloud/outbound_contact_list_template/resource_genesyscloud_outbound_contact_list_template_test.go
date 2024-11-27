package outbound_contact_list_template

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	obAttemptLimit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

func TestAccResourceOutboundContactListTemplateBasic(t *testing.T) {

	t.Parallel()
	var (
		resourceLabel             = "contact-list-template"
		name                      = "Test Contact List Template" + uuid.NewString()
		previewModeColumnName     = "Cell"
		previewModeAcceptedValues = []string{strconv.Quote(previewModeColumnName)}
		columnNames               = []string{
			strconv.Quote("Cell"),
			strconv.Quote("Home"),
			strconv.Quote("Work"),
			strconv.Quote("Personal"),
		}
		automaticTimeZoneMapping    = util.FalseValue
		attemptLimitResourceLabel   = "attempt-limit"
		attemptLimitDataSourceLabel = "attempt-limit-data"
		attemptLimitName            = "Test Attempt Limit " + uuid.NewString()

		nameUpdated                      = "Test Contact List Template" + uuid.NewString()
		automaticTimeZoneMappingUpdated  = util.TrueValue
		zipCodeColumnName                = "Zipcode"
		columnNamesUpdated               = append(columnNames, strconv.Quote(zipCodeColumnName))
		previewModeColumnNameUpdated     = "Home"
		previewModeAcceptedValuesUpdated = []string{strconv.Quote(previewModeColumnName), strconv.Quote(previewModeColumnNameUpdated)}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundContactListTemplate(
					resourceLabel,
					name,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					util.NullValue,
					util.NullValue,
					GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						strconv.Quote("Cell"),
					),
					GeneratePhoneColumnsBlock(
						"Home",
						"home",
						strconv.Quote("Home"),
					),
					GenerateEmailColumnsBlock(
						"Work",
						"work",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"Personal",
						"personal",
						util.NullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "4"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Cell"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Home"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Work"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Personal"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_column_name", previewModeColumnName),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "automatic_time_zone_mapping", automaticTimeZoneMapping),
				),
			},
			// Update
			{
				Config: GenerateOutboundContactListTemplate(
					resourceLabel,
					name,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValuesUpdated,
					columnNames,
					automaticTimeZoneMapping,
					util.NullValue,
					util.NullValue,
					GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						strconv.Quote("Cell"),
					),
					GeneratePhoneColumnsBlock(
						"Home",
						"home",
						strconv.Quote("Home"),
					),
					GenerateEmailColumnsBlock(
						"Work",
						"work",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"Personal",
						"personal",
						util.NullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "4"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Cell"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Home"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Work"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Personal"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_column_name", previewModeColumnName),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "automatic_time_zone_mapping", automaticTimeZoneMapping),
				),
			},
			{
				// Update (forcenew)
				Config: GenerateOutboundContactListTemplate(
					resourceLabel,
					nameUpdated,
					strconv.Quote(previewModeColumnNameUpdated),
					previewModeAcceptedValuesUpdated,
					columnNames,
					automaticTimeZoneMapping,
					util.NullValue,
					util.NullValue,
					GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						strconv.Quote("Cell"),
					),
					GeneratePhoneColumnsBlock(
						"Home",
						"home",
						strconv.Quote("Home"),
					),
					GenerateEmailColumnsBlock(
						"Work",
						"work",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"Personal",
						"personal",
						util.NullValue,
					),
					GeneratePhoneColumnsDataTypeSpecBlock(
						strconv.Quote("Cell"), // columnName
						strconv.Quote("TEXT"), // columnDataType
						"1",                   // min
						"11",                  // max
						"10",                  // maxLength
					),
					GeneratePhoneColumnsDataTypeSpecBlock(
						strconv.Quote("Home"), // columnName
						strconv.Quote("TEXT"), // columnDataType
						util.NullValue,        // min
						util.NullValue,        // max
						"5",                   // maxLength
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", nameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "4"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Cell"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Home"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Work"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Personal"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.type", "personal"),

					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.#", "2"),

					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.0.column_data_type", "TEXT"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.0.min", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.0.max", "11"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.0.max_length", "10"),

					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.1.column_name", "Home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.1.column_data_type", "TEXT"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.1.max_length", "5"),

					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_column_name", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "automatic_time_zone_mapping", automaticTimeZoneMapping),
				),
			},
			{
				Config: obAttemptLimit.GenerateAttemptLimitResource(
					attemptLimitResourceLabel,
					attemptLimitName,
					"5",
					"5",
					"America/Chicago",
					"TODAY",
				) + obAttemptLimit.GenerateOutboundAttemptLimitDataSource(
					attemptLimitDataSourceLabel,
					attemptLimitName,
					"genesyscloud_outbound_attempt_limit."+attemptLimitResourceLabel,
				) + GenerateOutboundContactListTemplate(
					resourceLabel,
					nameUpdated,
					strconv.Quote(previewModeColumnNameUpdated),
					previewModeAcceptedValuesUpdated,
					columnNamesUpdated,
					automaticTimeZoneMappingUpdated,
					strconv.Quote(zipCodeColumnName),
					"genesyscloud_outbound_attempt_limit."+attemptLimitResourceLabel+".id",
					GeneratePhoneColumnsBlock(
						"Cell",
						"cell",
						util.NullValue,
					),
					GeneratePhoneColumnsBlock(
						"Home",
						"home",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"Work",
						"work",
						strconv.Quote(zipCodeColumnName),
					),
					GenerateEmailColumnsBlock(
						"Personal",
						"personal",
						strconv.Quote(zipCodeColumnName),
					),
					GeneratePhoneColumnsDataTypeSpecBlock(
						strconv.Quote("Cell"), // columnName
						strconv.Quote("TEXT"), // columnDataType
						"2",                   // min
						"12",                  // max
						"11",                  // maxLength
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", nameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "5"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Cell"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Home"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Work"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "Personal"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", zipCodeColumnName),

					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "zip_code_column_name", zipCodeColumnName),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.contactable_time_column", zipCodeColumnName),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.1.contactable_time_column", zipCodeColumnName),

					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.#", "1"),

					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.0.column_data_type", "TEXT"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.0.min", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.0.max", "12"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_data_type_specifications.0.max_length", "11"),

					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_column_name", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "automatic_time_zone_mapping", automaticTimeZoneMappingUpdated),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_attempt_limit."+attemptLimitDataSourceLabel, "id",
						ResourceType+"."+resourceLabel, "attempt_limit_id"),
				),
			},
			{
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyContactListTemplateDestroyed,
	})
}

func testVerifyContactListTemplateDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}
		contactList, resp, err := outboundAPI.GetOutboundContactlisttemplate(rs.Primary.ID)
		if contactList != nil {
			return fmt.Errorf("contact list template (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Contact list template not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All contact lists template destroyed
	return nil
}
