package business_rules_decision_table

import (
	"fmt"
	"strconv"

	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

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
				Type:        schema.TypeString,
				Required:    true,
				Description: "The contract schema property key that describes the input value of this column.",
			},
		}
	}

	return map[string]*schema.Schema{
		"schema_property_key": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The contract schema property key that describes the input value of this column.",
		},
		"contractual": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: generateContractualSchema(depth - 1),
			},
			Description: "The nested contractual definition that is defined by a contract schema, if any.",
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
				Type:        schema.TypeString,
				Required:    true,
				Description: "The contract schema property key that describes the nested property value.",
			},
		}
	}

	return map[string]*schema.Schema{
		"schema_property_key": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The contract schema property key that describes the nested property value.",
		},
		"properties": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: generatePropertiesSchema(depth - 1),
			},
			Description: "The nested properties that are defined by a contract schema, if any.",
		},
	}
}

// Schema for expression blocks (used in inputs)
func expressionSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"contractual": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem:        &schema.Resource{Schema: contractualSchemaFunc().Schema},
				Description: "A value that is defined by a contract schema and used to form the left side of a logical condition.",
			},
			"comparator": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Equals", "NotEquals", "GreaterThan", "GreaterThanOrEquals", "LessThan", "LessThanOrEquals",
					"StartsWith", "NotStartsWith", "EndsWith", "NotEndsWith", "Contains", "NotContains",
					"ContainsAny", "NotContainsAny", "ContainsAll", "NotContainsAll", "ContainsExactly", "NotContainsExactly",
					"ContainsSequence", "NotContainsSequence", "IsSubset", "NotIsSubset", "IsSubsequence", "NotIsSubsequence",
				}, false),
				Description: "A comparator used to join the left and right sides of a logical condition. Valid values: Equals, " +
					"NotEquals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, StartsWith, NotStartsWith, EndsWith, " +
					"NotEndsWith, Contains, NotContains, ContainsAny, NotContainsAny, ContainsAll, NotContainsAll, ContainsExactly, NotContainsExactly, " +
					"ContainsSequence, NotContainsSequence, IsSubset, NotIsSubset, IsSubsequence, NotIsSubsequence.",
			},
		},
	}
}

// Schema for value blocks (used in outputs)
func valueSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"schema_property_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The contract schema property key that describes the output value of this column",
			},
			"properties": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Resource{Schema: propertiesSchemaFunc().Schema},
				Description: "The nested properties that are defined by a contract schema, if any.",
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
				Description: "A default string value for this column, will be cast to appropriate type according to the relevant contract schema property.",
			},

			"values": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A default list of string values for this column. Used for stringList data types.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"special": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "A default special value enum for this column.Valid values: Wildcard, Null, Empty, CurrentTime.",
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
				Description: "The ID of the column",
			},
			"defaults_to": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem:        &schema.Resource{Schema: defaultsToSchemaFunc().Schema},
				Description: "Default value configuration. Only one of 'value' or 'special' should be set.",
			},
			"expression": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem:        &schema.Resource{Schema: expressionSchemaFunc().Schema},
				Description: "The input column condition expression, comprising the left side and comparator of a logical condition in the form of left|comparator|right, where each row of the decision table will provide the right side to form a complete condition.",
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
				Description: "The ID of the column",
			},
			"defaults_to": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem:        &schema.Resource{Schema: defaultsToSchemaFunc().Schema},
				Description: "Default value configuration. Only one of 'value' or 'special' should be set.",
			},
			"value": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem:        &schema.Resource{Schema: valueSchemaFunc().Schema},
				Description: "The output data of this column that will be provided by each row.",
			},
		},
	}
}

