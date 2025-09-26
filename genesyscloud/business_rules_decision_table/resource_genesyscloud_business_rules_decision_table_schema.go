package business_rules_decision_table

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
resource_genesyscloud_business_rules_decision_table_schema.go holds four functions within it:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the business_rules_decision_table resource.
3.  The datasource schema definitions for the business_rules_decision_table datasource.
4.  The resource exporter configuration for the business_rules_decision_table exporter.
*/
const ResourceType = "genesyscloud_business_rules_decision_table"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceBusinessRulesDecisionTable())
	regInstance.RegisterDataSource(ResourceType, DataSourceBusinessRulesDecisionTable())
	regInstance.RegisterExporter(ResourceType, BusinessRulesDecisionTableExporter())
}

// Schema for contractual blocks (used in inputs)
func contractualSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: generateContractualSchema(2),
	}
}

// Helper function to generate nested contractual schema
func generateContractualSchema(depth int) map[string]*schema.Schema {
	if depth <= 0 {
		return map[string]*schema.Schema{
			"schema_property_key": {
				Type:     schema.TypeString,
				Required: true,
			},
		}
	}

	return map[string]*schema.Schema{
		"schema_property_key": {
			Type:     schema.TypeString,
			Required: true,
		},
		"contractual": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: generateContractualSchema(depth - 1),
			},
		},
	}
}

// Schema for properties blocks (used in outputs)
func propertiesSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: generatePropertiesSchema(2),
	}
}

// Helper function to generate nested properties schema
func generatePropertiesSchema(depth int) map[string]*schema.Schema {
	if depth <= 0 {
		return map[string]*schema.Schema{
			"schema_property_key": {
				Type:     schema.TypeString,
				Required: true,
			},
		}
	}

	return map[string]*schema.Schema{
		"schema_property_key": {
			Type:     schema.TypeString,
			Required: true,
		},
		"properties": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: generatePropertiesSchema(depth - 1),
			},
		},
	}
}

// Schema for expression blocks (used in inputs)
func expressionSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"contractual": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     &schema.Resource{Schema: contractualSchemaFunc().Schema},
			},
			"comparator": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Equals", "NotEquals", "GreaterThan", "GreaterThanOrEquals", "LessThan", "LessThanOrEquals",
					"StartsWith", "NotStartsWith", "EndsWith", "NotEndsWith", "Contains", "NotContains",
				}, false),
			},
		},
	}
}

// Schema for value blocks (used in outputs)
func valueSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"schema_property_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"properties": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Resource{Schema: propertiesSchemaFunc().Schema},
			},
		},
	}
}

// Schema for defaults_to object
func defaultsToSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"value": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Single string value. For queue columns, can be a UUID.",
			},

			"special": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Special enum value: Wildcard, Null, Empty, CurrentTime.",
				ValidateFunc: validation.StringInSlice([]string{"Wildcard", "Null", "Empty", "CurrentTime"}, false),
			},
		},
	}
}

// Schema for input columns
func inputColumnSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the input column",
			},
			"defaults_to": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        &schema.Resource{Schema: defaultsToSchemaFunc().Schema},
				Description: "Default value configuration. Only one of 'value' or 'special' should be set.",
			},
			"expression": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     &schema.Resource{Schema: expressionSchemaFunc().Schema},
			},
		},
	}
}

// Schema for output columns
func outputColumnSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the output column",
			},
			"defaults_to": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        &schema.Resource{Schema: defaultsToSchemaFunc().Schema},
				Description: "Default value configuration. Only one of 'value' or 'special' should be set.",
			},
			"value": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     &schema.Resource{Schema: valueSchemaFunc().Schema},
			},
		},
	}
}

// Schema for columns block
func columnsSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"inputs": {
				Description: "Input columns for the decision table",
				Required:    true,
				Type:        schema.TypeList,
				MinItems:    1,
				Elem:        &schema.Resource{Schema: inputColumnSchemaFunc().Schema},
			},
			"outputs": {
				Description: "Output columns for the decision table",
				Required:    true,
				Type:        schema.TypeList,
				MinItems:    1,
				Elem:        &schema.Resource{Schema: outputColumnSchemaFunc().Schema},
			},
		},
	}
}

