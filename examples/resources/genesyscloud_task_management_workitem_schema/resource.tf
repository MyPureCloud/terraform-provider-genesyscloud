resource "genesyscloud_task_management_workitem_schema" "example_schema" {
  enabled     = "true"
  name        = "Example Schema"
  description = "The workitem schema description"
  properties = jsonencode({
    "custom_attribute_1_text" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/text"
        }
      ],
      "title" : "custom_attribute_1",
      "description" : "Custom attribute for text",
      "minLength" : 0,
      "maxLength" : 100
    },
    "custom_attribute_2_longtext" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/longtext"
        }
      ],
      "title" : "custom_attribute_2",
      "description" : "Custom attribute for long text",
      "minLength" : 0,
      "maxLength" : 1000
    },
    "custom_attribute_3_url" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/url"
        }
      ],
      "title" : "custom_attribute_3",
      "description" : "Custom attribute for url",
      "minLength" : 0,
      "maxLength" : 200
    },
    "custom_attribute_4_identifier" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/identifier"
        }
      ],
      "title" : "custom_attribute_4",
      "description" : "Custom attribute for identifier",
      "minLength" : 0,
      "maxLength" : 100
    },
    "custom_attribute_5_enum" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/enum"
        }
      ],
      "title" : "custom_attribute_5",
      "description" : "Custom attribute for enum",
      "enum" : ["option_1", "option_2", "option_3"],
      "_enumProperties" : {
        "option_1" : {
          "title" : "Option 1",
          "_disabled" : false
        },
        "option_2" : {
          "title" : "Option 2",
          "_disabled" : false
        },
        "option_3" : {
          "title" : "Option 3",
          "_disabled" : false
        },
      },
    },
    "custom_attribute_6_date" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/date"
        }
      ],
      "title" : "custom_attribute_6",
      "description" : "Custom attribute for date",
    },
    "custom_attribute_7_datetime" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/datetime"
        }
      ],
      "title" : "custom_attribute_7",
      "description" : "Custom attribute for datetime",
    },
    "custom_attribute_8_integer" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/integer"
        }
      ],
      "title" : "custom_attribute_8",
      "description" : "Custom attribute for integer",
      "minimum" : 1,
      "maximum" : 1000
    },
    "custom_attribute_9_number" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/number"
        }
      ],
      "title" : "custom_attribute_9",
      "description" : "Custom attribute for number",
      "minimum" : 1,
      "maximum" : 1000
    },
    "custom_attribute_10_checkbox" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/checkbox"
        }
      ],
      "title" : "custom_attribute_10",
      "description" : "Custom attribute for checkbox"
    },
    "custom_attribute_11_tag" : {
      "allOf" : [
        {
          "$ref" : "#/definitions/tag"
        }
      ],
      "title" : "custom_attribute_11",
      "description" : "Custom attribute for tag",
      "items" : {
        "minLength" : 1,
        "maxLength" : 100
      },
      "minItems" : 0,
      "maxItems" : 10,
      "uniqueItems" : true
    },
  })
}
