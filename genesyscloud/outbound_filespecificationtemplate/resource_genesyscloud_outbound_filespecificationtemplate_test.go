package outbound_filespecificationtemplate

import (
	"fmt"
	"strings"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
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
		header                      = gcloud.TrueValue
		delimiter                   = "Comma"

		column1Name   = "Home"
		column1Number = "0"
		column2Name   = "Address"
		column2Number = "1"

		preprocessingRule1Find        = "Dr"
		preprocessingRule1ReplaceWith = "Drive"
		preprocessingRule1Global      = gcloud.FalseValue
		preprocessingRule1IgnoreCase  = gcloud.TrueValue

		preprocessingRule2Find        = "([0-9]{3})"
		preprocessingRule2ReplaceWith = "($1)"

		// Update
		nameUpdated                        = "tf-fst-" + uuid.NewString()
		descriptionUpdated                 = "TF Test file specification template Delimited Update"
		numberOfHeaderLinesSkippedUpdated  = "3"
		numberOfTrailerLinesSkippedUpdated = "1"
		headerUpdated                      = gcloud.FalseValue
		delimiterUpdated                   = "Custom"
		delimiterValueUpdated              = "^"

		column1NameUpdated   = "Work"
		column1NumberUpdated = "2"
		column2NameUpdated   = "Company"
		column2NumberUpdated = "3"

		preprocessingRule1FindUpdated        = "St"
		preprocessingRule1ReplaceWithUpdated = "Street"
		preprocessingRule1GlobalUpdated      = gcloud.TrueValue
		preprocessingRule1IgnoreCaseUpdated  = gcloud.FalseValue

		formatUpdated               = "FixedLength"
		column1StartPositionUpdated = "0"
		column1LengthUpdated        = "20"
		column2StartPositionUpdated = "20"
		column2LengthUpdated        = "20"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { gcloud.TestAccPreCheck(t) },
		ProviderFactories: gcloud.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: generateOutboundFileSpecificationTemplate(
					resourceId,
					name,
					description,
					format,
					numberOfHeaderLinesSkipped,
					numberOfTrailerLinesSkipped,
					header,
					delimiter,
					"",
					generateOutboundFileSpecificationTemplateColumnInformation(
						column1Name,
						column1Number,
						"",
						"",
					),
					generateOutboundFileSpecificationTemplateColumnInformation(
						column2Name,
						column2Number,
						"",
						"",
					),
					generateOutboundFileSpecificationTemplatePreprocessingRule(
						preprocessingRule1Find,
						preprocessingRule1ReplaceWith,
						preprocessingRule1Global,
						preprocessingRule1IgnoreCase,
					),
					generateOutboundFileSpecificationTemplatePreprocessingRule(
						preprocessingRule2Find,
						preprocessingRule2ReplaceWith,
						"",
						"",
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
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.1.global", gcloud.FalseValue),
					resource.TestCheckResourceAttr(resourceName+"."+resourceId, "preprocessing_rule.1.ignore_case", gcloud.FalseValue),
				),
			},
			{
				Config: generateOutboundFileSpecificationTemplate(
					resourceId,
					nameUpdated,
					descriptionUpdated,
					format,
					numberOfHeaderLinesSkippedUpdated,
					numberOfTrailerLinesSkippedUpdated,
					headerUpdated,
					delimiterUpdated,
					delimiterValueUpdated,
					generateOutboundFileSpecificationTemplateColumnInformation(
						column1NameUpdated,
						column1NumberUpdated,
						"",
						"",
					),
					generateOutboundFileSpecificationTemplateColumnInformation(
						column2NameUpdated,
						column2NumberUpdated,
						"",
						"",
					),
					generateOutboundFileSpecificationTemplatePreprocessingRule(
						preprocessingRule1FindUpdated,
						preprocessingRule1ReplaceWithUpdated,
						preprocessingRule1GlobalUpdated,
						preprocessingRule1IgnoreCaseUpdated,
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
					descriptionUpdated,
					formatUpdated,
					"",
					"",
					"",
					"",
					"",
					generateOutboundFileSpecificationTemplateColumnInformation(
						column1NameUpdated,
						"",
						column1StartPositionUpdated,
						column1LengthUpdated,
					),
					generateOutboundFileSpecificationTemplateColumnInformation(
						column2NameUpdated,
						"",
						column2StartPositionUpdated,
						column2LengthUpdated,
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
	if description != "" {
		description = fmt.Sprintf(`description = "%s"`, description)
	}
	if format != "" {
		format = fmt.Sprintf(`format = "%s"`, format)
	}
	if numberOfHeaderLinesSkipped != "" {
		numberOfHeaderLinesSkipped = fmt.Sprintf(`number_of_header_lines_skipped = "%s"`, numberOfHeaderLinesSkipped)
	}
	if numberOfTrailerLinesSkipped != "" {
		numberOfTrailerLinesSkipped = fmt.Sprintf(`number_of_trailer_lines_skipped = "%s"`, numberOfTrailerLinesSkipped)
	}
	if header != "" {
		header = fmt.Sprintf(`header = "%s"`, header)
	}
	if delimiter != "" {
		delimiter = fmt.Sprintf(`delimiter = "%s"`, delimiter)
	}
	if delimiterValue != "" {
		delimiterValue = fmt.Sprintf(`delimiter_value = "%s"`, delimiterValue)
	}
	return fmt.Sprintf(`
	resource "genesyscloud_outbound_filespecificationtemplate" "%s" {
		name = "%s"
		%s
		%s
		%s
		%s
		%s
		%s
		%s
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
	if columnName != "" {
		columnName = fmt.Sprintf(`column_name = "%s"`, columnName)
	}
	if columnNumber != "" {
		columnNumber = fmt.Sprintf(`column_number = "%s"`, columnNumber)
	}
	if startPosition != "" {
		startPosition = fmt.Sprintf(`start_position = "%s"`, startPosition)
	}
	if length != "" {
		length = fmt.Sprintf(`length = "%s"`, length)
	}
	return fmt.Sprintf(`
	column_information {
		%s
		%s
		%s
		%s
	}
`, columnName, columnNumber, startPosition, length)
}

func generateOutboundFileSpecificationTemplatePreprocessingRule(
	find string,
	replaceWith string,
	global string,
	ignoreCase string,
) string {
	if find != "" {
		find = fmt.Sprintf(`find = "%s"`, find)
	}
	if replaceWith != "" {
		replaceWith = fmt.Sprintf(`replace_with = "%s"`, replaceWith)
	}
	if global != "" {
		global = fmt.Sprintf(`global = "%s"`, global)
	}
	if ignoreCase != "" {
		ignoreCase = fmt.Sprintf(`ignore_case = "%s"`, ignoreCase)
	}
	return fmt.Sprintf(`
	preprocessing_rule {
		%s
		%s
		%s
		%s
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
		} else if gcloud.IsStatus404(resp) {
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
