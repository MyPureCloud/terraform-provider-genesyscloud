#!/usr/bin/env python3
"""Generate schema.tf and main.tf for the 8000-row / 30-column stress test."""

from pathlib import Path

OUT_DIR = Path(__file__).parent

# Genesys Cloud org limits (decision tables):
#   decision.table.rows.max                              = 8000
#   decision.tables.attributes.max                       = 30
#   decision.tables.string.type.attributes.max           = 20
#   decision.tables.queue.columns.max                    = 5
#   decision.tables.string.list.type.attributes.max      = 50
#   decision.table.rows.max.string.attributes.length.combined = 4000
#
# This stress test targets the first three limits exactly.

ROW_COUNT = 8000
CHUNK_SIZE = 1000  # Terraform range() is capped at 1024 values

INPUT_STRING_COUNT = 10
OUTPUT_STRING_COUNT = 10
# 10 + 10 = 20 string columns (decision.tables.string.type.attributes.max)

INPUT_OTHER = [
    ("input_enum", "enum", "Equals"),
    ("input_int", "integer", "Equals"),
    ("input_num", "number", "Equals"),
    ("input_bool", "boolean", "Equals"),
    ("input_date", "date", "Equals"),
]

OUTPUT_OTHER = [
    ("output_enum", "enum"),
    ("output_int", "integer"),
    ("output_num", "number"),
    ("output_bool", "boolean"),
    ("output_list", "stringList"),
]


def schema_string_property(key: str, side: str, index: int) -> list[str]:
    return [
        f"    {key} = {{",
        '      allOf = [{ "$ref" = "#/definitions/string" }]',
        f'      title = "{key}"',
        f'      description = "{side} string column {index:02d}"',
        "      minLength = 1",
        "      maxLength = 100",
        "    }",
    ]


def schema_enum_property(key: str, side: str) -> list[str]:
    return [
        f"    {key} = {{",
        '      allOf = [{ "$ref" = "#/definitions/enum" }]',
        f'      title = "{key}"',
        f'      description = "{side} enum column"',
        '      enum = ["opt_a", "opt_b", "opt_c"]',
        "      _enumProperties = {",
        '        opt_a = { title = "Option A" }',
        '        opt_b = { title = "Option B" }',
        '        opt_c = { title = "Option C" }',
        "      }",
        "    }",
    ]


def schema_integer_property(key: str, side: str) -> list[str]:
    return [
        f"    {key} = {{",
        '      allOf = [{ "$ref" = "#/definitions/integer" }]',
        f'      title = "{key}"',
        f'      description = "{side} integer column"',
        "      minimum = 1",
        "      maximum = 9999",
        "    }",
    ]


def schema_number_property(key: str, side: str) -> list[str]:
    return [
        f"    {key} = {{",
        '      allOf = [{ "$ref" = "#/definitions/number" }]',
        f'      title = "{key}"',
        f'      description = "{side} number column"',
        "      minimum = 0",
        "      maximum = 100",
        "    }",
    ]


def schema_boolean_property(key: str, side: str) -> list[str]:
    return [
        f"    {key} = {{",
        '      allOf = [{ "$ref" = "#/definitions/boolean" }]',
        f'      title = "{key}"',
        f'      description = "{side} boolean column"',
        "    }",
    ]


def schema_date_property(key: str, side: str) -> list[str]:
    return [
        f"    {key} = {{",
        '      allOf = [{ "$ref" = "#/definitions/date" }]',
        f'      title = "{key}"',
        f'      description = "{side} date column"',
        "    }",
    ]


def schema_string_list_property(key: str, side: str) -> list[str]:
    return [
        f"    {key} = {{",
        '      allOf = [{ "$ref" = "#/definitions/stringList" }]',
        f'      title = "{key}"',
        f'      description = "{side} string list column"',
        "      minItems = 1",
        "      maxItems = 5",
        "      uniqueItems = true",
        "      items = {",
        "        minLength = 3",
        "        maxLength = 20",
        "      }",
        "    }",
    ]


