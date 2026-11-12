---
page_title: "genesyscloud_case_management_caseplan Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud case management caseplan, including its stageplans and stepplans.
  Terraform manages the published caseplan. Every apply that changes a versioned field (anything other than name and description) creates a draft (or reuses an existing one), writes the configuration over it, and publishes it. Changes made to a draft outside Terraform are overwritten by the next such apply. name and description are not versioned and are updated in place without publishing.
  division_id, customer_intent, reference_prefix and data_schema cannot change once the caseplan is published, which happens on the first apply.
  Adding, removing or reordering stageplan blocks requires the add/delete stageplans feature in the org.
---
# genesyscloud_case_management_caseplan (Resource)

<!-- This document is automatically generated. Do not edit manually. Make changes to the schema, examples, or apis.md files in examples/resources/ and run 'make docs' to regenerate. -->

Genesys Cloud case management caseplan, including its stageplans and stepplans.

Terraform manages the **published** caseplan. Every apply that changes a versioned field (anything other than `name` and `description`) creates a draft (or reuses an existing one), writes the configuration over it, and publishes it. Changes made to a draft outside Terraform are overwritten by the next such apply. `name` and `description` are not versioned and are updated in place without publishing.

`division_id`, `customer_intent`, `reference_prefix` and `data_schema` cannot change once the caseplan is published, which happens on the first apply.

Adding, removing or reordering `stageplan` blocks requires the add/delete stageplans feature in the org.

## API Usage

