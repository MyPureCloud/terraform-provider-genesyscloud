package outbound_filespecificationtemplate

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceOutboundFileSpecificationTemplate(t *testing.T) {
	t.Parallel()
	var (
		resourceId = "file_specification_template"

		// Create
		name                        = "tf-fst-" + uuid.NewString()
		description                 = "TF Test file specification template Delimited"
		format                      = "Delimited"
		numberOfHeaderLinesSkipped  = "1"
		numberOfTrailerLinesSkipped = "2"
		header                      = util.TrueValue
		delimiter                   = "Comma"

		column1Name   = "Home"
		column1Number = "0"
		column2Name   = "Address"
		column2Number = "1"

		preprocessingRule1Find        = "Dr"
		preprocessingRule1ReplaceWith = "Drive"
		preprocessingRule1Global      = util.FalseValue
		preprocessingRule1IgnoreCase  = util.TrueValue

		preprocessingRule2Find        = "([0-9]{3})"
		preprocessingRule2ReplaceWith = "($1)"

		// Update 1
		nameUpdated                        = "tf-fst-" + uuid.NewString()
		descriptionUpdated                 = "TF Test file specification template Delimited Update"
		numberOfHeaderLinesSkippedUpdated  = "3"
		numberOfTrailerLinesSkippedUpdated = "1"
		headerUpdated                      = util.FalseValue
		delimiterUpdated                   = "Custom"
		delimiterValueUpdated              = "^"

		column1NameUpdated   = "Work"
		column1NumberUpdated = "2"
		column2NameUpdated   = "Company"
		column2NumberUpdated = "3"

		preprocessingRule1FindUpdated        = "St"
		preprocessingRule1ReplaceWithUpdated = "Street"
		preprocessingRule1GlobalUpdated      = util.TrueValue
		preprocessingRule1IgnoreCaseUpdated  = util.FalseValue

		// Update 2
		formatUpdated               = "FixedLength"
		column1StartPositionUpdated = "0"
		column1LengthUpdated        = "20"
		column2StartPositionUpdated = "20"
		column2LengthUpdated        = "20"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateOutboundFileSpecificationTemplate(
					resourceId,
					name,
					strconv.Quote(description),
					format,
					strconv.Quote(numberOfHeaderLinesSkipped),
					strconv.Quote(numberOfTrailerLinesSkipped),
					strconv.Quote(header),
					strconv.Quote(delimiter),
					util.NullValue,
					generateOutboundFileSpecificationTemplateColumnInformation(
						strconv.Quote(column1Name),
						strconv.Quote(column1Number),
						util.NullValue,
						util.NullValue,
					),
					generateOutboundFileSpecificationTemplateColumnInformation(
						strconv.Quote(column2Name),
						strconv.Quote(column2Number),
						util.NullValue,
						util.NullValue,
					),
					generateOutboundFileSpecificationTemplatePreprocessingRule(
						strconv.Quote(preprocessingRule1Find),
						strconv.Quote(preprocessingRule1ReplaceWith),
						strconv.Quote(preprocessingRule1Global),
						strconv.Quote(preprocessingRule1IgnoreCase),
					),
					generateOutboundFileSpecificationTemplatePreprocessingRule(
						strconv.Quote(preprocessingRule2Find),
						strconv.Quote(preprocessingRule2ReplaceWith),
						util.NullValue,
						util.NullValue,
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "name", name),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "description", description),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "format", format),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "number_of_header_lines_skipped", numberOfHeaderLinesSkipped),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "number_of_trailer_lines_skipped", numberOfTrailerLinesSkipped),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "header", header),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "delimiter", delimiter),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "delimiter_value", ""),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.#", "2"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.0.column_name", column1Name),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.0.column_number", column1Number),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.1.column_name", column2Name),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.1.column_number", column2Number),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.#", "2"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.0.find", preprocessingRule1Find),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.0.replace_with", preprocessingRule1ReplaceWith),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.0.global", preprocessingRule1Global),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.0.ignore_case", preprocessingRule1IgnoreCase),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.1.find", preprocessingRule2Find),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.1.replace_with", preprocessingRule2ReplaceWith),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.1.global", util.FalseValue),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.1.ignore_case", util.FalseValue),
				),
			},
			{
				Config: generateOutboundFileSpecificationTemplate(
					resourceId,
					nameUpdated,
					strconv.Quote(descriptionUpdated),
					format,
					strconv.Quote(numberOfHeaderLinesSkippedUpdated),
					strconv.Quote(numberOfTrailerLinesSkippedUpdated),
					strconv.Quote(headerUpdated),
					strconv.Quote(delimiterUpdated),
					strconv.Quote(delimiterValueUpdated),
					generateOutboundFileSpecificationTemplateColumnInformation(
						strconv.Quote(column1NameUpdated),
						strconv.Quote(column1NumberUpdated),
						util.NullValue,
						util.NullValue,
					),
					generateOutboundFileSpecificationTemplateColumnInformation(
						strconv.Quote(column2NameUpdated),
						strconv.Quote(column2NumberUpdated),
						util.NullValue,
						util.NullValue,
					),
					generateOutboundFileSpecificationTemplatePreprocessingRule(
						strconv.Quote(preprocessingRule1FindUpdated),
						strconv.Quote(preprocessingRule1ReplaceWithUpdated),
						strconv.Quote(preprocessingRule1GlobalUpdated),
						strconv.Quote(preprocessingRule1IgnoreCaseUpdated),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "format", format),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "delimiter", delimiterUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "delimiter_value", delimiterValueUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "number_of_header_lines_skipped", numberOfHeaderLinesSkippedUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "number_of_trailer_lines_skipped", numberOfTrailerLinesSkippedUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "header", headerUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.#", "2"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.0.column_name", column1NameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.0.column_number", column1NumberUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.1.column_name", column2NameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.1.column_number", column2NumberUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.0.find", preprocessingRule1FindUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.0.replace_with", preprocessingRule1ReplaceWithUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.0.global", preprocessingRule1GlobalUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.0.ignore_case", preprocessingRule1IgnoreCaseUpdated),
				),
			},
			{
				Config: generateOutboundFileSpecificationTemplate(
					resourceId,
					nameUpdated,
					strconv.Quote(descriptionUpdated),
					formatUpdated,
					util.NullValue,
					util.NullValue,
					util.NullValue,
					util.NullValue,
					util.NullValue,
					generateOutboundFileSpecificationTemplateColumnInformation(
						strconv.Quote(column1NameUpdated),
						util.NullValue,
						strconv.Quote(column1StartPositionUpdated),
						strconv.Quote(column1LengthUpdated),
					),
					generateOutboundFileSpecificationTemplateColumnInformation(
						strconv.Quote(column2NameUpdated),
						util.NullValue,
						strconv.Quote(column2StartPositionUpdated),
						strconv.Quote(column2LengthUpdated),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "name", nameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "format", formatUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.#", "2"),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.0.column_name", column1NameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.0.start_position", column1StartPositionUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.0.length", column1LengthUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.1.column_name", column2NameUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.1.start_position", column2StartPositionUpdated),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "column_information.1.length", column2LengthUpdated),
				),
			},
			{
				ResourceName:      resourceName + "." + resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundFileSpecificationTemplateDestroyed,
	})
}

