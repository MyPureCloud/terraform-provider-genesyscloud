package genesyscloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
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

func getAllArchitectDatatableRows(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	archAPI := platformclientv2.NewArchitectApiWithConfig(clientConfig)

	tables, err := getAllArchitectDatatables(ctx, clientConfig)
	if err != nil {
		return nil, err
	}

	for tableId, tableMeta := range tables {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			rows, _, getErr := archAPI.GetFlowsDatatableRows(tableId, pageNum, pageSize, false, "")
			if getErr != nil {
				return nil, diag.Errorf("Failed to get page of Datatable Rows: %v", getErr)
			}

			if rows.Entities == nil || len(*rows.Entities) == 0 {
				break
			}

			for _, row := range *rows.Entities {
				if keyVal, ok := row["key"]; ok {
					keyStr := keyVal.(string) // Keys must be strings
					resources[createDatatableRowId(tableId, keyStr)] = &resourceExporter.ResourceMeta{Name: tableMeta.Name + "_" + keyStr}
				}
			}
		}
	}

	return resources, nil
}

func ArchitectDatatableRowExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllArchitectDatatableRows),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"datatable_id": {RefType: "genesyscloud_architect_datatable"},
		},
		JsonEncodeAttributes: []string{"properties_json"},
	}
}

func ResourceArchitectDatatableRow() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud Architect Datatable Row",

		CreateContext: CreateWithPooledClient(createArchitectDatatableRow),
		ReadContext:   ReadWithPooledClient(readArchitectDatatableRow),
		UpdateContext: UpdateWithPooledClient(updateArchitectDatatableRow),
		DeleteContext: DeleteWithPooledClient(deleteArchitectDatatableRow),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"datatable_id": {
				Description: "Datatable ID that contains this row. If this is changed, a new row is created.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"key_value": {
				Description: "Value for this row's key. If this is changed, a new row is created.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"properties_json": {
				Description:      "JSON object containing properties and values for this row. Defaults will be set for missing properties.",
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: SuppressEquivalentJsonDiffs,
			},
		},
		CustomizeDiff: customizeDatatableRowDiff,
	}
}

func createArchitectDatatableRow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableId := d.Get("datatable_id").(string)
	keyStr := d.Get("key_value").(string)
	properties := d.Get("properties_json").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	rowMap, diagErr := buildSdkRowPropertyMap(properties, keyStr)
	if diagErr != nil {
		return diagErr
	}

	rowId := createDatatableRowId(tableId, keyStr)
	log.Printf("Creating Datatable Row %s", rowId)

	_, _, err := archAPI.PostFlowsDatatableRows(tableId, rowMap)
	if err != nil {
		return diag.Errorf("Failed to create Datatable Row %s: %s", rowId, err)
	}

	d.SetId(rowId)

	log.Printf("Created Datatable Row %s", d.Id())
	return readArchitectDatatableRow(ctx, d, meta)
}

func readArchitectDatatableRow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableId, keyStr := splitDatatableRowId(d.Id())
	if keyStr == "" {
		return diag.Errorf("Invalid Row ID %s", d.Id())
	}

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Reading Datatable Row %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		row, resp, getErr := archAPI.GetFlowsDatatableRow(tableId, keyStr, false)
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read Datatable Row %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read Datatable Row %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectDatatableRow())
		d.Set("datatable_id", tableId)
		d.Set("key_value", keyStr)

		// The key value is exposed through a separate attribute, so it should be removed from the value map
		delete(*row, "key")

		valueBytes, err := json.Marshal(*row)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to marshal row map %v: %v", *row, err))
		}
		d.Set("properties_json", string(valueBytes))

		log.Printf("Read Datatable Row %s", d.Id())
		return cc.CheckState()
	})
}

func updateArchitectDatatableRow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableId := d.Get("datatable_id").(string)
	keyStr := d.Get("key_value").(string)
	properties := d.Get("properties_json").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	rowMap, diagErr := buildSdkRowPropertyMap(properties, keyStr)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updating Datatable Row %s", d.Id())

	_, _, err := archAPI.PutFlowsDatatableRow(tableId, keyStr, rowMap)
	if err != nil {
		return diag.Errorf("Failed to update Datatable Row %s: %s", d.Id(), err)
	}

	log.Printf("Updated Datatable Row %s", d.Id())
	return readArchitectDatatableRow(ctx, d, meta)
}

func deleteArchitectDatatableRow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableId, keyStr := splitDatatableRowId(d.Id())
	if keyStr == "" {
		return diag.Errorf("Invalid Row ID %s", d.Id())
	}

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	log.Printf("Deleting Datatable Row %s", d.Id())
	resp, err := archAPI.DeleteFlowsDatatableRow(tableId, keyStr)
	if err != nil {
		if IsStatus404(resp) {
			// Parent datatable was probably deleted which caused the row to be deleted
			log.Printf("Datatable row already deleted %s", d.Id())
			return nil
		}
		return diag.Errorf("Failed to delete Datatable Row %s: %s", d.Id(), err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := archAPI.GetFlowsDatatableRow(tableId, keyStr, false)
		if err != nil {
			if IsStatus404(resp) {
				// Datatable deleted
				log.Printf("Deleted datatable row %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting datatable row %s: %s", d.Id(), err))
		}
		return retry.RetryableError(fmt.Errorf("Datatable row %s still exists", d.Id()))
	})
}

func buildSdkRowPropertyMap(propertiesJson string, keyStr string) (map[string]interface{}, diag.Diagnostics) {
	// Property value must be empty or a JSON object (map)
	propMap := map[string]interface{}{}
	if propertiesJson != "" {
		if err := json.Unmarshal([]byte(propertiesJson), &propMap); err != nil {
			return nil, diag.Errorf("Error parsing properties_json value %s: %v", propertiesJson, err)
		}
	}
	// Set the key value
	propMap["key"] = keyStr
	return propMap, nil
}

func customizeDatatableRowDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	archAPI := platformclientv2.NewArchitectApiWithConfig(sdkConfig)

	// Retrieve defaults from the datatable for this row
	datatable, getErr := getArchitectDatatableCached(tableId, archAPI)
	if getErr != nil {
		return fmt.Errorf("Failed to read datatable %s: %s", tableId, getErr)
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

// Prevent getting the datatable schema on every row diff
// by caching the results for the duration of the TF run
var archDatatableCache sync.Map

func getArchitectDatatableCached(tableID string, archAPI *platformclientv2.ArchitectApi) (*Datatable, error) {
	if table, ok := archDatatableCache.Load(tableID); ok {
		return table.(*Datatable), nil
	}

	datatable, _, getErr := sdkGetArchitectDatatable(tableID, "schema", archAPI)
	if getErr != nil {
		return nil, fmt.Errorf("Failed to read datatable %s: %s", tableID, getErr)
	}
	archDatatableCache.Store(tableID, datatable)
	return datatable, nil
}