SCHEMA_BUILDERS = {
    "string": lambda key, side, index=0: schema_string_property(key, side, index),
    "enum": lambda key, side, index=0: schema_enum_property(key, side),
    "integer": lambda key, side, index=0: schema_integer_property(key, side),
    "number": lambda key, side, index=0: schema_number_property(key, side),
    "boolean": lambda key, side, index=0: schema_boolean_property(key, side),
    "date": lambda key, side, index=0: schema_date_property(key, side),
    "stringList": lambda key, side, index=0: schema_string_list_property(key, side),
}


def schema_property_lines() -> list[str]:
    lines: list[str] = []
    for i in range(1, INPUT_STRING_COUNT + 1):
        lines.extend(schema_string_property(f"input_str_{i:02d}", "Input", i))
    for key, col_type, _ in INPUT_OTHER:
        lines.extend(SCHEMA_BUILDERS[col_type](key, "Input"))
    for i in range(1, OUTPUT_STRING_COUNT + 1):
        lines.extend(schema_string_property(f"output_str_{i:02d}", "Output", i))
    for key, col_type in OUTPUT_OTHER:
        lines.extend(SCHEMA_BUILDERS[col_type](key, "Output"))
    return lines


def input_string_column(key: str) -> str:
    return f"""    inputs {{
      defaults_to {{
        special = "Wildcard"
      }}
      expression {{
        contractual {{
          schema_property_key = "{key}"
        }}
        comparator = "Equals"
      }}
    }}"""


def input_other_column(key: str, col_type: str, comparator: str) -> str:
    defaults = """
      defaults_to {
        special = "Wildcard"
      }"""
    return f"""    inputs {{
{defaults}
      expression {{
        contractual {{
          schema_property_key = "{key}"
        }}
        comparator = "{comparator}"
      }}
    }}"""


def output_string_column(key: str) -> str:
    return f"""    outputs {{
      defaults_to {{
        value = "default"
      }}
      value {{
        schema_property_key = "{key}"
      }}
    }}"""


def output_other_column(key: str, col_type: str) -> str:
    if col_type == "stringList":
        defaults = """
      defaults_to {
        values = ["out_a", "out_b"]
      }"""
    elif col_type == "enum":
        defaults = """
      defaults_to {
        value = "opt_a"
      }"""
    elif col_type == "integer":
        defaults = """
      defaults_to {
        value = "1"
      }"""
    elif col_type == "number":
        defaults = """
      defaults_to {
        value = "1.0"
      }"""
    elif col_type == "boolean":
        defaults = """
      defaults_to {
        value = "false"
      }"""
    else:
        defaults = """
      defaults_to {
        special = "Null"
      }"""
    return f"""    outputs {{
{defaults}
      value {{
        schema_property_key = "{key}"
      }}
    }}"""


def row_chunks(total: int, chunk_size: int) -> list[tuple[int, int]]:
    chunks: list[tuple[int, int]] = []
    start = 0
    while start < total:
        count = min(chunk_size, total - start)
        if count > 1024:
            raise ValueError(f"chunk size {count} exceeds Terraform range() limit of 1024")
        chunks.append((start, count))
        start += count
    return chunks


def row_input_literal(col_type: str, idx_expr: str, col_index: int) -> str:
    if col_type == "string":
        if col_index == 1:
            value = f'format("key-%05d", {idx_expr})'
        else:
            value = f'format("in-%05d-c{col_index:02d}", {idx_expr})'
        literal_type = "string"
    elif col_type == "enum":
        value = f'element(["opt_a", "opt_b", "opt_c"], ({idx_expr}) % 3)'
        literal_type = "string"
    elif col_type == "integer":
        value = f'format("%d", ({idx_expr}) % 9999 + 1)'
        literal_type = "integer"
    elif col_type == "number":
        value = f'format("%.1f", ({idx_expr}) % 100 + 0.5)'
        literal_type = "number"
    elif col_type == "boolean":
        value = f'({idx_expr}) % 2 == 0 ? "true" : "false"'
        literal_type = "boolean"
    elif col_type == "date":
        value = '"2024-06-15"'
        literal_type = "date"
    else:
        raise ValueError(f"unsupported input type: {col_type}")
    return f"""      inputs {{
        literal {{
          value = {value}
          type  = "{literal_type}"
        }}
      }}"""


