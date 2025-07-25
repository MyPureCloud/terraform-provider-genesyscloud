resource "genesyscloud_business_rules_schema" "example_business_rules_schema" {
  enabled     = "true"
  name        = "Example Schema"
  description = "The business rules schema description"
  properties = jsonencode({
    "custom_attribute_boolean" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/boolean"
        }
      ],
      "title" : "custom_attribute_boolean",
      "description" : "Custom attribute for boolean"
    },
    "custom_attribute_date" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/date"
        }
      ],
      "title" : "custom_attribute_date",
      "description" : "Custom attribute for date, format: YYYY-MM-DD"
    },
    "custom_attribute_datetime" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/datetime"
        }
      ],
      "title" : "custom_attribute_datetime",
      "description" : "Custom attribute for date time, format: YYYY-MM-DDTHH:mm:ss.sssZ"
    },
    "custom_attribute_enum" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/enum"
        }
      ],
      "title" : "custom_attribute_enum",
      "description" : "Custom attribute for enum",
      "enum" : ["option_1", "option_2", "option_3"],
      "_enumProperties" : {
        "option_1" : {
          "title" : "Option 1"
        },
        "option_2" : {
          "title" : "Option 2"
        },
        "option_3" : {
          "title" : "Option 3"
        },
      },
    },
    "custom_attribute_integer" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/integer"
        }
      ],
      "title" : "custom_attribute_integer",
      "description" : "Custom attribute for integer",
      "minimum" : 1,
      "maximum" : 1000
    },
    "custom_attribute_number" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/number"
        }
      ],
      "title" : "custom_attribute_number",
      "description" : "Custom attribute for number",
      "minimum" : 1,
      "maximum" : 1000
    },
    "custom_attribute_queue" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/businessRulesQueue"
        }
      ],
      "title" : "custom_attribute_queue",
      "description" : "Custom attribute for queue",
    },
    "custom_attribute_string" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/string"
        }
      ],
      "title" : "custom_attribute_string",
      "description" : "Custom attribute for string",
      "minLength" : 1,
      "maxLength" : 100
    }
  })
}
