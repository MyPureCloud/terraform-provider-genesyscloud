---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "genesyscloud_task_management_worktype_flow_datebased_rule Data Source - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud task management datebased rule data source. Select a task management datebased rule by name
---

# genesyscloud_task_management_worktype_flow_datebased_rule (Data Source)

Genesys Cloud task management datebased rule data source. Select a task management datebased rule by name

## Example Usage

```terraform
data "genesyscloud_task_management_worktype_flow_datebased_rule" "datebased_rule_data" {
  worktype_id = genesyscloud_task_management_worktype.example.id
  name        = "DateBased Rule"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the Rule.
- `worktype_id` (String) The Worktype ID of the Rule.

### Read-Only

- `id` (String) The ID of this resource.
