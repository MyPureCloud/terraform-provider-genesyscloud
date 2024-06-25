package architect_datatable

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceArchitectDatatable(t *testing.T) {
	var (
		tableResource1 = "arch-table1"
		tableName1     = "Terraform Table1-" + uuid.NewString()
		tableName2     = "Terraform Table2-" + uuid.NewString()
		tableDesc1     = "Terraform test table1"
		tableDesc2     = "Terraform test table 2"

		propNameKey = "key"
		propInt     = "test-int"
		propBool    = "Test Bool"
		propNum     = "Test num"

		propTitleKey  = "key-title"
		propTitleInt  = "int-title"
		propTitleNum  = "num-title"
		propTitleBool = "bool-title"

		typeString = "string"
		typeBool   = "boolean"
		typeInt    = "integer"
		typeNum    = "number"

		defInt1  = "100"
		defNum1  = "10.1"
		defBool1 = "true"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create architect_datatable with a key and one other property
				Config: generateArchitectDatatableResource(
					tableResource1,
					tableName1,
					strconv.Quote(tableDesc1),
					generateArchitectDatatableProperty(propBool, typeBool, util.NullValue, util.NullValue),
					generateArchitectDatatableProperty(propNameKey, typeString, util.NullValue, util.NullValue),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "name", tableName1),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "description", tableDesc1),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.0.name", propBool),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.0.type", typeBool),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.1.name", propNameKey),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.1.type", typeString),
				),
			},
			{
				// Update with a new name, description, and additional properties
				Config: generateArchitectDatatableResource(
					tableResource1,
					tableName2,
					strconv.Quote(tableDesc2),
					generateArchitectDatatableProperty(propNameKey, typeString, strconv.Quote(propTitleKey), util.NullValue),
					generateArchitectDatatableProperty(propInt, typeInt, strconv.Quote(propTitleInt), strconv.Quote(defInt1)),
					generateArchitectDatatableProperty(propBool, typeBool, strconv.Quote(propTitleBool), strconv.Quote(defBool1)),
					generateArchitectDatatableProperty(propNum, typeNum, strconv.Quote(propTitleNum), strconv.Quote(defNum1)),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "name", tableName2),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "description", tableDesc2),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.0.name", propNameKey),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.0.type", typeString),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.0.title", propTitleKey),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.1.name", propInt),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.1.type", typeInt),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.1.title", propTitleInt),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.1.default", defInt1),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.2.name", propBool),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.2.type", typeBool),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.2.title", propTitleBool),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.2.default", defBool1),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.3.name", propNum),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.3.type", typeNum),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.3.title", propTitleNum),
					resource.TestCheckResourceAttr("genesyscloud_architect_datatable."+tableResource1, "properties.3.default", defNum1),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_architect_datatable." + tableResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyDatatablesDestroyed,
	})
}

func testVerifyDatatablesDestroyed(state *terraform.State) error {
	archAPI := platformclientv2.NewArchitectApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_architect_datatable" {
			continue
		}

		datatable, resp, err := sdkGetArchitectDatatable(rs.Primary.ID, "", archAPI)
		if datatable != nil {
			return fmt.Errorf("Datatable (%s) still exists", rs.Primary.ID)
		} else if util.IsStatus404(resp) {
			// Datatable not found as expected
			continue
		} else {
			// Unexpected error
			return fmt.Errorf("Unexpected error: %s", err)
		}
	}
	// Success. All Datatables destroyed
	return nil
}

func generateArchitectDatatableResource(
	resourceID string,
	name string,
	description string,
	properties ...string) string {
	return fmt.Sprintf(`resource "genesyscloud_architect_datatable" "%s" {
		name = "%s"
		description = %s
		%s
	}
	`, resourceID, name, description, strings.Join(properties, "\n"))
}

func generateArchitectDatatableProperty(
	name string,
	propType string,
	title string,
	defaultVal string) string {
	return fmt.Sprintf(`properties {
		name = "%s"
		type = "%s"
		title = %s
        default = %s
	}
	`, name, propType, title, defaultVal)
}

// used for testing only
func sdkGetArchitectDatatable(datatableId string, expand string, api *platformclientv2.ArchitectApi) (*Datatable, *platformclientv2.APIResponse, error) {
	apiClient := &api.Configuration.APIClient

	// create path and map variables
	path := api.Configuration.BasePath + "/api/v2/flows/datatables/" + datatableId

	headerParams := make(map[string]string)
	queryParams := make(map[string]string)

	// oauth required
	if api.Configuration.AccessToken != "" {
		headerParams["Authorization"] = "Bearer " + api.Configuration.AccessToken
	}
	// add default headers if any
	for key := range api.Configuration.DefaultHeader {
		headerParams[key] = api.Configuration.DefaultHeader[key]
	}

	queryParams["expand"] = apiClient.ParameterToString(expand, "")

	headerParams["Content-Type"] = "application/json"
	headerParams["Accept"] = "application/json"

	var successPayload *Datatable
	response, err := apiClient.CallAPI(path, http.MethodGet, nil, headerParams, queryParams, nil, "", nil)
	if err != nil {
		// Nothing special to do here, but do avoid processing the response
	} else if response.Error != nil {
		err = errors.New(response.ErrorMessage)
	} else {
		err = json.Unmarshal(response.RawBody, &successPayload)
	}
	return successPayload, response, err
}
