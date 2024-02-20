package outbound_filespecificationtemplate

import (
	"fmt"
	"testing"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundFileSpecificationTemplate(t *testing.T) {

	var (
		resourceId                  = "file_specification_template"
		dataSourceId                = "file_specification_template_data"
		name                        = "File Specification Template" + uuid.NewString()
		description                 = "TF Test File specification template"
		format                      = "Delimited"
		numberOfHeaderLinesSkipped  = "1"
		numberOfTrailerLinesSkipped = "2"
		header                      = gcloud.TrueValue
		delimiter                   = "Custom"
		delimiterValue              = "^"
		column1Name                 = "Phone"
		column1Number               = "0"
		column2Name                 = "Address"
		column2Number               = "1"
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
					delimiterValue,
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
				) + generateOutboundFileSpecificationTemplateDataSource(
					dataSourceId,
					name,
					"genesyscloud_outbound_filespecificationtemplate."+resourceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_filespecificationtemplate."+resourceId, "id",
						"data.genesyscloud_outbound_filespecificationtemplate."+dataSourceId, "id"),
				),
			},
		},
	})
}

func generateOutboundFileSpecificationTemplateDataSource(id string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_filespecificationtemplate" "%s" {
	name       = "%s"
	depends_on = [%s]
}
`, id, name, dependsOn)
}
