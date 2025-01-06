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
	"github.com/mypurecloud/platform-client-sdk-go/v150/platformclientv2"
)

func TestAccResourceOutboundFileSpecificationTemplate(t *testing.T) {
	t.Parallel()
	var (
		resourceLabel = "file_specification_template"

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
					resourceLabel,
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
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "description", description),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "format", format),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "number_of_header_lines_skipped", numberOfHeaderLinesSkipped),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "number_of_trailer_lines_skipped", numberOfTrailerLinesSkipped),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "header", header),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "delimiter", delimiter),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "delimiter_value", ""),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.0.column_name", column1Name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.0.column_number", column1Number),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.1.column_name", column2Name),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.1.column_number", column2Number),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.0.find", preprocessingRule1Find),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.0.replace_with", preprocessingRule1ReplaceWith),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.0.global", preprocessingRule1Global),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.0.ignore_case", preprocessingRule1IgnoreCase),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.1.find", preprocessingRule2Find),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.1.replace_with", preprocessingRule2ReplaceWith),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.1.global", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.1.ignore_case", util.FalseValue),
				),
			},
			{
				Config: generateOutboundFileSpecificationTemplate(
					resourceLabel,
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
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", nameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "format", format),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "delimiter", delimiterUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "delimiter_value", delimiterValueUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "number_of_header_lines_skipped", numberOfHeaderLinesSkippedUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "number_of_trailer_lines_skipped", numberOfTrailerLinesSkippedUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "header", headerUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.0.column_name", column1NameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.0.column_number", column1NumberUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.1.column_name", column2NameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.1.column_number", column2NumberUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.0.find", preprocessingRule1FindUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.0.replace_with", preprocessingRule1ReplaceWithUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.0.global", preprocessingRule1GlobalUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "preprocessing_rule.0.ignore_case", preprocessingRule1IgnoreCaseUpdated),
				),
			},
			{
				Config: generateOutboundFileSpecificationTemplate(
					resourceLabel,
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
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "name", nameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "format", formatUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.0.column_name", column1NameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.0.start_position", column1StartPositionUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.0.length", column1LengthUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.1.column_name", column2NameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.1.start_position", column2StartPositionUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+resourceLabel, "column_information.1.length", column2LengthUpdated),
				),
			},
			{
				ResourceName:      ResourceType + "." + resourceLabel,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyOutboundFileSpecificationTemplateDestroyed,
	})
}

func generateOutboundFileSpecificationTemplate(
	resourceLabel string,
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
		resourceLabel,
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
		if rs.Type != ResourceType {
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
