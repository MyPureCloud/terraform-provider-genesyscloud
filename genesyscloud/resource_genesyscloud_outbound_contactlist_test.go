package genesyscloud

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v98/platformclientv2"
)

func TestAccResourceOutboundContactList(t *testing.T) {
	t.Parallel()
	var (
		resourceId                = "contact-list"
		name                      = "Test Contact List " + uuid.NewString()
		previewModeColumnName     = "Cell"
		previewModeAcceptedValues = []string{strconv.Quote(previewModeColumnName)}
		columnNames               = []string{strconv.Quote("Cell"), strconv.Quote("Home"), strconv.Quote("Work"), strconv.Quote("Personal")}
		automaticTimeZoneMapping  = falseValue
		attemptLimitResourceID    = "attempt-limit"
		attemptLimitDataSourceID  = "attempt-limit-data"
		attemptLimitName          = "Test Attempt Limit " + uuid.NewString()

		nameUpdated                      = "Test Contact List " + uuid.NewString()
		automaticTimeZoneMappingUpdated  = trueValue
		zipCodeColumnName                = "Zipcode"
		columnNamesUpdated               = append(columnNames, strconv.Quote(zipCodeColumnName))
		previewModeColumnNameUpdated     = "Home"
		previewModeAcceptedValuesUpdated = []string{strconv.Quote(previewModeColumnName), strconv.Quote(previewModeColumnNameUpdated)}
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: generateOutboundContactList(
					resourceId,
					name,
					"",
					previewModeColumnName,
					previewModeAcceptedValues,
					columnNames,
					automaticTimeZoneMapping,
					"",
					"",
					generatePhoneColumnsBlock(
						"Cell",
						"cell",
						"Cell",
					),
					generatePhoneColumnsBlock(
						"Home",
						"home",
						"Home",
					),
					generateEmailColumnsBlock(
						"Work",
						"work",
						"Work",
					),
					generateEmailColumnsBlock(
						"Personal",
						"personal",
						"Personal",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "name", name),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.0", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.1", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.2", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.3", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.contactable_time_column", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.contactable_time_column", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_column_name", previewModeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMapping),
					testDefaultHomeDivision("genesyscloud_outbound_contact_list."+resourceId),
				),
			},
			{
				// Update
				Config: generateOutboundContactList(
					resourceId,
					nameUpdated,
					"",
					previewModeColumnNameUpdated,
					previewModeAcceptedValuesUpdated,
					columnNames,
					automaticTimeZoneMapping,
					"",
					"",
					generatePhoneColumnsBlock(
						"Cell",
						"cell",
						"Cell",
					),
					generatePhoneColumnsBlock(
						"Home",
						"home",
						"Home",
					),
					generateEmailColumnsBlock(
						"Work",
						"work",
						"Work",
					),
					generateEmailColumnsBlock(
						"Personal",
						"personal",
						"Personal",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.0", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.1", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.2", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.3", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.callable_time_column", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.callable_time_column", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.contactable_time_column", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.contactable_time_column", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_column_name", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMapping),
					testDefaultHomeDivision("genesyscloud_outbound_contact_list."+resourceId),
				),
			},
			{
				Config: generateAttemptLimitResource(
					attemptLimitResourceID,
					attemptLimitName,
					"5",
					"5",
					"America/Chicago",
					"TODAY",
				) + generateOutboundAttemptLimitDataSource(
					attemptLimitDataSourceID,
					attemptLimitName,
					"genesyscloud_outbound_attempt_limit."+attemptLimitResourceID,
				) + `data "genesyscloud_auth_division_home" "home" {}` + generateOutboundContactList(
					resourceId,
					nameUpdated,
					"data.genesyscloud_auth_division_home.home.id",
					previewModeColumnNameUpdated,
					previewModeAcceptedValuesUpdated,
					columnNamesUpdated,
					automaticTimeZoneMappingUpdated,
					zipCodeColumnName,
					"genesyscloud_outbound_attempt_limit."+attemptLimitResourceID+".id",
					generatePhoneColumnsBlock(
						"Cell",
						"cell",
						"",
					),
					generatePhoneColumnsBlock(
						"Home",
						"home",
						"",
					),
					generateEmailColumnsBlock(
						"Work",
						"work",
						"",
					),
					generateEmailColumnsBlock(
						"Personal",
						"personal",
						"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.0", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.1", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.2", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.3", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "column_names.4", zipCodeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "zip_code_column_name", zipCodeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.column_name", "Cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.0.type", "cell"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.column_name", "Home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "phone_columns.1.type", "home"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.column_name", "Work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.1.type", "work"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.column_name", "Personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "email_columns.0.type", "personal"),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_column_name", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.0", previewModeColumnName),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "preview_mode_accepted_values.1", previewModeColumnNameUpdated),
					resource.TestCheckResourceAttr("genesyscloud_outbound_contact_list."+resourceId, "automatic_time_zone_mapping", automaticTimeZoneMappingUpdated),
					resource.TestCheckResourceAttrPair("data.genesyscloud_outbound_attempt_limit."+attemptLimitDataSourceID, "id",
						"genesyscloud_outbound_contact_list."+resourceId, "attempt_limit_id"),
					testDefaultHomeDivision("genesyscloud_outbound_contact_list."+resourceId),
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

func generateOutboundContactList(
	resourceId string,
	name string,
	divisionId string,
	previewModeColumnName string,
	previewModeAcceptedValues []string,
	columnNames []string,
	automaticTimeZoneMapping string,
	zipCodeColumnName string,
	attemptLimitId string,
	nestedBlocks ...string) string {
	if divisionId != "" {
		divisionId = fmt.Sprintf(`division_id = %s`, divisionId)
	}
	if previewModeColumnName != "" {
		previewModeColumnName = fmt.Sprintf(`preview_mode_column_name = "%s"`, previewModeColumnName)
	}
	if automaticTimeZoneMapping != "" {
		automaticTimeZoneMapping = fmt.Sprintf(`automatic_time_zone_mapping = %s`, automaticTimeZoneMapping)
	}
	if zipCodeColumnName != "" {
		zipCodeColumnName = fmt.Sprintf(`zip_code_column_name = "%s"`, zipCodeColumnName)
	}
	if attemptLimitId != "" {
		attemptLimitId = fmt.Sprintf(`attempt_limit_id = %s`, attemptLimitId)
	}
	return fmt.Sprintf(`
resource "genesyscloud_outbound_contact_list" "%s" {
	name = "%s"
	%s
	%s
	preview_mode_accepted_values = [%s]
	column_names = [%s] 
	%s
	%s
	%s
	%s
}
`, resourceId, name, divisionId, previewModeColumnName, strings.Join(previewModeAcceptedValues, ", "),
		strings.Join(columnNames, ", "), automaticTimeZoneMapping, zipCodeColumnName, attemptLimitId, strings.Join(nestedBlocks, "\n"))
}

func generatePhoneColumnsBlock(columnName string, columnType string, callableTimeColumn string) string {
	if callableTimeColumn != "" {
		callableTimeColumn = fmt.Sprintf(`callable_time_column = "%s"`, callableTimeColumn)
	}
	return fmt.Sprintf(`
	phone_columns {
		column_name = "%s"
		type        = "%s"
		%s
	}
`, columnName, columnType, callableTimeColumn)
}

func generateEmailColumnsBlock(columnName string, columnType string, contactableTimeColumn string) string {
	if contactableTimeColumn != "" {
		contactableTimeColumn = fmt.Sprintf(`contactable_time_column = "%s"`, contactableTimeColumn)
	}
	return fmt.Sprintf(`
	email_columns {
		column_name = "%s"
		type        = "%s"
		%s
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
		} else if isStatus404(resp) {
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