// ResourceBusinessRulesDecisionTable registers the genesyscloud_business_rules_decision_table resource with Terraform
func ResourceBusinessRulesDecisionTable() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud business rules decision table. Creates version 1 automatically with the specified columns. Columns cannot be modified after creation - requires resource recreation.`,

		CreateContext: provider.CreateWithPooledClient(createBusinessRulesDecisionTable),
		ReadContext:   provider.ReadWithPooledClient(readBusinessRulesDecisionTable),
		UpdateContext: provider.UpdateWithPooledClient(updateBusinessRulesDecisionTable),
		DeleteContext: provider.DeleteWithPooledClient(deleteBusinessRulesDecisionTable),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "The name of the decision table",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},

			"description": {
				Description: "The decision table description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"division_id": {
				Description: "The ID of the division the decision table belongs to",
				Type:        schema.TypeString,
				Required:    true,
			},

			"schema_id": {
				Description: "The ID of the rules schema used by the decision table",
				Type:        schema.TypeString,
				Required:    true,
			},

			"columns": {
				Description: "Columns for the decision table. Cannot be modified after creation - requires resource recreation.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        columnsSchemaFunc(),
			},

			"rows": {
				Description: "Decision table rows containing input conditions and output actions. Rows are added to the latest draft version and published automatically. At least one row is required to publish the table.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem:        rowSchemaFunc(),
			},

			"version": {
				Description: "Current version number of the decision table.",
				Type:        schema.TypeInt,
				Computed:    true,
			},

			"status": {
				Description: "Current status of the decision table (Draft, Published, etc.).",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
		CustomizeDiff: validateRowsAgainstColumns,
	}
}

// QueueIdResolver is a custom resolver that intelligently converts queue UUIDs to references
// for both column defaults and row values when they contain actual queue IDs
func QueueIdResolver(configMap map[string]interface{}, exporters map[string]*resourceExporter.ResourceExporter, resourceType string) error {
	// Check if this is a queue-related column by looking at the schema_property_key
	// We need to examine the parent column structure to determine this

	// For now, we'll implement a simple check - if the value looks like a UUID and
	// we have routing queue exporters available, we'll attempt to convert it

	value, ok := configMap["value"].(string)
	if !ok || value == "" {
		return nil // No value to convert
	}

	// Check if this looks like a UUID (basic validation)
	if len(value) != 36 || !strings.Contains(value, "-") {
		return nil // Not a UUID, don't convert
	}

	// Check if we have routing queue exporters
	if exporter, ok := exporters["genesyscloud_routing_queue"]; ok {
		// Try to find the queue in the exporter's sanitized resource map
		if queueExport, exists := exporter.SanitizedResourceMap[value]; exists && queueExport != nil {
			exportId := queueExport.BlockLabel
			configMap["value"] = fmt.Sprintf("${genesyscloud_routing_queue.%s.id}", exportId)
		}
	}

	return nil
}

// Schema for literal values (used in both inputs and outputs)
// IMPORTANT: Only ONE of the following fields should be set per literal value
func literalValueSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"value": {
				Description: `The literal value. IMPORTANT: All values must be wrapped in quotes, even numbers and booleans.

