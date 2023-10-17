package outbound_contact_list

import (
	"fmt"
	"strconv"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	obAttemptLimit "terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
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
		automaticTimeZoneMapping = falseValue
		attemptLimitResourceID   = "attempt-limit"
		attemptLimitDataSourceID = "attempt-limit-data"
		attemptLimitName         = "Test Attempt Limit " + uuid.NewString()

		nameUpdated                      = "Test Contact List " + uuid.NewString()
		automaticTimeZoneMappingUpdated  = trueValue
		zipCodeColumnName                = "Zipcode"
		columnNamesUpdated               = append(columnNames, strconv.Quote(zipCodeColumnName))
		previewModeColumnNameUpdated     = "Home"
		previewModeAcceptedValuesUpdated = []string{strconv.Quote(previewModeColumnName), strconv.Quote(previewModeColumnNameUpdated)}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundContactList(
					resourceId,
					name,
					nullValue,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					nullValue,
					nullValue,
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
					generateEmailColumnsBlock(
						"Work",
						"work",
						nullValue,
					),
					generateEmailColumnsBlock(
						"Personal",
						"personal",
						nullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.#", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.#", "4"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Cell"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Home"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Work"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_column_name", previewModeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMapping),
					gcloud.TestDefaultHomeDivision("genesyscloud_outbound_contact_list."+resourceId),
				),
			},
			// Update
			{
				Config: GenerateOutboundContactList(
					resourceId,
					name,
					nullValue,
					strconv.Quote(previewModeColumnName),
					previewModeAcceptedValuesUpdated,
					columnNames,
					automaticTimeZoneMapping,
					nullValue,
					nullValue,
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
					generateEmailColumnsBlock(
						"Work",
						"work",
						nullValue,
					),
					generateEmailColumnsBlock(
						"Personal",
						"personal",
						nullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.#", "0"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.#", "4"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Cell"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Home"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Work"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_column_name", previewModeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMapping),
					gcloud.TestDefaultHomeDivision("genesyscloud_outbound_contact_list."+resourceId),
				),
			},
			{
				// Update (forcenew)
				Config: GenerateOutboundContactList(
					resourceId,
					nameUpdated,
					nullValue,
					strconv.Quote(previewModeColumnNameUpdated),
					previewModeAcceptedValuesUpdated,
					columnNames,
					automaticTimeZoneMapping,
					nullValue,
					nullValue,
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
					generateEmailColumnsBlock(
						"Work",
						"work",
						nullValue,
					),
					generateEmailColumnsBlock(
						"Personal",
						"personal",
						nullValue,
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
						nullValue,             // min
						nullValue,             // max
						"5",                   // maxLength
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.#", "4"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Cell"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Home"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Work"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.type", "personal"),

					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.#", "2"),

					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.0.column_data_type", "TEXT"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.0.min", "1"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.0.max", "11"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.0.max_length", "10"),

					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.1.column_name", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.1.column_data_type", "TEXT"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.1.max_length", "5"),

					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_column_name", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMapping),
					gcloud.TestDefaultHomeDivision("genesyscloud_outbound_contact_list."+resourceId),
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
						nullValue,
					),
					GeneratePhoneColumnsBlock(
						"Home",
						"home",
						nullValue,
					),
					generateEmailColumnsBlock(
						"Work",
						"work",
						strconv.Quote(zipCodeColumnName),
					),
					generateEmailColumnsBlock(
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
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.#", "5"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Cell"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Home"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Work"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", "Personal"),
					gcloud.ValidateStringInArray("genesyscloud_outbound_contact_list."+resourceId, "column_names", zipCodeColumnName),

					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "zip_code_column_name", zipCodeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.contactable_time_column", zipCodeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.contactable_time_column", zipCodeColumnName),

					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.#", "1"),

					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.0.column_data_type", "TEXT"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.0.min", "2"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.0.max", "12"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_data_type_specifications.0.max_length", "11"),

					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_column_name", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMappingUpdated),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_attempt_limit."+attemptLimitDataSourceID, "id",
						"genesyscloud_outbound_contact_list."+resourceId, "attempt_limit_id"),
					gcloud.TestDefaultHomeDivision("genesyscloud_outbound_contact_list."+resourceId),
				),
			},
			{
				ResourceName:      "genesyscloud_outbound_contact_list." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyContactListDestroyed,
	})
}

func generateEmailColumnsBlock(columnName, columnType, contactableTimeColumn string) string {
	return fmt.Sprintf(`
	email_columns {
		column_name             = "%s"
		type                    = "%s"
		contactable_time_column = %s
	}
`, columnName, columnType, contactableTimeColumn)
}

func testVerifyContactListDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_outbound_contact_list" {
			continue
		}
		contactList, resp, err := outboundAPI.GetOutboundContactlist(rs.Primary.ID, false, false)
		if contactList != nil {
			return fmt.Errorf("contact list (%s) still exists", rs.Primary.ID)
		} else if gcloud.IsStatus404(resp) {
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
