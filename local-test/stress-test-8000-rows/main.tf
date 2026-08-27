resource "genesyscloud_business_rules_decision_table" "stress" {
  name        = "terraform-stress-table-${var.run_id}"
  description = "Stress test: 8000 rows, 30 columns (20 string + enum/int/num/bool/date/list)"
  division_id = data.genesyscloud_auth_division_home.home.id
  schema_id   = genesyscloud_business_rules_schema.stress.id

  timeouts {
    create = "360m"
    read   = "90m"
  }

  columns {
    inputs {
      defaults_to {
        special = "Wildcard"
      }
      expression {
        contractual {
          schema_property_key = "input_str_01"
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
          schema_property_key = "input_str_02"
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
          schema_property_key = "input_str_03"
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
          schema_property_key = "input_str_04"
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
          schema_property_key = "input_str_05"
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
          schema_property_key = "input_str_06"
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
          schema_property_key = "input_str_07"
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
          schema_property_key = "input_str_08"
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
          schema_property_key = "input_str_09"
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
          schema_property_key = "input_str_10"
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
          schema_property_key = "input_enum"
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
          schema_property_key = "input_int"
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
          schema_property_key = "input_num"
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
          schema_property_key = "input_bool"
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
          schema_property_key = "input_date"
        }
        comparator = "Equals"
      }
    }
    outputs {
      defaults_to {
        value = "default"
      }
      value {
        schema_property_key = "output_str_01"
      }
    }
    outputs {
      defaults_to {
        value = "default"
      }
      value {
        schema_property_key = "output_str_02"
      }
    }
    outputs {
      defaults_to {
        value = "default"
      }
      value {
        schema_property_key = "output_str_03"
      }
    }
    outputs {
      defaults_to {
        value = "default"
      }
      value {
        schema_property_key = "output_str_04"
      }
    }
    outputs {
      defaults_to {
        value = "default"
      }
      value {
        schema_property_key = "output_str_05"
      }
    }
    outputs {
      defaults_to {
        value = "default"
      }
      value {
        schema_property_key = "output_str_06"
      }
    }
    outputs {
      defaults_to {
        value = "default"
      }
      value {
        schema_property_key = "output_str_07"
      }
    }
    outputs {
      defaults_to {
        value = "default"
      }
      value {
        schema_property_key = "output_str_08"
      }
    }
    outputs {
      defaults_to {
        value = "default"
      }
      value {
        schema_property_key = "output_str_09"
      }
    }
    outputs {
      defaults_to {
        value = "default"
      }
      value {
        schema_property_key = "output_str_10"
      }
    }
    outputs {

      defaults_to {
        value = "opt_a"
      }
      value {
        schema_property_key = "output_enum"
      }
    }
    outputs {

      defaults_to {
        value = "1"
      }
      value {
        schema_property_key = "output_int"
      }
    }
    outputs {

      defaults_to {
        value = "1.0"
      }
      value {
        schema_property_key = "output_num"
      }
    }
    outputs {

      defaults_to {
        value = "false"
      }
      value {
        schema_property_key = "output_bool"
      }
    }
    outputs {

      defaults_to {
        values = ["out_a", "out_b"]
      }
      value {
        schema_property_key = "output_list"
      }
    }
  }

  dynamic "rows" {
    for_each = range(1000)
    content {
      inputs {
        literal {
          value = format("key-%05d", rows.value + 0)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c02", rows.value + 0)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c03", rows.value + 0)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c04", rows.value + 0)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c05", rows.value + 0)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c06", rows.value + 0)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c07", rows.value + 0)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c08", rows.value + 0)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c09", rows.value + 0)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c10", rows.value + 0)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 0) % 3)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("%d", (rows.value + 0) % 9999 + 1)
          type  = "integer"
        }
      }
      inputs {
        literal {
          value = format("%.1f", (rows.value + 0) % 100 + 0.5)
          type  = "number"
        }
      }
      inputs {
        literal {
          value = (rows.value + 0) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      inputs {
        literal {
          value = "2024-06-15"
          type  = "date"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c01", rows.value + 0)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c02", rows.value + 0)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c03", rows.value + 0)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c04", rows.value + 0)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c05", rows.value + 0)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c06", rows.value + 0)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c07", rows.value + 0)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c08", rows.value + 0)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c09", rows.value + 0)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c10", rows.value + 0)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 0) % 3)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("%d", (rows.value + 0) % 9999 + 1)
          type  = "integer"
        }
      }
      outputs {
        literal {
          value = format("%.1f", (rows.value + 0) % 100 + 0.5)
          type  = "number"
        }
      }
      outputs {
        literal {
          value = (rows.value + 0) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      outputs {
        literal {
          value = format("list-%05d-a,list-%05d-b", rows.value + 0, rows.value + 0)
          type  = "stringList"
        }
      }
    }
  }

  dynamic "rows" {
    for_each = range(1000)
    content {
      inputs {
        literal {
          value = format("key-%05d", rows.value + 1000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c02", rows.value + 1000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c03", rows.value + 1000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c04", rows.value + 1000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c05", rows.value + 1000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c06", rows.value + 1000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c07", rows.value + 1000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c08", rows.value + 1000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c09", rows.value + 1000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c10", rows.value + 1000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 1000) % 3)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("%d", (rows.value + 1000) % 9999 + 1)
          type  = "integer"
        }
      }
      inputs {
        literal {
          value = format("%.1f", (rows.value + 1000) % 100 + 0.5)
          type  = "number"
        }
      }
      inputs {
        literal {
          value = (rows.value + 1000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      inputs {
        literal {
          value = "2024-06-15"
          type  = "date"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c01", rows.value + 1000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c02", rows.value + 1000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c03", rows.value + 1000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c04", rows.value + 1000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c05", rows.value + 1000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c06", rows.value + 1000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c07", rows.value + 1000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c08", rows.value + 1000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c09", rows.value + 1000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c10", rows.value + 1000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 1000) % 3)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("%d", (rows.value + 1000) % 9999 + 1)
          type  = "integer"
        }
      }
      outputs {
        literal {
          value = format("%.1f", (rows.value + 1000) % 100 + 0.5)
          type  = "number"
        }
      }
      outputs {
        literal {
          value = (rows.value + 1000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      outputs {
        literal {
          value = format("list-%05d-a,list-%05d-b", rows.value + 1000, rows.value + 1000)
          type  = "stringList"
        }
      }
    }
  }

  dynamic "rows" {
    for_each = range(1000)
    content {
      inputs {
        literal {
          value = format("key-%05d", rows.value + 2000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c02", rows.value + 2000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c03", rows.value + 2000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c04", rows.value + 2000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c05", rows.value + 2000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c06", rows.value + 2000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c07", rows.value + 2000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c08", rows.value + 2000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c09", rows.value + 2000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c10", rows.value + 2000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 2000) % 3)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("%d", (rows.value + 2000) % 9999 + 1)
          type  = "integer"
        }
      }
      inputs {
        literal {
          value = format("%.1f", (rows.value + 2000) % 100 + 0.5)
          type  = "number"
        }
      }
      inputs {
        literal {
          value = (rows.value + 2000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      inputs {
        literal {
          value = "2024-06-15"
          type  = "date"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c01", rows.value + 2000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c02", rows.value + 2000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c03", rows.value + 2000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c04", rows.value + 2000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c05", rows.value + 2000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c06", rows.value + 2000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c07", rows.value + 2000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c08", rows.value + 2000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c09", rows.value + 2000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c10", rows.value + 2000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 2000) % 3)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("%d", (rows.value + 2000) % 9999 + 1)
          type  = "integer"
        }
      }
      outputs {
        literal {
          value = format("%.1f", (rows.value + 2000) % 100 + 0.5)
          type  = "number"
        }
      }
      outputs {
        literal {
          value = (rows.value + 2000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      outputs {
        literal {
          value = format("list-%05d-a,list-%05d-b", rows.value + 2000, rows.value + 2000)
          type  = "stringList"
        }
      }
    }
  }

  dynamic "rows" {
    for_each = range(1000)
    content {
      inputs {
        literal {
          value = format("key-%05d", rows.value + 3000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c02", rows.value + 3000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c03", rows.value + 3000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c04", rows.value + 3000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c05", rows.value + 3000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c06", rows.value + 3000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c07", rows.value + 3000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c08", rows.value + 3000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c09", rows.value + 3000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c10", rows.value + 3000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 3000) % 3)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("%d", (rows.value + 3000) % 9999 + 1)
          type  = "integer"
        }
      }
      inputs {
        literal {
          value = format("%.1f", (rows.value + 3000) % 100 + 0.5)
          type  = "number"
        }
      }
      inputs {
        literal {
          value = (rows.value + 3000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      inputs {
        literal {
          value = "2024-06-15"
          type  = "date"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c01", rows.value + 3000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c02", rows.value + 3000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c03", rows.value + 3000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c04", rows.value + 3000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c05", rows.value + 3000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c06", rows.value + 3000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c07", rows.value + 3000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c08", rows.value + 3000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c09", rows.value + 3000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c10", rows.value + 3000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 3000) % 3)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("%d", (rows.value + 3000) % 9999 + 1)
          type  = "integer"
        }
      }
      outputs {
        literal {
          value = format("%.1f", (rows.value + 3000) % 100 + 0.5)
          type  = "number"
        }
      }
      outputs {
        literal {
          value = (rows.value + 3000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      outputs {
        literal {
          value = format("list-%05d-a,list-%05d-b", rows.value + 3000, rows.value + 3000)
          type  = "stringList"
        }
      }
    }
  }

  dynamic "rows" {
    for_each = range(1000)
    content {
      inputs {
        literal {
          value = format("key-%05d", rows.value + 4000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c02", rows.value + 4000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c03", rows.value + 4000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c04", rows.value + 4000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c05", rows.value + 4000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c06", rows.value + 4000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c07", rows.value + 4000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c08", rows.value + 4000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c09", rows.value + 4000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c10", rows.value + 4000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 4000) % 3)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("%d", (rows.value + 4000) % 9999 + 1)
          type  = "integer"
        }
      }
      inputs {
        literal {
          value = format("%.1f", (rows.value + 4000) % 100 + 0.5)
          type  = "number"
        }
      }
      inputs {
        literal {
          value = (rows.value + 4000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      inputs {
        literal {
          value = "2024-06-15"
          type  = "date"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c01", rows.value + 4000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c02", rows.value + 4000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c03", rows.value + 4000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c04", rows.value + 4000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c05", rows.value + 4000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c06", rows.value + 4000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c07", rows.value + 4000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c08", rows.value + 4000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c09", rows.value + 4000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c10", rows.value + 4000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 4000) % 3)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("%d", (rows.value + 4000) % 9999 + 1)
          type  = "integer"
        }
      }
      outputs {
        literal {
          value = format("%.1f", (rows.value + 4000) % 100 + 0.5)
          type  = "number"
        }
      }
      outputs {
        literal {
          value = (rows.value + 4000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      outputs {
        literal {
          value = format("list-%05d-a,list-%05d-b", rows.value + 4000, rows.value + 4000)
          type  = "stringList"
        }
      }
    }
  }

  dynamic "rows" {
    for_each = range(1000)
    content {
      inputs {
        literal {
          value = format("key-%05d", rows.value + 5000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c02", rows.value + 5000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c03", rows.value + 5000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c04", rows.value + 5000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c05", rows.value + 5000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c06", rows.value + 5000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c07", rows.value + 5000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c08", rows.value + 5000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c09", rows.value + 5000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c10", rows.value + 5000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 5000) % 3)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("%d", (rows.value + 5000) % 9999 + 1)
          type  = "integer"
        }
      }
      inputs {
        literal {
          value = format("%.1f", (rows.value + 5000) % 100 + 0.5)
          type  = "number"
        }
      }
      inputs {
        literal {
          value = (rows.value + 5000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      inputs {
        literal {
          value = "2024-06-15"
          type  = "date"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c01", rows.value + 5000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c02", rows.value + 5000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c03", rows.value + 5000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c04", rows.value + 5000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c05", rows.value + 5000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c06", rows.value + 5000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c07", rows.value + 5000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c08", rows.value + 5000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c09", rows.value + 5000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c10", rows.value + 5000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 5000) % 3)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("%d", (rows.value + 5000) % 9999 + 1)
          type  = "integer"
        }
      }
      outputs {
        literal {
          value = format("%.1f", (rows.value + 5000) % 100 + 0.5)
          type  = "number"
        }
      }
      outputs {
        literal {
          value = (rows.value + 5000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      outputs {
        literal {
          value = format("list-%05d-a,list-%05d-b", rows.value + 5000, rows.value + 5000)
          type  = "stringList"
        }
      }
    }
  }

  dynamic "rows" {
    for_each = range(1000)
    content {
      inputs {
        literal {
          value = format("key-%05d", rows.value + 6000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c02", rows.value + 6000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c03", rows.value + 6000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c04", rows.value + 6000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c05", rows.value + 6000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c06", rows.value + 6000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c07", rows.value + 6000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c08", rows.value + 6000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c09", rows.value + 6000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c10", rows.value + 6000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 6000) % 3)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("%d", (rows.value + 6000) % 9999 + 1)
          type  = "integer"
        }
      }
      inputs {
        literal {
          value = format("%.1f", (rows.value + 6000) % 100 + 0.5)
          type  = "number"
        }
      }
      inputs {
        literal {
          value = (rows.value + 6000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      inputs {
        literal {
          value = "2024-06-15"
          type  = "date"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c01", rows.value + 6000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c02", rows.value + 6000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c03", rows.value + 6000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c04", rows.value + 6000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c05", rows.value + 6000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c06", rows.value + 6000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c07", rows.value + 6000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c08", rows.value + 6000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c09", rows.value + 6000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c10", rows.value + 6000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 6000) % 3)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("%d", (rows.value + 6000) % 9999 + 1)
          type  = "integer"
        }
      }
      outputs {
        literal {
          value = format("%.1f", (rows.value + 6000) % 100 + 0.5)
          type  = "number"
        }
      }
      outputs {
        literal {
          value = (rows.value + 6000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      outputs {
        literal {
          value = format("list-%05d-a,list-%05d-b", rows.value + 6000, rows.value + 6000)
          type  = "stringList"
        }
      }
    }
  }

  dynamic "rows" {
    for_each = range(1000)
    content {
      inputs {
        literal {
          value = format("key-%05d", rows.value + 7000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c02", rows.value + 7000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c03", rows.value + 7000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c04", rows.value + 7000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c05", rows.value + 7000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c06", rows.value + 7000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c07", rows.value + 7000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c08", rows.value + 7000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c09", rows.value + 7000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("in-%05d-c10", rows.value + 7000)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 7000) % 3)
          type  = "string"
        }
      }
      inputs {
        literal {
          value = format("%d", (rows.value + 7000) % 9999 + 1)
          type  = "integer"
        }
      }
      inputs {
        literal {
          value = format("%.1f", (rows.value + 7000) % 100 + 0.5)
          type  = "number"
        }
      }
      inputs {
        literal {
          value = (rows.value + 7000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      inputs {
        literal {
          value = "2024-06-15"
          type  = "date"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c01", rows.value + 7000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c02", rows.value + 7000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c03", rows.value + 7000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c04", rows.value + 7000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c05", rows.value + 7000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c06", rows.value + 7000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c07", rows.value + 7000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c08", rows.value + 7000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c09", rows.value + 7000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("out-%05d-c10", rows.value + 7000)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = element(["opt_a", "opt_b", "opt_c"], (rows.value + 7000) % 3)
          type  = "string"
        }
      }
      outputs {
        literal {
          value = format("%d", (rows.value + 7000) % 9999 + 1)
          type  = "integer"
        }
      }
      outputs {
        literal {
          value = format("%.1f", (rows.value + 7000) % 100 + 0.5)
          type  = "number"
        }
      }
      outputs {
        literal {
          value = (rows.value + 7000) % 2 == 0 ? "true" : "false"
          type  = "boolean"
        }
      }
      outputs {
        literal {
          value = format("list-%05d-a,list-%05d-b", rows.value + 7000, rows.value + 7000)
          type  = "stringList"
        }
      }
    }
  }
}