// Schema for columns block
func columnsSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"inputs": {
				Description: "The input columns for the decision table",
				Required:    true,
				Type:        schema.TypeList,
				MinItems:    1,
				Elem:        &schema.Resource{Schema: inputColumnSchemaFunc().Schema},
			},
			"outputs": {
				Description: "The output columns for the decision table",
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
				Description:  "The decision table name.",
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},

			"description": {
				Description: "The decision table description.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			"division_id": {
				Description: "The ID of the division the decision table belongs to.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"schema_id": {
				Description: "The ID of the rules schema used by the decision table.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"columns": {
				Description: "Columns for the decision table. Cannot be modified after creation - requires resource recreation.\n\nNote: The order of input and output columns defines the positional mapping for row values. Row inputs/outputs must be provided in the same order as their corresponding columns.",
				Required:    true,
				ForceNew:    true,
				Type:        schema.TypeList,
				MaxItems:    1,
				Elem:        columnsSchemaFunc(),
			},
			"rows": {
				Description: "Decision table rows containing input conditions and output results. Rows are added to the latest draft version and published automatically. At least one row is required to publish the table.\n\nIMPORTANT: Row inputs and outputs must follow the same positional order as defined in the columns. The first input/output corresponds to the first column, second to second column, etc.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem:        rowSchemaFunc(),
			},

			"version": {
				Description: "Current version number of this published decision table.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
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
			"type": {
				Description: `The type of the literal value. Set to empty string "" to use column default.

								Supported types:
								- string: A string value
								- integer: A positive or negative whole number, including zero
								- number: A positive or negative decimal number, including zero
								- date: A date value, must be in the format of yyyy-MM-dd, e.g. 2024-09-23. Dates are represented as an ISO-8601 string
								- datetime: A date time value, must be in the format of yyyy-MM-dd'T'HH:mm:ss.SSSZ, e.g. 2024-10-02T01:01:01.111Z. Date time is represented as an ISO-8601 string
								- special: A special value enum, such as Wildcard, Null, etc. Valid values: Wildcard, Null, Empty, CurrentTime
								- boolean: A boolean value
								- stringList: A list of string values, provided as comma-separated string
								- "": An empty string "" to use column default`,
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"string", "integer", "number", "date", "datetime", "boolean", "special", "stringList", ""}, false),
			},
			"value": {
				Description: `The literal value. IMPORTANT: All values must be wrapped in quotes, even numbers and booleans.
								Set to empty string "" to use column default.

								Examples:
								- String: "VIP", "Hello World"
								- Integer: "42", "0", "-10"
								- Number: "3.14", "0.0", "-1.5" (formatting differences like "1.0" vs "1" are automatically handled)
								- Boolean: "true", "false"
								- Date: "2023-01-01"
								- DateTime: "2023-01-01T12:00:00.000Z"
								- Special: "Wildcard", "Null", "Empty", "CurrentTime"
								- StringList: "item1,item2,item3" (comma-separated string)
								- Default: Empty string "" uses column default`,
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "",
				DiffSuppressFunc: suppressNumberFormattingDiff,
			},
		},
	}
}

// Schema for row parameters (used in both inputs and outputs)
func rowParameterSchemaFunc() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"column_id": {
				Description: "The unique identifier of the column. Auto-generated by the system.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"literal": {
				Description: "The literal value for this parameter. Use an empty block {} or empty values (value = \"\", type = \"\") to use column default.",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem:        literalValueSchemaFunc(),
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
				Description: "The absolute index of this row in the decision table, starting at 1. Auto-generated by the system.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"inputs": {
				Description: "Input values (conditions) for this decision row. Values are matched to input columns by position (index) - first input corresponds to first input column, second to second, etc. Missing values will use column defaults provided by the decision table columns defaults_to field.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        rowParameterSchemaFunc(),
			},
			"outputs": {
				Description: "Output values (results) for this decision row. Values are matched to output columns by position (index) - first output corresponds to first output column, second to second, etc. Missing values will use column defaults provided by the decision table columns defaults_to field.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        rowParameterSchemaFunc(),
			},
		},
	}
}

// BusinessRulesDecisionTableExporter returns the resourceExporter object used to hold the genesyscloud_business_rules_decision_table exporter's config
func BusinessRulesDecisionTableExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllBusinessRulesDecisionTables),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"division_id": {RefType: "genesyscloud_auth_division"},
			"schema_id":   {RefType: "genesyscloud_business_rules_schema"},
		},
		ExcludedAttributes: []string{
			"version",
			"columns.inputs.id",
			"columns.outputs.id",
			"rows.inputs.column_id",
			"rows.outputs.column_id",
			"rows.row_id",
			"rows.row_index",
		},
		// Note: To export routing queue resources that are referenced in decision tables,
		// include "genesyscloud_routing_queue" in the export filter resources.
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
				Description: "The decision table name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"version": {
				Description: "The published version of the decision table.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
		},
	}
}

// suppressNumberFormattingDiff suppresses diffs when the only difference is number formatting
// (e.g., "1.0" vs "1", "1.00" vs "1", etc.) for number type literals.
// This allows users to write "1.0" in their Terraform config while the API returns "1",
// without causing unnecessary plan diffs.
func suppressNumberFormattingDiff(k, old, new string, d *schema.ResourceData) bool {
	// Get the type field from the same resource path
	typeKey := strings.Replace(k, ".value", ".type", 1)
	literalType := d.Get(typeKey).(string)

	// Only suppress diffs for number type literals
	if literalType != "number" {
		return false
	}

	// If either value is empty, don't suppress (let normal diff handling work)
	if old == "" || new == "" {
		return false
	}

	// Parse both values as floats
	oldFloat, oldErr := strconv.ParseFloat(old, 64)
	newFloat, newErr := strconv.ParseFloat(new, 64)

	// If either parsing fails, don't suppress (let normal diff handling work)
	if oldErr != nil || newErr != nil {
		return false
	}

	// Suppress diff if the numeric values are equal
	return oldFloat == newFloat
}