def row_output_literal(col_type: str, idx_expr: str, col_index: int) -> str:
    if col_type == "string":
        value = f'format("out-%05d-c{col_index:02d}", {idx_expr})'
        literal_type = "string"
    elif col_type == "enum":
        value = f'element(["opt_a", "opt_b", "opt_c"], ({idx_expr}) % 3)'
        literal_type = "string"
    elif col_type == "integer":
        value = f'format("%d", ({idx_expr}) % 9999 + 1)'
        literal_type = "integer"
    elif col_type == "number":
        value = f'format("%.1f", ({idx_expr}) % 100 + 0.5)'
        literal_type = "number"
    elif col_type == "boolean":
        value = f'({idx_expr}) % 2 == 0 ? "true" : "false"'
        literal_type = "boolean"
    elif col_type == "stringList":
        value = f'format("list-%05d-a,list-%05d-b", {idx_expr}, {idx_expr})'
        literal_type = "stringList"
    else:
        raise ValueError(f"unsupported output type: {col_type}")
    return f"""      outputs {{
        literal {{
          value = {value}
          type  = "{literal_type}"
        }}
      }}"""


def row_literal_blocks(row_offset: int) -> list[str]:
    idx = f"rows.value + {row_offset}"
    lines: list[str] = []
    for i in range(1, INPUT_STRING_COUNT + 1):
        lines.append(row_input_literal("string", idx, i))
    for _, col_type, _ in INPUT_OTHER:
        lines.append(row_input_literal(col_type, idx, 0))
    for i in range(1, OUTPUT_STRING_COUNT + 1):
        lines.append(row_output_literal("string", idx, i))
    for _, col_type in OUTPUT_OTHER:
        lines.append(row_output_literal(col_type, idx, 0))
    return lines


def generate_row_dynamics(total: int) -> str:
    blocks: list[str] = []
    for row_offset, count in row_chunks(total, CHUNK_SIZE):
        row_content = "\n".join(row_literal_blocks(row_offset))
        blocks.append(f"""  dynamic "rows" {{
    for_each = range({count})
    content {{
{row_content}
    }}
  }}""")
    return "\n\n".join(blocks)


def generate_schema_tf() -> str:
    props = "\n".join(schema_property_lines())
    return f"""resource "genesyscloud_business_rules_schema" "stress" {{
  enabled     = "true"
  name        = "terraform-stress-schema-${{var.run_id}}"
  description = "Stress test schema: 30 columns (20 string + 10 mixed types)"

  properties = jsonencode({{
{props}
  }})
}}
"""


def generate_main_tf() -> str:
    column_blocks: list[str] = []
    for i in range(1, INPUT_STRING_COUNT + 1):
        column_blocks.append(input_string_column(f"input_str_{i:02d}"))
    for key, col_type, comparator in INPUT_OTHER:
        column_blocks.append(input_other_column(key, col_type, comparator))
    for i in range(1, OUTPUT_STRING_COUNT + 1):
        column_blocks.append(output_string_column(f"output_str_{i:02d}"))
    for key, col_type in OUTPUT_OTHER:
        column_blocks.append(output_other_column(key, col_type))

    columns = "\n".join(column_blocks)
    row_dynamics = generate_row_dynamics(ROW_COUNT)
    total_columns = INPUT_STRING_COUNT + len(INPUT_OTHER) + OUTPUT_STRING_COUNT + len(OUTPUT_OTHER)

    return f"""resource "genesyscloud_business_rules_decision_table" "stress" {{
  name        = "terraform-stress-table-${{var.run_id}}"
  description = "Stress test: {ROW_COUNT} rows, {total_columns} columns (20 string + enum/int/num/bool/date/list)"
  division_id = data.genesyscloud_auth_division_home.home.id
  schema_id   = genesyscloud_business_rules_schema.stress.id

  timeouts {{
    create = "360m"
    read   = "90m"
  }}

  columns {{
{columns}
  }}

{row_dynamics}
}}
"""


def main() -> None:
    (OUT_DIR / "schema.tf").write_text(generate_schema_tf())
    (OUT_DIR / "main.tf").write_text(generate_main_tf())
    print(f"Generated schema.tf and main.tf in {OUT_DIR}")


if __name__ == "__main__":
    main()
