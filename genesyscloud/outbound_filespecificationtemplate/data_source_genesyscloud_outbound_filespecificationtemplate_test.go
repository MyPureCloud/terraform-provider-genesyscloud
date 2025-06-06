package outbound_filespecificationtemplate

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOutboundFileSpecificationTemplate(t *testing.T) {

	var (
		resourceLabel               = "file_specification_template"
		dataSourceLabel             = "file_specification_template_data"
		name                        = "File Specification Template" + uuid.NewString()
		description                 = "TF Test File specification template"
		format                      = "Delimited"
		numberOfHeaderLinesSkipped  = "1"
		numberOfTrailerLinesSkipped = "2"
		header                      = util.TrueValue
		delimiter                   = "Custom"
		delimiterValue              = "^"
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
					strconv.Quote(delimiterValue),
				) + generateOutboundFileSpecificationTemplateDataSource(
					dataSourceLabel,
					name,
					"genesyscloud_outbound_filespecificationtemplate."+resourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("genesyscloud_outbound_filespecificationtemplate."+resourceLabel, "id",
						"data.genesyscloud_outbound_filespecificationtemplate."+dataSourceLabel, "id"),
				),
			},
		},
	})
}

func generateOutboundFileSpecificationTemplateDataSource(dataSourceLabel string, name string, dependsOn string) string {
	return fmt.Sprintf(`
data "genesyscloud_outbound_filespecificationtemplate" "%s" {
	name       = "%s"
	depends_on = [%s]
}
`, dataSourceLabel, name, dependsOn)
}
