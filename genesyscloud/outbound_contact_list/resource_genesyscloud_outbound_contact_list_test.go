package outbound_contact_list

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	testrunner "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/testrunner"

	obAttemptLimit "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/outbound_attempt_limit"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func TestAccResourceOutboundContactListBasicWithoutContacts(t *testing.T) {

	t.Parallel()
	var (
		resourceLabel             = "contact-list"
		name                      = "Test Contact List " + uuid.NewString()
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
					resourceLabel,
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
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "0"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+resourceLabel, "contacts_file_content_hash"),
					provider.TestDefaultHomeDivision(ResourceType+"."+resourceLabel),
				),
			},
			// Update
			{
				Config: GenerateOutboundContactList(
					resourceLabel,
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
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "0"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+resourceLabel, "contacts_file_content_hash"),
					provider.TestDefaultHomeDivision(ResourceType+"."+resourceLabel),
				),
			},
			{
				// Update (forcenew)
				Config: GenerateOutboundContactList(
					resourceLabel,
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
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "0"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+resourceLabel, "contacts_file_content_hash"),
					provider.TestDefaultHomeDivision(ResourceType+"."+resourceLabel),
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
				) + `data "genesyscloud_auth_division_home" "home" {}` + GenerateOutboundContactList(
					resourceLabel,
					nameUpdated,
					"data.genesyscloud_auth_division_home.home.id",
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
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "0"),
					resource.TestCheckNoResourceAttr(ResourceType+"."+resourceLabel, "contacts_file_content_hash"),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_attempt_limit."+attemptLimitDataSourceLabel, "id",
						ResourceType+"."+resourceLabel, "attempt_limit_id"),
					provider.TestDefaultHomeDivision(ResourceType+"."+resourceLabel),
				),
			},
			{
				ResourceName:      ResourceType + "." + resourceLabel,
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
		if rs.Type != ResourceType {
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

func TestAccResourceOutboundContactListWithContacts(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel = "contact-list-with-contacts"
		name          = "Test Contact List " + uuid.NewString()
		columnNames   = []string{
			strconv.Quote("id"),
			strconv.Quote("firstName"),
			strconv.Quote("lastName"),
			strconv.Quote("phone"),
			strconv.Quote("email"),
		}
		// Create mock CSV file contact data
		testContactsContentWithTwoRecords = `id,firstName,lastName,phone,email
100,John,Doe,+13175555555,john.doe@example.com
101,Jane,Smith,+13175555556,jane.smith@example.com`

		testContactsContentWithThreeRecords = testContactsContentWithTwoRecords + `
102,Bob,Johnson,+13175555557,bob.johnson@example.com`

		testContactsContentWithFourRecords = testContactsContentWithThreeRecords + `
103,Charlie,Brown,000000000000,charlie.brown@example.com`

		testContactsContentWithFiveRecords = testContactsContentWithFourRecords + `
104,Jenny,Doe,+15558675309,jenny@jenny.com`
	)
	// Create a temporary file for the contacts
	tmpFile, err := os.CreateTemp(testrunner.GetTestDataPath(), "contacts*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write the test contacts to the temp file
	if err := os.WriteFile(tmpFile.Name(), []byte(testContactsContentWithTwoRecords), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a second temporary file for the contacts
	tmpFile2, err := os.CreateTemp(testrunner.GetTestDataPath(), "contacts*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile2.Name())

	// Write the test contacts to the temp file
	if err := os.WriteFile(tmpFile2.Name(), []byte(testContactsContentWithFourRecords), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a third temporary file for the contacts
	tmpFile3, err := os.CreateTemp(testrunner.GetTestDataPath(), "contacts*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile3.Name())

	// Write the test contacts to the temp file
	if err := os.WriteFile(tmpFile3.Name(), []byte(testContactsContentWithFiveRecords), 0644); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundContactList(
					resourceLabel,
					name,
					util.NullValue, // division_id
					util.NullValue, // preview_mode_column_name
					[]string{},     // preview_mode_accepted_values
					columnNames,
					util.FalseValue, // automatic_time_zone_mapping
					util.NullValue,  // zipcode_column_names
					util.NullValue,  // attempt_limit_id
					GeneratePhoneColumnsBlock(
						"phone",
						"phone",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"email",
						"email",
						util.NullValue,
					),
					GenerateContactsFile(
						tmpFile.Name(), // contacts_filepath
						"id",           // contacts_id_name
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "5"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "id"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "firstName"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "lastName"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "phone"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "email"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.column_name", "phone"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.type", "phone"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.column_name", "email"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.type", "email"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "2"),
				),
			},
			// Test updating the contents of the contacts file
			{
				PreConfig: func() {
					// Update CSV file content
					err := os.WriteFile(tmpFile.Name(), []byte(testContactsContentWithThreeRecords), 0644)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: GenerateOutboundContactList(
					resourceLabel,
					name,
					util.NullValue,
					util.NullValue,
					[]string{},
					columnNames,
					util.FalseValue,
					util.NullValue,
					util.NullValue,
					GeneratePhoneColumnsBlock(
						"phone",
						"phone",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"email",
						"email",
						util.NullValue,
					),
					GenerateContactsFile(
						tmpFile.Name(), // Same file, but in real usage this could be a different file
						"id",           // contacts_id_name
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "5"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "3"),
				),
			},
			// Test when the contacts file path changes
			{
				Config: GenerateOutboundContactList(
					resourceLabel,
					name,
					util.NullValue,
					util.NullValue,
					[]string{},
					columnNames,
					util.FalseValue,
					util.NullValue,
					util.NullValue,
					GeneratePhoneColumnsBlock(
						"phone",
						"phone",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"email",
						"email",
						util.NullValue,
					),
					GenerateContactsFile(
						tmpFile2.Name(), // Same file, but in real usage this could be a different file
						"id",            // contacts_id_name
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "5"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "4"),
				),
			},
			// Test a blank file of contacts
			{
				PreConfig: func() {
					// Remove file content
					err := os.WriteFile(tmpFile2.Name(), []byte(""), 0644)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: GenerateOutboundContactList(
					resourceLabel,
					name,
					util.NullValue,
					util.NullValue,
					[]string{},
					columnNames,
					util.FalseValue,
					util.NullValue,
					util.NullValue,
					GeneratePhoneColumnsBlock(
						"phone",
						"phone",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"email",
						"email",
						util.NullValue,
					),
					GenerateContactsFile(
						tmpFile2.Name(), // Same file, but in real usage this could be a different file
						"id",            // contacts_id_name
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "5"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "0"),
				),
			},
			// Test that contacts can be re-uploaded
			{
				Config: GenerateOutboundContactList(
					resourceLabel,
					name,
					util.NullValue,
					util.NullValue,
					[]string{},
					columnNames,
					util.FalseValue,
					util.NullValue,
					util.NullValue,
					GeneratePhoneColumnsBlock(
						"phone",
						"phone",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"email",
						"email",
						util.NullValue,
					),
					GenerateContactsFile(
						tmpFile.Name(), // Same file, but in real usage this could be a different file
						"id",           // contacts_id_name
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "5"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "3"),
				),
			},
			// Test when the contacts file is non-existant
			{
				PreConfig: func() {
					// Remove file of contacts
					err := os.Remove(tmpFile.Name())
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: GenerateOutboundContactList(
					resourceLabel,
					name,
					util.NullValue,
					util.NullValue,
					[]string{},
					columnNames,
					util.FalseValue,
					util.NullValue,
					util.NullValue,
					GeneratePhoneColumnsBlock(
						"phone",
						"phone",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"email",
						"email",
						util.NullValue,
					),
					GenerateContactsFile(
						tmpFile.Name(), // Same file, but in real usage this could be a different file
						"id",           // contacts_id_name
					),
				),
				ExpectError: regexp.MustCompile("could not open.*no such file"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "5"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "3"),
				),
			},
			// Ensure we can re-upload the file after it was removed
			{
				Config: GenerateOutboundContactList(
					resourceLabel,
					name,
					util.NullValue, // division_id
					util.NullValue, // preview_mode_column_name
					[]string{},     // preview_mode_accepted_values
					columnNames,
					util.FalseValue, // automatic_time_zone_mapping
					util.NullValue,  // zipcode_column_names
					util.NullValue,  // attempt_limit_id
					GeneratePhoneColumnsBlock(
						"phone",
						"phone",
						util.NullValue,
					),
					GenerateEmailColumnsBlock(
						"email",
						"email",
						util.NullValue,
					),
					GenerateContactsFile(
						tmpFile3.Name(), // contacts_filepath
						"id",            // contacts_id_name
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_names.#", "5"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "id"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "firstName"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "lastName"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "phone"),
					util.ValidateStringInArray(ResourceType+"."+resourceLabel, "column_names", "email"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.column_name", "phone"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "phone_columns.0.type", "phone"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.column_name", "email"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "email_columns.0.type", "email"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "contacts_record_count", "5"),
				),
			},
			{
				// Import
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"contacts_file_content_hash",
					"contacts_filepath",
					"contacts_id_name",
				},
			},
		},
		CheckDestroy: testVerifyContactListDestroyed,
	})
}

// You might also want to add a test for invalid contact data
func TestAccResourceOutboundContactListWithInvalidContacts(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel = "contact-list-invalid"
		name          = "Test Contact List " + uuid.NewString()
		columnNames   = []string{
			strconv.Quote("Id"),
			strconv.Quote("Cell"),
			strconv.Quote("Home"),
			strconv.Quote("Name"),
		}
		// Create invalid CSV data (missing required columns)
		invalidContactsContent = `WrongColumn,Name
+13175551234,John Doe
+13175559876,Jane Smith`
	)

	// Create a temporary file for the invalid contacts
	tmpFile, err := os.CreateTemp(testrunner.GetTestDataPath(), "invalid_contacts*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if err := os.WriteFile(tmpFile.Name(), []byte(invalidContactsContent), 0644); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateOutboundContactList(
					resourceLabel,
					name,
					util.NullValue, // division_id
					util.NullValue,
					[]string{},
					columnNames,
					util.FalseValue,
					util.NullValue,
					util.NullValue,
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
					GenerateContactsFile(
						tmpFile.Name(), // contacts_filepath
						"Id",           // contacts_id_name
					),
				),
				ExpectError: regexp.MustCompile(`failed to validate contacts file: CSV file is missing required columns`), // Adjust the error message based on your actual implementation
			},
		},
	})
}
