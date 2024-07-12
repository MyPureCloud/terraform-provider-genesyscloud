package outbound_contact_list

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
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceOutboundContactListBasic(t *testing.T) {

	t.Parallel()
	var (
		resourceId                = "contact-list"
		name                      = "Test Contact List " + uuid.NewString()
		previewModeColumnName     = "Cell"
		previewModeAcceptedValues = []string{strconv.Quote(previewModeColumnName)}
		columnNames               = []string{
			strconv.Quote("Cell"),
			strconv.Quote("Home"),
			strconv.Quote("Work"),
			strconv.Quote("Personal"),
		}
		automaticTimeZoneMapping = util.FalseValue
		attemptLimitResourceID   = "attempt-limit"
		attemptLimitDataSourceID = "attempt-limit-data"
		attemptLimitName         = "Test Attempt Limit " + uuid.NewString()

		nameUpdated                      = "Test Contact List " + uuid.NewString()
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
				Config: GenerateOutboundContactList(
					resourceId,
					name,
					util.NullValue,
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
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "name", name),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.#", "0"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_names.#", "4"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Cell"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Home"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Work"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Personal"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_column_name", previewModeColumnName),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMapping),
					provider.TestDefaultHomeDivision(resourceName+"."+resourceId),
				),
			},
			// Update
			{
				Config: GenerateOutboundContactList(
					resourceId,
					name,
					util.NullValue,
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
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "name", name),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.#", "0"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_names.#", "4"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Cell"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Home"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Work"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Personal"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_column_name", previewModeColumnName),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMapping),
					provider.TestDefaultHomeDivision(resourceName+"."+resourceId),
				),
			},
			{
				// Update (forcenew)
				Config: GenerateOutboundContactList(
					resourceId,
					nameUpdated,
					util.NullValue,
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
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_names.#", "4"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Cell"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Home"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Work"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Personal"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.0.type", "personal"),

					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.#", "2"),

					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.0.column_data_type", "TEXT"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.0.min", "1"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.0.max", "11"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.0.max_length", "10"),

					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.1.column_name", "Home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.1.column_data_type", "TEXT"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.1.max_length", "5"),

					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_column_name", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMapping),
					provider.TestDefaultHomeDivision(resourceName+"."+resourceId),
				),
			},
			{
				Config: obAttemptLimit.GenerateAttemptLimitResource(
					attemptLimitResourceID,
					attemptLimitName,
					"5",
					"5",
					"America/Chicago",
					"TODAY",
				) + obAttemptLimit.GenerateOutboundAttemptLimitDataSource(
					attemptLimitDataSourceID,
					attemptLimitName,
					"genesyscloud_outbound_attempt_limit."+attemptLimitResourceID,
				) + `data "genesyscloud_auth_division_home" "home" {}` + GenerateOutboundContactList(
					resourceId,
					nameUpdated,
					"data.genesyscloud_auth_division_home.home.id",
					strconv.Quote(previewModeColumnNameUpdated),
					previewModeAcceptedValuesUpdated,
					columnNamesUpdated,
					automaticTimeZoneMappingUpdated,
					strconv.Quote(zipCodeColumnName),
					"genesyscloud_outbound_attempt_limit."+attemptLimitResourceID+".id",
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
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_names.#", "5"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Cell"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Home"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Work"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", "Personal"),
					util.ValidateStringInArray(resourceName+"."+resourceId, "column_names", zipCodeColumnName),

					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "zip_code_column_name", zipCodeColumnName),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.0.contactable_time_column", zipCodeColumnName),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "email_columns.1.contactable_time_column", zipCodeColumnName),

					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.#", "1"),

					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.0.column_name", "Cell"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.0.column_data_type", "TEXT"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.0.min", "2"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.0.max", "12"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_data_type_specifications.0.max_length", "11"),

					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_column_name", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMappingUpdated),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_attempt_limit."+attemptLimitDataSourceID, "id",
						resourceName+"."+resourceId, "attempt_limit_id"),
					provider.TestDefaultHomeDivision(resourceName+"."+resourceId),
				),
			},
			{
				ResourceName:      resourceName + "." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyContactListDestroyed,
	})
}

func testVerifyContactListDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != resourceName {
			continue
		}
		contactList, resp, err := outboundAPI.GetOutboundContactlist(rs.Primary.ID, false, false)
		if contactList != nil {
			return fmt.Errorf("contact list (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Contact list not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All contact lists destroyed
	return nil
}
