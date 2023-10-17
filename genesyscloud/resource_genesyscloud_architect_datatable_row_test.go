package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

func TestAccResourceArchitectDatatableRow(t *testing.T) {
	var (
		tableResource1 = "arch-table1"
		rowResource1   = "table-row-1"
		tableName1     = "Terraform Table1-" + uuid.NewString()

		propNameKey = "key"
		propInt     = "test-int"
		propBool    = "Test Bool"
		propNum     = "Test num"
		propStr     = "Test str"

		typeString = "string"
		typeBool   = "boolean"
		typeInt    = "integer"
		typeNum    = "number"

		defInt1  = "100"
		defBool1 = "true"
		defStr   = "defStr"
		defNum   = "0" // Default number when no default is given

		keyVal1 = "tf test-key1"
		keyVal2 = "tf test-key2"
		intVal1 = "1"
		intVal2 = "2"
		numVal1 = "1.1"
		numVal2 = "2.5"
		strVal1 = "test str-val1"
		strVal2 = "test str-val2"

		tableConfig = generateArchitectDatatableResource(
			tableResource1,
			tableName1,
			nullValue,
			generateArchitectDatatableProperty(propNameKey, typeString, strconv.Quote(propNameKey), nullValue),
			generateArchitectDatatableProperty(propInt, typeInt, strconv.Quote(propInt), strconv.Quote(defInt1)),
			generateArchitectDatatableProperty(propBool, typeBool, strconv.Quote(propBool), strconv.Quote(defBool1)),
			generateArchitectDatatableProperty(propNum, typeNum, strconv.Quote(propNum), nullValue), // No default
			generateArchitectDatatableProperty(propStr, typeString, strconv.Quote(propStr), strconv.Quote(defStr)),
		)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create datatable with a key and property of each type. Add 1 row with all defaults
				Config: tableConfig + generateArchitectDatatableRowResource(
					rowResource1,
					"genesyscloud_architect_datatable."+tableResource1+".id",
					keyVal1,
					GenerateJsonEncodedProperties(
						GenerateJsonProperty(propInt, intVal1), // Most props in state should be default
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable_row."+rowResource1, "key_value", keyVal1),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_datatable_row."+rowResource1, "datatable_id", "genesyscloud_architect_datatable."+tableResource1, "id"),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propInt, intVal1),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propBool, defBool1),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propNum, defNum),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propStr, defStr),
				),
			},
			{
				// Update row with all properties
				Config: tableConfig + generateArchitectDatatableRowResource(
					rowResource1,
					"genesyscloud_architect_datatable."+tableResource1+".id",
					keyVal1,
					GenerateJsonEncodedProperties(
						GenerateJsonProperty(propInt, intVal1),
						GenerateJsonProperty(propStr, strconv.Quote(strVal1)),
						GenerateJsonProperty(propBool, falseValue),
						GenerateJsonProperty(propNum, numVal1),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable_row."+rowResource1, "key_value", keyVal1),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_datatable_row."+rowResource1, "datatable_id", "genesyscloud_architect_datatable."+tableResource1, "id"),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propInt, intVal1),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propBool, falseValue),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propNum, numVal1),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propStr, strVal1),
				),
			},
			{
				// Update row by removing and modifying existing properties
				Config: tableConfig + generateArchitectDatatableRowResource(
					rowResource1,
					"genesyscloud_architect_datatable."+tableResource1+".id",
					keyVal1,
					GenerateJsonEncodedProperties(
						GenerateJsonProperty(propInt, intVal2),
						GenerateJsonProperty(propStr, strconv.Quote(strVal2)),
						GenerateJsonProperty(propNum, numVal2),
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable_row."+rowResource1, "key_value", keyVal1),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_datatable_row."+rowResource1, "datatable_id", "genesyscloud_architect_datatable."+tableResource1, "id"),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propInt, intVal2),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propBool, defBool1),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propNum, numVal2),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propStr, strVal2),
				),
			},
			{
				// Update row with a new key. This should delete the original row and create a new one
				Config: tableConfig + generateArchitectDatatableRowResource(
					rowResource1,
					"genesyscloud_architect_datatable."+tableResource1+".id",
					keyVal2,
					// Raw JSON to test the JSON diff logic
					fmt.Sprintf(`"{  \"%s\":   %s,  \"%s\": \"%s\"}"`, propInt, intVal2, propStr, strVal2),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable_row."+rowResource1, "key_value", keyVal2),
					resource.TestCheckResourceAttrPair("genesyscloud_architect_datatable_row."+rowResource1, "datatable_id", "genesyscloud_architect_datatable."+tableResource1, "id"),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propInt, intVal2),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propBool, defBool1),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propNum, defNum),
					ValidateValueInJsonAttr("genesyscloud_architect_datatable_row."+rowResource1, "properties_json", propStr, strVal2),
				),
			},
			{
				// Import/Read
				ResourceName:            "genesyscloud_architect_datatable_row." + rowResource1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"properties_json"}, // ImportState ignores DiffSuppressFuncs, so it cannot validate the JSON
				ImportStateIdFunc:       importDatatableRowId(tableResource1, keyVal2),
			},
		},
		CheckDestroy: testVerifyDatatableRowsDestroyed,
	})
}

func generateArchitectDatatableRowResource(
	resourceID string,
	tableID string,
	keyVal string,
	properties string) string {
	return fmt.Sprintf(`resource "genesyscloud_architect_datatable_row" "%s" {
		datatable_id = %s
		key_value = "%s"
		properties_json = %s
	}
	`, resourceID, tableID, keyVal, properties)
}

func testVerifyDatatableRowsDestroyed(state *terraform.State) error {
	archAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_datatable_row" {
			continue
		}

		tableID, keyStr := splitDatatableRowId(rs.Primary.ID)
		row, resp, err := archAPI.GetFlowsDatatableRow(tableID, keyStr, false)
		if row != nil {
			return fmt.Errorf("Datatable Row (%s) still exists", rs.Primary.ID)
		} else if IsStatus404(resp) {
			// Datatable Row not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All Datatable Rows destroyed
	return nil
}

func importDatatableRowId(tableResource string, rowKey string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		if tableRes, ok := state.RootModule().Resources["genesyscloud_architect_datatable."+tableResource]; ok {
			return createDatatableRowId(tableRes.Primary.ID, rowKey), nil
		} else {
			return "", fmt.Errorf("Failed to find table resource %s in state", tableResource)
		}
	}
}
