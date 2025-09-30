---
page_title: "genesyscloud_business_rules_decision_table Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud business rules decision table. Creates version 1 automatically with the specified columns. Columns cannot be modified after creation - requires resource recreation.
---
# genesyscloud_business_rules_decision_table (Resource)

Genesys Cloud business rules decision table. Creates version 1 automatically with the specified columns. Columns cannot be modified after creation - requires resource recreation.

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [POST /api/v2/businessrules/decisiontables](https://developer.genesys.cloud/platform/preview-apis#post-api-v2-businessrules-decisiontables)
* [GET /api/v2/businessrules/decisiontables/{tableId}](https://developer.genesys.cloud/platform/preview-apis#get-api-v2-businessrules-decisiontables--tableId-)
* [PATCH /api/v2/businessrules/decisiontables/{tableId}](https://developer.genesys.cloud/platform/preview-apis#patch-api-v2-businessrules-decisiontables--tableId-)
* [DELETE /api/v2/businessrules/decisiontables/{tableId}](https://developer.genesys.cloud/platform/preview-apis#delete-api-v2-businessrules-decisiontables--tableId-)
* [GET /api/v2/businessrules/decisiontables/search](https://developer.genesys.cloud/platform/preview-apis#get-api-v2-businessrules-decisiontables-search)
* [GET /api/v2/businessrules/decisiontables/{tableId}/versions/{tableVersion}](https://developer.genesys.cloud/platform/preview-apis#get-api-v2-businessrules-decisiontables--tableId--versions--tableVersion-)
* [POST /api/v2/businessrules/decisiontables/{tableId}/versions](https://developer.genesys.cloud/platform/preview-apis#post-api-v2-businessrules-decisiontables--tableId--versions)
* [PUT /api/v2/businessrules/decisiontables/{tableId}/versions/{tableVersion}/publish](https://developer.genesys.cloud/platform/preview-apis#put-api-v2-businessrules-decisiontables--tableId--versions--tableVersion--publish)
* [DELETE /api/v2/businessrules/decisiontables/{tableId}/versions/{tableVersion}](https://developer.genesys.cloud/platform/preview-apis#delete-api-v2-businessrules-decisiontables--tableId--versions--tableVersion-)
* [GET /api/v2/businessrules/decisiontables/{tableId}/versions/{tableVersion}/rows](https://developer.genesys.cloud/platform/preview-apis#get-api-v2-businessrules-decisiontables--tableId--versions--tableVersion--rows)
* [POST /api/v2/businessrules/decisiontables/{tableId}/versions/{tableVersion}/rows](https://developer.genesys.cloud/platform/preview-apis#post-api-v2-businessrules-decisiontables--tableId--versions--tableVersion--rows)
* [PUT /api/v2/businessrules/decisiontables/{tableId}/versions/{tableVersion}/rows/{rowId}](https://developer.genesys.cloud/platform/preview-apis#put-api-v2-businessrules-decisiontables--tableId--versions--tableVersion--rows--rowId-)
* [DELETE /api/v2/businessrules/decisiontables/{tableId}/versions/{tableVersion}/rows/{rowId}](https://developer.genesys.cloud/platform/preview-apis#delete-api-v2-businessrules-decisiontables--tableId--versions--tableVersion--rows--rowId-)


## Example Usage

```terraform
resource "genesyscloud_business_rules_decision_table" "example_decision_table" {
  name        = "Example Decision Table"
  description = "Example Decision Table created by terraform"
  division_id = data.genesyscloud_auth_division_home.home.id
  schema_id   = genesyscloud_business_rules_schema.example_business_rules_schema.id

  columns {
    inputs {
      expression {
        defaults_to {
          value = "anything"
        }
        contractual {
          schema_property_key = "custom_attribute_string"
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        special = "Wildcard"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_enum"
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        value = "5"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_integer"
        }
        comparator = "GreaterThan"
      }
    }

    inputs {
      defaults_to {
        value = "1000.0"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_number"
        }
        comparator = "LessThanOrEquals"
      }
    }

    inputs {
      defaults_to {
        value = "true"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_boolean"
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        special = "CurrentTime"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_date"
        }
        comparator = "GreaterThanOrEquals"
      }
    }

    inputs {
      defaults_to {
        special = "CurrentTime"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_date"
        }
        comparator = "LessThanOrEquals"
      }
    }

    inputs {
      defaults_to {
        special = "CurrentTime"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_datetime"
        }
        comparator = "NotEquals"
      }
    }

    inputs {
      defaults_to {
        special = "Wildcard"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_for_empty_literal_block"
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        special = "Wildcard"
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_for_empty_literal_value_type"
        }
        comparator = "Equals"
      }
    }

    inputs {
      defaults_to {
        value = data.genesyscloud_routing_queue.standard_queue.id
      }
      expression {
        contractual {
          schema_property_key = "custom_attribute_queue"
          contractual {
            schema_property_key = "queue"
            contractual {
              schema_property_key = "id"
            }
          }
        }
        comparator = "Equals"
      }
    }

    outputs {
      defaults_to {
        value = data.genesyscloud_routing_queue.vip_queue.id
      }
      value {
        schema_property_key = "custom_attribute_queue"
        properties {
          schema_property_key = "queue"
          properties {
            schema_property_key = "id"
          }
        }
      }
    }

    outputs {
      defaults_to {
        value = "Premium Support"
      }
      value {
        schema_property_key = "custom_attribute_string"
      }
    }

    outputs {
      defaults_to {
        special = "Null"
      }
      value {
        schema_property_key = "custom_attribute_enum"
      }
    }
  }

  rows {
    inputs {
      literal {
        value = "John Doe"
        type  = "string"
      }
    }
    inputs {
      literal {
        value = "option_1"
        type  = "string"
      }
    }
    inputs {
      literal {
        value = "85"
        type  = "integer"
      }
    }
    inputs {
      literal {
        value = "15000.0"
        type  = "number"
      }
    }
    inputs {
      literal {
        value = "true"
        type  = "boolean"
      }
    }
    inputs {
      literal {
        value = "2023-01-01"
        type  = "date"
      }
    }
    inputs {
      literal {
        value = "2023-12-31"
        type  = "date"
      }
    }
    inputs {
      literal {
        value = "2023-12-01T10:30:00.000Z"
        type  = "datetime"
      }
    }

    inputs {
      literal {} // Empty literal block with no value or type specified to use column default and must be provided
    }

    inputs {
      literal { // Literal block with empty value and type to use column default and must be provided
        value = ""
        type  = ""
      }
    }
    inputs {
      literal {
        value = data.genesyscloud_routing_queue.standard_queue.id
        type  = "string"
      }
    }
    outputs {
      literal {
        value = data.genesyscloud_routing_queue.vip_queue.id
        type  = "string"
      }
    }
    outputs {
      literal {
        value = "Premium Support"
        type  = "string"
      }
    }
    outputs {
      literal {
        value = "option_2"
        type  = "string"
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `columns` (Block List, Min: 1, Max: 1) Columns for the decision table. Cannot be modified after creation - requires resource recreation.

Note: The order of input and output columns defines the positional mapping for row values. Row inputs/outputs must be provided in the same order as their corresponding columns. (see [below for nested schema](#nestedblock--columns))
- `division_id` (String) The ID of the division the decision table belongs to.
- `name` (String) The decision table name.
- `rows` (Block List, Min: 1) Decision table rows containing input conditions and output results. Rows are added to the latest draft version and published automatically. At least one row is required to publish the table.

IMPORTANT: Row inputs and outputs must follow the same positional order as defined in the columns. The first input/output corresponds to the first column, second to second column, etc. (see [below for nested schema](#nestedblock--rows))
- `schema_id` (String) The ID of the rules schema used by the decision table.

### Optional

- `description` (String) The decision table description.

### Read-Only

- `id` (String) The ID of this resource.
- `version` (Number) Current version number of this published decision table.

<a id="nestedblock--columns"></a>
### Nested Schema for `columns`

Required:

- `inputs` (Block List, Min: 1) The input columns for the decision table (see [below for nested schema](#nestedblock--columns--inputs))
- `outputs` (Block List, Min: 1) The output columns for the decision table (see [below for nested schema](#nestedblock--columns--outputs))

<a id="nestedblock--columns--inputs"></a>
### Nested Schema for `columns.inputs`

Required:

- `defaults_to` (Block List, Min: 1, Max: 1) Default value configuration. Only one of 'value' or 'special' should be set. (see [below for nested schema](#nestedblock--columns--inputs--defaults_to))
- `expression` (Block List, Min: 1, Max: 1) The input column condition expression, comprising the left side and comparator of a logical condition in the form of left|comparator|right, where each row of the decision table will provide the right side to form a complete condition. (see [below for nested schema](#nestedblock--columns--inputs--expression))

Read-Only:

- `id` (String) The ID of the column

<a id="nestedblock--columns--inputs--defaults_to"></a>
### Nested Schema for `columns.inputs.defaults_to`

Optional:

- `special` (String) A default special value enum for this column.Valid values: Wildcard, Null, Empty, CurrentTime.
- `value` (String) A default string value for this column, will be cast to appropriate type according to the relevant contract schema property.


<a id="nestedblock--columns--inputs--expression"></a>
### Nested Schema for `columns.inputs.expression`

Required:

- `comparator` (String) A comparator used to join the left and right sides of a logical condition. Valid values: Equals, NotEquals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, StartsWith, NotStartsWith, EndsWith, NotEndsWith, Contains, NotContains.
- `contractual` (Block List, Min: 1, Max: 1) A value that is defined by a contract schema and used to form the left side of a logical condition. (see [below for nested schema](#nestedblock--columns--inputs--expression--contractual))

<a id="nestedblock--columns--inputs--expression--contractual"></a>
### Nested Schema for `columns.inputs.expression.contractual`

Required:

- `schema_property_key` (String) The contract schema property key that describes the input value of this column.

Optional:

- `contractual` (Block List, Max: 1) The nested contractual definition that is defined by a contract schema, if any. (see [below for nested schema](#nestedblock--columns--inputs--expression--contractual--contractual))

<a id="nestedblock--columns--inputs--expression--contractual--contractual"></a>
### Nested Schema for `columns.inputs.expression.contractual.contractual`

Required:

- `schema_property_key` (String) The contract schema property key that describes the input value of this column.

Optional:

- `contractual` (Block List, Max: 1) The nested contractual definition that is defined by a contract schema, if any. (see [below for nested schema](#nestedblock--columns--inputs--expression--contractual--contractual--contractual))

<a id="nestedblock--columns--inputs--expression--contractual--contractual--contractual"></a>
### Nested Schema for `columns.inputs.expression.contractual.contractual.contractual`

Required:

- `schema_property_key` (String) The contract schema property key that describes the input value of this column.






<a id="nestedblock--columns--outputs"></a>
### Nested Schema for `columns.outputs`

Required:

- `defaults_to` (Block List, Min: 1, Max: 1) Default value configuration. Only one of 'value' or 'special' should be set. (see [below for nested schema](#nestedblock--columns--outputs--defaults_to))
- `value` (Block List, Min: 1, Max: 1) The output data of this column that will be provided by each row. (see [below for nested schema](#nestedblock--columns--outputs--value))

Read-Only:

- `id` (String) The ID of the column

<a id="nestedblock--columns--outputs--defaults_to"></a>
### Nested Schema for `columns.outputs.defaults_to`

Optional:

- `special` (String) A default special value enum for this column.Valid values: Wildcard, Null, Empty, CurrentTime.
- `value` (String) A default string value for this column, will be cast to appropriate type according to the relevant contract schema property.


<a id="nestedblock--columns--outputs--value"></a>
### Nested Schema for `columns.outputs.value`

Required:

- `schema_property_key` (String) The contract schema property key that describes the output value of this column

Optional:

- `properties` (Block List) The nested properties that are defined by a contract schema, if any. (see [below for nested schema](#nestedblock--columns--outputs--value--properties))

<a id="nestedblock--columns--outputs--value--properties"></a>
### Nested Schema for `columns.outputs.value.properties`

Required:

- `schema_property_key` (String) The contract schema property key that describes the nested property value.

Optional:

- `properties` (Block List) The nested properties that are defined by a contract schema, if any. (see [below for nested schema](#nestedblock--columns--outputs--value--properties--properties))

<a id="nestedblock--columns--outputs--value--properties--properties"></a>
### Nested Schema for `columns.outputs.value.properties.properties`

Required:

- `schema_property_key` (String) The contract schema property key that describes the nested property value.

Optional:

- `properties` (Block List) The nested properties that are defined by a contract schema, if any. (see [below for nested schema](#nestedblock--columns--outputs--value--properties--properties--properties))

<a id="nestedblock--columns--outputs--value--properties--properties--properties"></a>
### Nested Schema for `columns.outputs.value.properties.properties.properties`

Required:

- `schema_property_key` (String) The contract schema property key that describes the nested property value.







<a id="nestedblock--rows"></a>
### Nested Schema for `rows`

Optional:

- `inputs` (Block List) Input values (conditions) for this decision row. Values are matched to input columns by position (index) - first input corresponds to first input column, second to second, etc. Missing values will use column defaults provided by the decision table columns defaults_to field. (see [below for nested schema](#nestedblock--rows--inputs))
- `outputs` (Block List) Output values (results) for this decision row. Values are matched to output columns by position (index) - first output corresponds to first output column, second to second, etc. Missing values will use column defaults provided by the decision table columns defaults_to field. (see [below for nested schema](#nestedblock--rows--outputs))

Read-Only:

- `row_id` (String) Unique identifier for this row within the decision table. Auto-generated by the system.
- `row_index` (Number) The absolute index of this row in the decision table, starting at 1. Auto-generated by the system.

<a id="nestedblock--rows--inputs"></a>
### Nested Schema for `rows.inputs`

Required:

- `literal` (Block List, Min: 1, Max: 1) The literal value for this parameter. Use an empty block {} or empty values (value = "", type = "") to use column default. (see [below for nested schema](#nestedblock--rows--inputs--literal))

Read-Only:

- `column_id` (String) The unique identifier of the column. Auto-generated by the system.

<a id="nestedblock--rows--inputs--literal"></a>
### Nested Schema for `rows.inputs.literal`

Optional:

- `type` (String) The type of the literal value. Set to empty string "" to use column default.

								Supported types:
								- string: A string value
								- integer: A positive or negative whole number, including zero
								- number: A positive or negative decimal number, including zero
								- date: A date value, must be in the format of yyyy-MM-dd, e.g. 2024-09-23. Dates are represented as an ISO-8601 string
								- datetime: A date time value, must be in the format of yyyy-MM-dd'T'HH:mm:ss.SSSZ, e.g. 2024-10-02T01:01:01.111Z. Date time is represented as an ISO-8601 string
								- special: A special value enum, such as Wildcard, Null, etc. Valid values: Wildcard, Null, Empty, CurrentTime
								- boolean: A boolean value
								- "": An empty string "" to use column default
- `value` (String) The literal value. IMPORTANT: All values must be wrapped in quotes, even numbers and booleans.
								Set to empty string "" to use column default.

								Examples:
								- String: "VIP", "Hello World"
								- Integer: "42", "0", "-10"
								- Number: "3.14", "0.0", "-1.5" (formatting differences like "1.0" vs "1" are automatically handled)
								- Boolean: "true", "false"
								- Date: "2023-01-01"
								- DateTime: "2023-01-01T12:00:00.000Z"
								- Special: "Wildcard", "Null", "Empty", "CurrentTime"
								- Default: Empty string "" uses column default Defaults to ``.



<a id="nestedblock--rows--outputs"></a>
### Nested Schema for `rows.outputs`

Required:

- `literal` (Block List, Min: 1, Max: 1) The literal value for this parameter. Use an empty block {} or empty values (value = "", type = "") to use column default. (see [below for nested schema](#nestedblock--rows--outputs--literal))

Read-Only:

- `column_id` (String) The unique identifier of the column. Auto-generated by the system.

<a id="nestedblock--rows--outputs--literal"></a>
### Nested Schema for `rows.outputs.literal`

Optional:

- `type` (String) The type of the literal value. Set to empty string "" to use column default.

								Supported types:
								- string: A string value
								- integer: A positive or negative whole number, including zero
								- number: A positive or negative decimal number, including zero
								- date: A date value, must be in the format of yyyy-MM-dd, e.g. 2024-09-23. Dates are represented as an ISO-8601 string
								- datetime: A date time value, must be in the format of yyyy-MM-dd'T'HH:mm:ss.SSSZ, e.g. 2024-10-02T01:01:01.111Z. Date time is represented as an ISO-8601 string
								- special: A special value enum, such as Wildcard, Null, etc. Valid values: Wildcard, Null, Empty, CurrentTime
								- boolean: A boolean value
								- "": An empty string "" to use column default
- `value` (String) The literal value. IMPORTANT: All values must be wrapped in quotes, even numbers and booleans.
								Set to empty string "" to use column default.

								Examples:
								- String: "VIP", "Hello World"
								- Integer: "42", "0", "-10"
								- Number: "3.14", "0.0", "-1.5" (formatting differences like "1.0" vs "1" are automatically handled)
								- Boolean: "true", "false"
								- Date: "2023-01-01"
								- DateTime: "2023-01-01T12:00:00.000Z"
								- Special: "Wildcard", "Null", "Empty", "CurrentTime"
								- Default: Empty string "" uses column default Defaults to ``.

