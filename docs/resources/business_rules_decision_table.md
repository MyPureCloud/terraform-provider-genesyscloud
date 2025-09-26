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
        contractual {
          schema_property_key = "custom_attribute_string"
        }
        comparator = "Equals"
      }
    }

    inputs {
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
        special = "Null"
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
      value {
        schema_property_key = "custom_attribute_enum"
      }
    }
  }

  rows {
    inputs {
      schema_property_key = "custom_attribute_string"
      literal {
        value = "John Doe"
        type  = "string"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_enum"
      literal {
        value = "option_1"
        type  = "string"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_integer"
      literal {
        value = "85"
        type  = "integer"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_number"
      literal {
        value = "15000.0"
        type  = "number"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_boolean"
      literal {
        value = "true"
        type  = "boolean"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_date"
      comparator          = "GreaterThanOrEquals"
      literal {
        value = "2023-01-01"
        type  = "date"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_date"
      comparator          = "LessThanOrEquals"
      literal {
        value = "2023-12-31"
        type  = "date"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_datetime"
      literal {
        value = "2023-12-01T10:30:00.000Z"
        type  = "datetime"
      }
    }
    inputs {
      schema_property_key = "custom_attribute_queue"
      literal {
        value = data.genesyscloud_routing_queue.standard_queue.id
        type  = "string"
      }
    }
    outputs {
      schema_property_key = "custom_attribute_queue"
      literal {
        value = data.genesyscloud_routing_queue.vip_queue.id
        type  = "string"
      }
    }
    outputs {
      schema_property_key = "custom_attribute_string"
      literal {
        value = "Premium Support"
        type  = "string"
      }
    }
    outputs {
      schema_property_key = "custom_attribute_enum"
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

- `columns` (Block List, Min: 1, Max: 1) Columns for the decision table. Cannot be modified after creation - requires resource recreation. (see [below for nested schema](#nestedblock--columns))
- `division_id` (String) The ID of the division the decision table belongs to
- `name` (String) The name of the decision table
- `rows` (Block List, Min: 1) Decision table rows containing input conditions and output actions. Rows are added to the latest draft version and published automatically. At least one row is required to publish the table. (see [below for nested schema](#nestedblock--rows))
- `schema_id` (String) The ID of the rules schema used by the decision table

### Optional

- `description` (String) The decision table description

### Read-Only

- `id` (String) The ID of this resource.
- `status` (String) Current status of the decision table (Draft, Published, etc.).
- `version` (Number) Current version number of the decision table.

<a id="nestedblock--columns"></a>
### Nested Schema for `columns`

Required:

- `inputs` (Block List, Min: 1) Input columns for the decision table (see [below for nested schema](#nestedblock--columns--inputs))
- `outputs` (Block List, Min: 1) Output columns for the decision table (see [below for nested schema](#nestedblock--columns--outputs))

<a id="nestedblock--columns--inputs"></a>
### Nested Schema for `columns.inputs`

Required:

- `expression` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--columns--inputs--expression))

Optional:

- `defaults_to` (Block List, Max: 1) Default value configuration. Only one of 'value' or 'special' should be set. (see [below for nested schema](#nestedblock--columns--inputs--defaults_to))

Read-Only:

- `id` (String) The ID of the input column

<a id="nestedblock--columns--inputs--expression"></a>
### Nested Schema for `columns.inputs.expression`

Required:

- `comparator` (String)
- `contractual` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--columns--inputs--expression--contractual))

<a id="nestedblock--columns--inputs--expression--contractual"></a>
### Nested Schema for `columns.inputs.expression.contractual`

Required:

- `schema_property_key` (String)

Optional:

- `contractual` (Block List, Max: 1) (see [below for nested schema](#nestedblock--columns--inputs--expression--contractual--contractual))

<a id="nestedblock--columns--inputs--expression--contractual--contractual"></a>
### Nested Schema for `columns.inputs.expression.contractual.contractual`

Required:

- `schema_property_key` (String)

Optional:

- `contractual` (Block List, Max: 1) (see [below for nested schema](#nestedblock--columns--inputs--expression--contractual--contractual--contractual))

<a id="nestedblock--columns--inputs--expression--contractual--contractual--contractual"></a>
### Nested Schema for `columns.inputs.expression.contractual.contractual.contractual`

Required:

- `schema_property_key` (String)





<a id="nestedblock--columns--inputs--defaults_to"></a>
### Nested Schema for `columns.inputs.defaults_to`

Optional:

- `special` (String) Special enum value: Wildcard, Null, Empty, CurrentTime.
- `value` (String) Single string value. For queue columns, can be a UUID.



<a id="nestedblock--columns--outputs"></a>
### Nested Schema for `columns.outputs`

Required:

- `value` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--columns--outputs--value))

Optional:

- `defaults_to` (Block List, Max: 1) Default value configuration. Only one of 'value' or 'special' should be set. (see [below for nested schema](#nestedblock--columns--outputs--defaults_to))

Read-Only:

- `id` (String) The ID of the output column

<a id="nestedblock--columns--outputs--value"></a>
### Nested Schema for `columns.outputs.value`

Required:

- `schema_property_key` (String)

Optional:

- `properties` (Block List) (see [below for nested schema](#nestedblock--columns--outputs--value--properties))

<a id="nestedblock--columns--outputs--value--properties"></a>
### Nested Schema for `columns.outputs.value.properties`

Required:

- `schema_property_key` (String)

Optional:

- `properties` (Block List) (see [below for nested schema](#nestedblock--columns--outputs--value--properties--properties))

<a id="nestedblock--columns--outputs--value--properties--properties"></a>
### Nested Schema for `columns.outputs.value.properties.properties`

Required:

- `schema_property_key` (String)

Optional:

- `properties` (Block List) (see [below for nested schema](#nestedblock--columns--outputs--value--properties--properties--properties))

<a id="nestedblock--columns--outputs--value--properties--properties--properties"></a>
### Nested Schema for `columns.outputs.value.properties.properties.properties`

Required:

- `schema_property_key` (String)





<a id="nestedblock--columns--outputs--defaults_to"></a>
### Nested Schema for `columns.outputs.defaults_to`

Optional:

- `special` (String) Special enum value: Wildcard, Null, Empty, CurrentTime.
- `value` (String) Single string value. For queue columns, can be a UUID.




<a id="nestedblock--rows"></a>
### Nested Schema for `rows`

Optional:

- `inputs` (Block List) Input values (conditions) for this decision row. Each input specifies which column it belongs to using schema_property_key and optionally comparator. (see [below for nested schema](#nestedblock--rows--inputs))
- `outputs` (Block List) Output values (actions) for this decision row. Each output specifies which column it belongs to using schema_property_key and optionally comparator. (see [below for nested schema](#nestedblock--rows--outputs))

Read-Only:

- `row_id` (String) Unique identifier for this row within the decision table. Auto-generated by the system.
- `row_index` (Number) The position of this row in the decision table (1-based). Auto-generated by the system.

<a id="nestedblock--rows--inputs"></a>
### Nested Schema for `rows.inputs`

Required:

- `literal` (Block List, Min: 1, Max: 1) The literal value for this input parameter (see [below for nested schema](#nestedblock--rows--inputs--literal))
- `schema_property_key` (String) The schema property key that identifies which input column this value belongs to.

Optional:

- `comparator` (String) The comparator for this input column. Required when multiple columns have the same schema_property_key with different comparators. Optional when only one column exists for the schema_property_key.

Read-Only:

- `column_id` (String) The unique identifier of the input column. Auto-generated by the system.

<a id="nestedblock--rows--inputs--literal"></a>
### Nested Schema for `rows.inputs.literal`

Required:

- `type` (String) The type of the literal value.
- `value` (String) The literal value. IMPORTANT: All values must be wrapped in quotes, even numbers and booleans.

Examples:
- String: "VIP", "Hello World"
- Integer: "42", "0", "-10"
- Number: "3.14", "0.0", "-1.5"
- Boolean: "true", "false"
- Date: "2023-01-01"
- DateTime: "2023-01-01T12:00:00.000Z"
- Special: "Wildcard", "Null", "Empty", "CurrentTime"



<a id="nestedblock--rows--outputs"></a>
### Nested Schema for `rows.outputs`

Required:

- `literal` (Block List, Min: 1, Max: 1) The literal value for this output parameter. Only ONE field should be set per literal value (see [below for nested schema](#nestedblock--rows--outputs--literal))
- `schema_property_key` (String) The schema property key that identifies which output column this value belongs to.

Optional:

- `comparator` (String) The comparator for this output column. Required when multiple columns have the same schema_property_key with different comparators. Optional when only one column exists for the schema_property_key.

Read-Only:

- `column_id` (String) The unique identifier of the output column. Auto-generated by the system.

<a id="nestedblock--rows--outputs--literal"></a>
### Nested Schema for `rows.outputs.literal`

Required:

- `type` (String) The type of the literal value.
- `value` (String) The literal value. IMPORTANT: All values must be wrapped in quotes, even numbers and booleans.

Examples:
- String: "VIP", "Hello World"
- Integer: "42", "0", "-10"
- Number: "3.14", "0.0", "-1.5"
- Boolean: "true", "false"
- Date: "2023-01-01"
- DateTime: "2023-01-01T12:00:00.000Z"
- Special: "Wildcard", "Null", "Empty", "CurrentTime"

