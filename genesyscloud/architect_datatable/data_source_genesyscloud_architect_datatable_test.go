package architect_datatable

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceArchitectDatatable(t *testing.T) {
	var (
		tableResourceLabel = "arch-table1"
		tableName          = "Terraform Table1-" + uuid.NewString()
		tableDesc          = "Terraform test table1"

		propNameKey = "key"
		propBool    = "Test Bool"
		typeString  = "string"
		typeBool    = "boolean"

		tableDataSourceLabel = "arch-table1-ds"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create architect_datatable with a key and one other property
				Config: generateArchitectDatatableResource(
					tableResourceLabel,
					tableName,
					strconv.Quote(tableDesc),
					generateArchitectDatatableProperty(propBool, typeBool, util.NullValue, util.NullValue),
					generateArchitectDatatableProperty(propNameKey, typeString, util.NullValue, util.NullValue),
				) + generateArchitectDatatableDataSource(
					tableDataSourceLabel,
					"genesyscloud_architect_datatable."+tableResourceLabel+".name",
					"genesyscloud_architect_datatable."+tableResourceLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.genesyscloud_architect_datatable."+tableDataSourceLabel, "id",
						"genesyscloud_architect_datatable."+tableResourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateArchitectDatatableDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_architect_datatable" "%s" {
		name = %s
		depends_on=[%s]
	}
	`, resourceLabel, name, dependsOnResource)
}