func generateOutboundFileSpecificationTemplate(
	resourceId string,
	name string,
	description string,
	format string,
	numberOfHeaderLinesSkipped string,
	numberOfTrailerLinesSkipped string,
	header string,
	delimiter string,
	delimiterValue string,
	nestedBlocks ...string,
) string {
	return fmt.Sprintf(`
	resource "genesyscloud_outbound_filespecificationtemplate" "%s" {
		name = "%s"
		description = %s
		format = "%s"
		number_of_header_lines_skipped = %s
		number_of_trailer_lines_skipped = %s
		header = %s
		delimiter = %s
		delimiter_value = %s
		%s
	}
	`,
		resourceId,
		name,
		description,
		format,
		numberOfHeaderLinesSkipped,
		numberOfTrailerLinesSkipped,
		header,
		delimiter,
		delimiterValue,
		strings.Join(nestedBlocks, "\n"))
}

func generateOutboundFileSpecificationTemplateColumnInformation(
	columnName string,
	columnNumber string,
	startPosition string,
	length string,
) string {
	return fmt.Sprintf(`
	column_information {
		column_name = %s
		column_number = %s
		start_position = %s
		length = %s
	}
`, columnName, columnNumber, startPosition, length)
}

func generateOutboundFileSpecificationTemplatePreprocessingRule(
	find string,
	replaceWith string,
	global string,
	ignoreCase string,
) string {
	return fmt.Sprintf(`
	preprocessing_rule {
		find = %s
		replace_with = %s
		global = %s
		ignore_case = %s
	}
`, find, replaceWith, global, ignoreCase)
}

func testVerifyOutboundFileSpecificationTemplateDestroyed(state *terraform.State) error {
	outboundAPI := platformclientv2.NewOutboundApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != resourceName {
			continue
		}

		fileSpecificationTemplate, resp, err := outboundAPI.GetOutboundFilespecificationtemplate(rs.Primary.ID)
		if fileSpecificationTemplate != nil {
			return fmt.Errorf("file specification template (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// File specification template not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("unexpected error: %s", err)
		}
	}
	// Success. All file specification templates destroyed
	return nil
}
