package architect_datatable_row

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

type Datatableproperty struct {
	Id           *string      `json:"$id,omitempty"`
	VarType      *string      `json:"type,omitempty"`
	Title        *string      `json:"title,omitempty"`
	Default      *interface{} `json:"default,omitempty"`
	DisplayOrder *int         `json:"displayOrder,omitempty"`
}

// Overriding the SDK Datatable document as it does not allow setting additionalProperties to 'false' as required by the API
type Jsonschemadocument struct {
	Schema               *string                       `json:"$schema,omitempty"`
	VarType              *string                       `json:"type,omitempty"`
	Required             *[]string                     `json:"required,omitempty"`
	Properties           *map[string]Datatableproperty `json:"properties,omitempty"`
	AdditionalProperties *interface{}                  `json:"additionalProperties,omitempty"`
}

type Datatable struct {
	Id          *string                            `json:"id,omitempty"`
	Name        *string                            `json:"name,omitempty"`
	Description *string                            `json:"description,omitempty"`
	Division    *platformclientv2.Writabledivision `json:"division,omitempty"`
	Schema      *Jsonschemadocument                `json:"schema,omitempty"`
}

func getAllArchitectDatatableRows(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	archProxy := getArchitectDatatableRowProxy(clientConfig)

	tables, resp, err := archProxy.getAllArchitectDatatable(ctx)
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get architect datatables error: %s", err), resp)
	}

	for _, tableMeta := range *tables {
		rows, resp, err := archProxy.getAllArchitectDatatableRows(ctx, *tableMeta.Id)

		if err != nil {
			return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get architect Datatable Rows error: %s", err), resp)
		}

		for _, row := range *rows {
			if keyVal, ok := row["key"]; ok {
				keyStr := keyVal.(string) // Keys must be strings
				resources[createDatatableRowId(*tableMeta.Id, keyStr)] = &resourceExporter.ResourceMeta{BlockLabel: *tableMeta.Name + "_" + keyStr}
			}
		}
	}

	return resources, nil
}

func createArchitectDatatableRow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableId := d.Get("datatable_id").(string)
	keyStr := d.Get("key_value").(string)
	properties := d.Get("properties_json").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	archProxy := getArchitectDatatableRowProxy(sdkConfig)

	rowMap, diagErr := buildSdkRowPropertyMap(properties, keyStr)
	if diagErr != nil {
		return diagErr
	}

	rowId := createDatatableRowId(tableId, keyStr)
	log.Printf("Creating Datatable Row %s", rowId)

	_, resp, err := archProxy.createArchitectDatatableRow(ctx, tableId, &rowMap)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create Datatable Row %s error: %s", d.Id(), err), resp)
	}

	d.SetId(rowId)

	log.Printf("Created Datatable Row %s", d.Id())
	return readArchitectDatatableRow(ctx, d, meta)
}

func readArchitectDatatableRow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableId, keyStr := splitDatatableRowId(d.Id())
	if keyStr == "" {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Invalid Row ID %s", d.Id()), fmt.Errorf("keyStr is nil"))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	archProxy := getArchitectDatatableRowProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceArchitectDatatableRow(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading Datatable Row %s", d.Id())

	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		row, resp, getErr := archProxy.getArchitectDatatableRow(ctx, tableId, keyStr)
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Datatable Row %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read Datatable Row %s | error: %s", d.Id(), getErr), resp))
		}

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
		return cc.CheckState(d)
	})
}

func updateArchitectDatatableRow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableId := d.Get("datatable_id").(string)
	keyStr := d.Get("key_value").(string)
	properties := d.Get("properties_json").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	archProxy := getArchitectDatatableRowProxy(sdkConfig)

	rowMap, diagErr := buildSdkRowPropertyMap(properties, keyStr)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updating Datatable Row %s", d.Id())

	_, resp, err := archProxy.updateArchitectDatatableRow(ctx, tableId, keyStr, &rowMap)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update Datatable Row %s error: %s", d.Id(), err), resp)
	}

	log.Printf("Updated Datatable Row %s", d.Id())
	return readArchitectDatatableRow(ctx, d, meta)
}

func deleteArchitectDatatableRow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableId, keyStr := splitDatatableRowId(d.Id())
	if keyStr == "" {
		return util.BuildDiagnosticError(ResourceType, fmt.Sprintf("Invalid Row ID %s", d.Id()), fmt.Errorf("keyStr is nil"))
	}

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	archProxy := getArchitectDatatableRowProxy(sdkConfig)

	log.Printf("Deleting Datatable Row %s", d.Id())
	resp, err := archProxy.deleteArchitectDatatableRow(ctx, tableId, keyStr)
	if err != nil {
		if util.IsStatus404(resp) {
			// Parent architect_datatable was probably deleted which caused the row to be deleted
			log.Printf("Datatable row already deleted %s", d.Id())
			return nil
		}
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete Datatable Row %s error: %s", d.Id(), err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := archProxy.getArchitectDatatableRow(ctx, tableId, keyStr)
		if err != nil {
			if util.IsStatus404(resp) {
				// Datatable deleted
				log.Printf("Deleted architect_datatable row %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting architect_datatable row %s | error: %s", d.Id(), err), resp))
		}
		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Datatable row %s still exists", d.Id()), resp))
	})
}