Examples:
- String: "VIP", "Hello World"
- Integer: "42", "0", "-10"
- Number: "3.14", "0.0", "-1.5"
- Boolean: "true", "false"
- Date: "2023-01-01"
- DateTime: "2023-01-01T12:00:00.000Z"
- Special: "Wildcard", "Null", "Empty", "CurrentTime"`,
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(i interface{}, k string) (warnings []string, errors []error) {
					value := i.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("value cannot be empty"))
						return warnings, errors
					}
					return warnings, errors
				},
			},
			"type": {
				Description:  "The type of the literal value.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"string", "integer", "number", "date", "datetime", "boolean", "special"}, false),
			},
		},
	}
}

// Schema for decision table rows
func rowSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"row_id": {
				Description: "Unique identifier for this row within the decision table. Auto-generated by the system.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"row_index": {
				Description: "The position of this row in the decision table (1-based). Auto-generated by the system.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"inputs": {
				Description: "Input values (conditions) for this decision row. Each input specifies which column it belongs to using schema_property_key and optionally comparator.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"column_id": {
							Description: "The unique identifier of the input column. Auto-generated by the system.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"schema_property_key": {
							Description: "The schema property key that identifies which input column this value belongs to.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"comparator": {
							Description: "The comparator for this input column. Required when multiple columns have the same schema_property_key with different comparators. Optional when only one column exists for the schema_property_key.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"literal": {
							Description: "The literal value for this input parameter",
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Elem:        literalValueSchemaFunc(),
						},
					},
				},
			},
			"outputs": {
				Description: "Output values (actions) for this decision row. Each output specifies which column it belongs to using schema_property_key and optionally comparator.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"column_id": {
							Description: "The unique identifier of the output column. Auto-generated by the system.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"schema_property_key": {
							Description: "The schema property key that identifies which output column this value belongs to.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"comparator": {
							Description: "The comparator for this output column. Required when multiple columns have the same schema_property_key with different comparators. Optional when only one column exists for the schema_property_key.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"literal": {
							Description: "The literal value for this output parameter. Only ONE field should be set per literal value",
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Elem:        literalValueSchemaFunc(),
						},
					},
				},
			},
		},
	}
}

// validateRowsAgainstColumns validates that row schema property keys exist in column definitions
func validateRowsAgainstColumns(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	// Get columns configuration
	columnsInterface := diff.Get("columns")
	if columnsInterface == nil {
		return nil // No columns to validate against
	}

	columnsList, ok := columnsInterface.([]interface{})
	if !ok || len(columnsList) == 0 {
		return nil // No columns to validate against
	}

	columnsMap, ok := columnsList[0].(map[string]interface{})
	if !ok {
		return nil // Invalid columns structure
	}

	// Convert Terraform columns to SDK format for validation
	sdkColumns, err := convertTerraformColumnsToSDK(columnsMap)
	if err != nil {
		return fmt.Errorf("failed to convert columns for validation: %s", err)
	}

	// Get rows configuration
	rowsInterface := diff.Get("rows")
	if rowsInterface == nil {
		return nil // No rows to validate
	}

	rowsList, ok := rowsInterface.([]interface{})
	if !ok {
		return nil // Invalid rows structure
	}

	// Validate rows against columns
	if err := validateSchemaPropertyKeys(sdkColumns, rowsList); err != nil {
		return fmt.Errorf("row validation failed: %s", err)
	}

	return nil
}

// BusinessRulesDecisionTableExporter returns the resourceExporter object used to hold the genesyscloud_business_rules_decision_table exporter's config
func BusinessRulesDecisionTableExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllBusinessRulesDecisionTables),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
			"schema_id":   {RefType: "genesyscloud_business_rules_schema"},
		},
		// Note: To export routing queue resources that are referenced in decision tables,
		// include "genesyscloud_routing_queue" in the export filter resources.
		// The RefAttrs above will automatically convert division_id and schema_id UUIDs
		// to proper resource references during export.
		// Note: Queue UUIDs in column defaults and row values are intelligently converted to references
		// only when they are actual queue IDs, avoiding false conversions of non-queue values (e.g., priority, skill levels).
		CustomAttributeResolver: map[string]*resourceExporter.RefAttrCustomResolver{
			"columns.outputs.defaults_to.value": {ResolverFunc: QueueIdResolver},
			"columns.inputs.defaults_to.value":  {ResolverFunc: QueueIdResolver},
			"rows.*.inputs.*.literal.value":     {ResolverFunc: QueueIdResolver},
			"rows.*.outputs.*.literal.value":    {ResolverFunc: QueueIdResolver},
		},
	}
}

// DataSourceBusinessRulesDecisionTable registers the genesyscloud_business_rules_decision_table data source
func DataSourceBusinessRulesDecisionTable() *schema.Resource {
	return &schema.Resource{
		Description: `Genesys Cloud business rules decision table data source. Select a business rules decision table by its name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceBusinessRulesDecisionTableRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `business rules decision table name`,
				Type:        schema.TypeString,
				Required:    true,
			},
			"version": {
				Description: `The published version of the decision table.`,
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}