The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/casemanagement/caseplans](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-casemanagement-caseplans)
* [POST /api/v2/casemanagement/caseplans](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-casemanagement-caseplans)
* [DELETE /api/v2/casemanagement/caseplans/{caseplanId}](https://developer.genesys.cloud/devapps/api-explorer#delete-api-v2-casemanagement-caseplans--caseplanId-)
* [GET /api/v2/casemanagement/caseplans/{caseplanId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-casemanagement-caseplans--caseplanId-)
* [PATCH /api/v2/casemanagement/caseplans/{caseplanId}](https://developer.genesys.cloud/devapps/api-explorer#patch-api-v2-casemanagement-caseplans--caseplanId-)
* [POST /api/v2/casemanagement/caseplans/{caseplanId}/dataschemas](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-casemanagement-caseplans--caseplanId--dataschemas)
* [PUT /api/v2/casemanagement/caseplans/{caseplanId}/dataschemas/{schemaKeyName}](https://developer.genesys.cloud/devapps/api-explorer#put-api-v2-casemanagement-caseplans--caseplanId--dataschemas--schemaKeyName-)
* [PUT /api/v2/casemanagement/caseplans/{caseplanId}/intakesettings](https://developer.genesys.cloud/devapps/api-explorer#put-api-v2-casemanagement-caseplans--caseplanId--intakesettings)
* [POST /api/v2/casemanagement/caseplans/{caseplanId}/publish](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-casemanagement-caseplans--caseplanId--publish)
* [POST /api/v2/casemanagement/caseplans/{caseplanId}/stageplans](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-casemanagement-caseplans--caseplanId--stageplans)
* [DELETE /api/v2/casemanagement/caseplans/{caseplanId}/stageplans/{stageplanId}](https://developer.genesys.cloud/devapps/api-explorer#delete-api-v2-casemanagement-caseplans--caseplanId--stageplans--stageplanId-)
* [PATCH /api/v2/casemanagement/caseplans/{caseplanId}/stageplans/{stageplanId}](https://developer.genesys.cloud/devapps/api-explorer#patch-api-v2-casemanagement-caseplans--caseplanId--stageplans--stageplanId-)
* [POST /api/v2/casemanagement/caseplans/{caseplanId}/stageplans/{stageplanId}/reposition](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-casemanagement-caseplans--caseplanId--stageplans--stageplanId--reposition)
* [PATCH /api/v2/casemanagement/caseplans/{caseplanId}/stageplans/{stageplanId}/stepplans/{stepplanId}](https://developer.genesys.cloud/devapps/api-explorer#patch-api-v2-casemanagement-caseplans--caseplanId--stageplans--stageplanId--stepplans--stepplanId-)
* [POST /api/v2/casemanagement/caseplans/{caseplanId}/versions](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-casemanagement-caseplans--caseplanId--versions)
* [GET /api/v2/casemanagement/caseplans/{caseplanId}/versions/{versionId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-casemanagement-caseplans--caseplanId--versions--versionId-)
* [GET /api/v2/casemanagement/caseplans/{caseplanId}/versions/{versionId}/dataschemas](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-casemanagement-caseplans--caseplanId--versions--versionId--dataschemas)
* [GET /api/v2/casemanagement/caseplans/{caseplanId}/versions/{versionId}/intakesettings](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-casemanagement-caseplans--caseplanId--versions--versionId--intakesettings)
* [GET /api/v2/casemanagement/caseplans/{caseplanId}/versions/{versionId}/stageplans](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-casemanagement-caseplans--caseplanId--versions--versionId--stageplans)
* [GET /api/v2/casemanagement/caseplans/{caseplanId}/versions/{versionId}/stageplans/{stageplanId}/stepplans](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-casemanagement-caseplans--caseplanId--versions--versionId--stageplans--stageplanId--stepplans)

## Permissions and Scopes

The following permissions are required to use this resource:

* `caseManagement:caseplan:add`
* `caseManagement:caseplan:delete`
* `caseManagement:caseplan:edit`
* `caseManagement:caseplan:publish`
* `caseManagement:caseplan:version`
* `caseManagement:caseplan:view`
* `caseManagement:caseplanDataSchemas:add`
* `caseManagement:caseplanDataSchemas:edit`
* `caseManagement:caseplanDataSchemas:view`
* `caseManagement:caseplanIntakeSettings:edit`
* `caseManagement:caseplanIntakeSettings:view`
* `caseManagement:stageplan:edit`
* `caseManagement:stageplan:view`
* `caseManagement:stepplan:edit`
* `caseManagement:stepplan:view`

The following OAuth scopes are required to use this resource:

* `case-management`
* `case-management:readonly`


## Example Usage

```terraform
# Terraform manages the published caseplan. Every apply that changes a versioned attribute
# (due/ttl, default_case_owner, intake_settings, stageplan) creates or reuses a draft, writes the
# differences, and publishes. name and description are patched without publishing.
# division_id, customer_intent, reference_prefix and data_schema cannot change after the first publish.
resource "genesyscloud_case_management_caseplan" "example" {
  name                            = "Example caseplan"
  description                     = "Example case management caseplan"
  division_id                     = data.genesyscloud_auth_division_home.home.id
  reference_prefix                = "EXPL"
  default_due_duration_in_seconds = 1296000
  default_ttl_seconds             = 31536000

  customer_intent {
    id = genesyscloud_intents_customerintents.example_customer_intent.id
  }

  default_case_owner {
    id = genesyscloud_user.example_user.id
  }

  data_schema {
    id = genesyscloud_task_management_workitem_schema.example_schema.id
  }

  intake_settings {
    property      = "custom_attribute_1_text"
    required      = false
    display_order = 1
  }

  # 1-5 stageplans, in order. Each stageplan has exactly one stepplan.
  stageplan {
    name        = "Intake"
    description = "Capture the request"
    stepplan {
      name = "Triage"
    }
  }

  stageplan {
    name = "Investigate"
    stepplan {
      name          = "Investigate request"
      activity_type = "Workitem"
      workitem_settings {
        worktype_id = genesyscloud_task_management_worktype.example_worktype.id
      }
    }
  }

  stageplan {
    name = "Resolve"
    stepplan {
      name = "Close out"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `data_schema` (Block List, Min: 1, Max: 1) Task management workitem schema bound to case data for this caseplan. Publishing requires a data schema. Cannot be changed after the caseplan has been published. (see [below for nested schema](#nestedblock--data_schema))
- `name` (String) The name of the Caseplan.

### Optional

- `customer_intent` (Block List, Max: 1) The customer intent for the Cases created from the caseplan. Cannot be changed after the caseplan has been published. (see [below for nested schema](#nestedblock--customer_intent))
- `default_case_owner` (Block List, Max: 1) The default case owner for Cases created from the Caseplan. (see [below for nested schema](#nestedblock--default_case_owner))
- `default_due_duration_in_seconds` (Number) The default due duration in seconds for Cases created from the Caseplan.
- `default_ttl_seconds` (Number) The default TTL in seconds for Cases created from the Caseplan.
- `description` (String) The description of the Caseplan.
- `division_id` (String) The division to which this entity belongs. Cannot be changed after the caseplan has been published.
- `intake_settings` (Block List, Max: 10) Intake field configuration when collecting case data. Up to 10 entries. (see [below for nested schema](#nestedblock--intake_settings))
- `reference_prefix` (String) The prefix used when creating the reference for Cases from the Caseplan. Cannot be changed after the caseplan has been published.
- `stageplan` (Block List, Max: 5) Ordered stageplans; the block order is the process-flow order. A new caseplan starts with 3 stageplans, which the configured blocks are mapped onto by position. Stageplans are matched to existing ones by name, then by position; renaming and moving the same stageplan in one apply replaces it with a new id. When no blocks are configured, stageplans are not managed. (see [below for nested schema](#nestedblock--stageplan))

### Read-Only

- `has_draft` (Boolean) Whether an unpublished draft exists. A draft created outside Terraform is overwritten and published by the next apply that changes a versioned field.
- `id` (String) The ID of this resource.
- `published_version` (Number) The published version number. 0 when the caseplan has never been published.

<a id="nestedblock--data_schema"></a>
### Nested Schema for `data_schema`

Required:

- `id` (String) Workitem schema id.


<a id="nestedblock--customer_intent"></a>
### Nested Schema for `customer_intent`

Optional:

- `id` (String) Customer intent id (maps to customerIntentId on create).


<a id="nestedblock--default_case_owner"></a>
### Nested Schema for `default_case_owner`

Optional:

- `id` (String) User id for default case owner (maps to defaultCaseOwnerId on create).


<a id="nestedblock--intake_settings"></a>
### Nested Schema for `intake_settings`

Required:

- `property` (String) Property name from the bound workitem schema (data schema).

Optional:

- `display_order` (Number) Display order for this property in the intake UI. Defaults to `0`.
- `required` (Boolean) Whether this property is required at intake. Defaults to `false`.


<a id="nestedblock--stageplan"></a>
### Nested Schema for `stageplan`

Required:

- `name` (String) The name of the stageplan. Must be unique within the caseplan; stageplans are matched to existing ones by name, then by position.
- `stepplan` (Block List, Min: 1, Max: 1) The stepplan of this stageplan. The API supports exactly one stepplan per stageplan. (see [below for nested schema](#nestedblock--stageplan--stepplan))

Optional:

- `description` (String) The description of the stageplan.

Read-Only:

- `id` (String) Stageplan id. Stable across caseplan versions.

<a id="nestedblock--stageplan--stepplan"></a>
### Nested Schema for `stageplan.stepplan`

Required:

- `name` (String) The name of the stepplan.

Optional:

- `activity_type` (String) "None" completes the step as soon as it activates. "Workitem" creates a workitem from workitem_settings.worktype_id and completes when the workitem completes. Defaults to `None`.
- `description` (String) The description of the stepplan.
- `workitem_settings` (Block List, Max: 1) Workitem settings. Required when activity_type is "Workitem", and not allowed otherwise. (see [below for nested schema](#nestedblock--stageplan--stepplan--workitem_settings))

Read-Only:

- `id` (String) Stepplan id.

<a id="nestedblock--stageplan--stepplan--workitem_settings"></a>
### Nested Schema for `stageplan.stepplan.workitem_settings`

Required:

- `worktype_id` (String) Worktype id. The worktype's schema must be the caseplan's data schema, and its division must match the caseplan division (unless the caseplan division is "*").

