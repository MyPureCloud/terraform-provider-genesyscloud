resource "genesyscloud_business_rules_schema" "stress" {
  enabled     = "true"
  name        = "terraform-stress-schema-${var.run_id}"
  description = "Stress test schema: 30 columns (20 string + 10 mixed types)"

  properties = jsonencode({
    input_str_01 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "input_str_01"
      description = "Input string column 01"
      minLength   = 1
      maxLength   = 100
    }
    input_str_02 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "input_str_02"
      description = "Input string column 02"
      minLength   = 1
      maxLength   = 100
    }
    input_str_03 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "input_str_03"
      description = "Input string column 03"
      minLength   = 1
      maxLength   = 100
    }
    input_str_04 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "input_str_04"
      description = "Input string column 04"
      minLength   = 1
      maxLength   = 100
    }
    input_str_05 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "input_str_05"
      description = "Input string column 05"
      minLength   = 1
      maxLength   = 100
    }
    input_str_06 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "input_str_06"
      description = "Input string column 06"
      minLength   = 1
      maxLength   = 100
    }
    input_str_07 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "input_str_07"
      description = "Input string column 07"
      minLength   = 1
      maxLength   = 100
    }
    input_str_08 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "input_str_08"
      description = "Input string column 08"
      minLength   = 1
      maxLength   = 100
    }
    input_str_09 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "input_str_09"
      description = "Input string column 09"
      minLength   = 1
      maxLength   = 100
    }
    input_str_10 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "input_str_10"
      description = "Input string column 10"
      minLength   = 1
      maxLength   = 100
    }
    input_enum = {
      allOf       = [{ "$ref" = "#/definitions/enum" }]
      title       = "input_enum"
      description = "Input enum column"
      enum        = ["opt_a", "opt_b", "opt_c"]
      _enumProperties = {
        opt_a = { title = "Option A" }
        opt_b = { title = "Option B" }
        opt_c = { title = "Option C" }
      }
    }
    input_int = {
      allOf       = [{ "$ref" = "#/definitions/integer" }]
      title       = "input_int"
      description = "Input integer column"
      minimum     = 1
      maximum     = 9999
    }
    input_num = {
      allOf       = [{ "$ref" = "#/definitions/number" }]
      title       = "input_num"
      description = "Input number column"
      minimum     = 0
      maximum     = 100
    }
    input_bool = {
      allOf       = [{ "$ref" = "#/definitions/boolean" }]
      title       = "input_bool"
      description = "Input boolean column"
    }
    input_date = {
      allOf       = [{ "$ref" = "#/definitions/date" }]
      title       = "input_date"
      description = "Input date column"
    }
    output_str_01 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "output_str_01"
      description = "Output string column 01"
      minLength   = 1
      maxLength   = 100
    }
    output_str_02 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "output_str_02"
      description = "Output string column 02"
      minLength   = 1
      maxLength   = 100
    }
    output_str_03 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "output_str_03"
      description = "Output string column 03"
      minLength   = 1
      maxLength   = 100
    }
    output_str_04 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "output_str_04"
      description = "Output string column 04"
      minLength   = 1
      maxLength   = 100
    }
    output_str_05 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "output_str_05"
      description = "Output string column 05"
      minLength   = 1
      maxLength   = 100
    }
    output_str_06 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "output_str_06"
      description = "Output string column 06"
      minLength   = 1
      maxLength   = 100
    }
    output_str_07 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "output_str_07"
      description = "Output string column 07"
      minLength   = 1
      maxLength   = 100
    }
    output_str_08 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "output_str_08"
      description = "Output string column 08"
      minLength   = 1
      maxLength   = 100
    }
    output_str_09 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "output_str_09"
      description = "Output string column 09"
      minLength   = 1
      maxLength   = 100
    }
    output_str_10 = {
      allOf       = [{ "$ref" = "#/definitions/string" }]
      title       = "output_str_10"
      description = "Output string column 10"
      minLength   = 1
      maxLength   = 100
    }
    output_enum = {
      allOf       = [{ "$ref" = "#/definitions/enum" }]
      title       = "output_enum"
      description = "Output enum column"
      enum        = ["opt_a", "opt_b", "opt_c"]
      _enumProperties = {
        opt_a = { title = "Option A" }
        opt_b = { title = "Option B" }
        opt_c = { title = "Option C" }
      }
    }
    output_int = {
      allOf       = [{ "$ref" = "#/definitions/integer" }]
      title       = "output_int"
      description = "Output integer column"
      minimum     = 1
      maximum     = 9999
    }
    output_num = {
      allOf       = [{ "$ref" = "#/definitions/number" }]
      title       = "output_num"
      description = "Output number column"
      minimum     = 0
      maximum     = 100
    }
    output_bool = {
      allOf       = [{ "$ref" = "#/definitions/boolean" }]
      title       = "output_bool"
      description = "Output boolean column"
    }
    output_list = {
      allOf       = [{ "$ref" = "#/definitions/stringList" }]
      title       = "output_list"
      description = "Output string list column"
      minItems    = 1
      maxItems    = 5
      uniqueItems = true
      items = {
        minLength = 3
        maxLength = 20
      }
    }
  })
}
