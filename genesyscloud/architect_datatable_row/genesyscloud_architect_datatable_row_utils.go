package architect_datatable_row

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

// Row IDs structured as {table-id}/{key-value}
func createDatatableRowId(tableId string, keyVal string) string {
	return strings.Join([]string{tableId, keyVal}, "/")
}

func splitDatatableRowId(rowId string) (string, string) {
	split := strings.SplitN(rowId, "/", 2)
	if len(split) == 2 {
		return split[0], split[1]
	}
	return "", ""
}

func buildSdkRowPropertyMap(propertiesJson string, keyStr string) (map[string]interface{}, diag.Diagnostics) {
	// Property value must be empty or a JSON object (map)
	propMap := map[string]interface{}{}
	if propertiesJson != "" {
		if err := json.Unmarshal([]byte(propertiesJson), &propMap); err != nil {
			return nil, util.BuildDiagnosticError(resourceName, fmt.Sprintf("Error parsing properties_json value %s", propertiesJson), err)
		}
	}
	// Set the key value
	propMap["key"] = keyStr
	return propMap, nil
}

func customizeDatatableRowDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	// Defaults must be set on missing properties

	if !diff.NewValueKnown("properties_json") {
		// properties_json value not yet in final state. Nothing to do.
		return nil
	}

	if !diff.NewValueKnown("datatable_id") {
		// datatable_id not yet in final state, but properties_json is marked as known.
		// There may be computed defaults to set on properties_json that we do not know yet.
		diff.SetNewComputed("properties_json")
		return nil
	}

	tableId := diff.Get("datatable_id").(string)
	keyStr := diff.Get("key_value").(string)
	id := createDatatableRowId(tableId, keyStr)

	propertiesJson := diff.Get("properties_json").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig

	// Retrieve defaults from the architect_datatable for this row
	datatable, getErr := getArchitectDatatableCached(ctx, tableId, sdkConfig)
	if getErr != nil {
		return fmt.Errorf("Failed to read architect_datatable %s: %s", tableId, getErr)
	}

	// Parse resource properties into map
	configMap := map[string]interface{}{}
	if propertiesJson == "" {
		propertiesJson = "{}" // empty object by default
	}
	if err := json.Unmarshal([]byte(propertiesJson), &configMap); err != nil {
		return fmt.Errorf("Failure to parse properties_json for %s: %s", id, err)
	}

	// For each property in the schema, check if a value is set in the config
	if datatable.Schema != nil && datatable.Schema.Properties != nil {
		for name, prop := range *datatable.Schema.Properties {
			if name == "key" {
				// Skip setting the key value
				continue
			}
			if _, set := configMap[name]; !set {
				// Property in schema not set. Override diff with expected default.
				if prop.Default != nil {
					configMap[name] = *prop.Default
				} else if *prop.VarType == "boolean" {
					// Booleans default to false
					configMap[name] = false
				} else if *prop.VarType == "string" {
					// Strings default to empty
					configMap[name] = ""
				} else if *prop.VarType == "integer" || *prop.VarType == "number" {
					// Numbers default to 0
					configMap[name] = 0
				}
			}
		}
	}

	// Marshal back to string and set as the diff value
	result, err := json.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("Failure to marshal properties for %s: %s", id, err)
	}

	diff.SetNew("properties_json", string(result))
	return nil
}

// Prevent getting the architect_datatable schema on every row diff
// by caching the results for the duration of the TF run
var archDatatableCache sync.Map

func getArchitectDatatableCached(ctx context.Context, tableID string, config *platformclientv2.Configuration) (*Datatable, error) {
	archProxy := getArchitectDatatableRowProxy(config)
	if table, ok := archDatatableCache.Load(tableID); ok {
		return table.(*Datatable), nil
	}

	datatable, _, getErr := archProxy.getArchitectDatatable(ctx, tableID, "schema")
	if getErr != nil {
		return nil, fmt.Errorf("Failed to read architect_datatable %s: %s", tableID, getErr)
	}
	archDatatableCache.Store(tableID, datatable)
	return datatable, nil
}
