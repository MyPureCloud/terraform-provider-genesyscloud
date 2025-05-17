---
page_title: "genesyscloud_task_management_worktype_status_transition Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud task management worktype status Transition
---
# genesyscloud_task_management_worktype_status_transition (Resource)

Genesys Cloud task management worktype status Transition

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/taskmanagement/worktypes/{worktypeId}/statuses](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-taskmanagement-worktypes--worktypeId--statuses)
* [POST /api/v2/taskmanagement/worktypes/{worktypeId}/statuses](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-taskmanagement-worktypes--worktypeId--statuses)
* [GET /api/v2/taskmanagement/worktypes/{worktypeId}/statuses/{statusId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-taskmanagement-worktypes--worktypeId--statuses--statusId-)
* [PATCH /api/v2/taskmanagement/worktypes/{worktypeId}/statuses/{statusId}](https://developer.genesys.cloud/devapps/api-explorer#patch-api-v2-taskmanagement-worktypes--worktypeId--statuses--statusId-)
* [DELETE /api/v2/taskmanagement/worktypes/{worktypeId}/statuses/{statusId}](https://developer.genesys.cloud/devapps/api-explorer#delete-api-v2-taskmanagement-worktypes--worktypeId--statuses--statusId-)



## Example Usage

```terraform
resource "genesyscloud_task_management_worktype_status_transition" "backlog" {
  worktype_id                     = genesyscloud_task_management_worktype.example_worktype.id
  status_id                       = genesyscloud_task_management_worktype_status.backlog.id
  destination_status_ids          = [genesyscloud_task_management_worktype_status.open.id, genesyscloud_task_management_worktype_status.working.id, genesyscloud_task_management_worktype_status.closed.id]
  default_destination_status_id   = genesyscloud_task_management_worktype_status.open.id
  status_transition_delay_seconds = 86500
  status_transition_time          = "04:20:00"

}
resource "genesyscloud_task_management_worktype_status_transition" "open" {
  worktype_id                     = genesyscloud_task_management_worktype.example_worktype.id
  status_id                       = genesyscloud_task_management_worktype_status.open.id
  destination_status_ids          = [genesyscloud_task_management_worktype_status.working.id, genesyscloud_task_management_worktype_status.waiting.id, genesyscloud_task_management_worktype_status.backlog.id, genesyscloud_task_management_worktype_status.resolved.id, genesyscloud_task_management_worktype_status.closed.id]
  default_destination_status_id   = genesyscloud_task_management_worktype_status.working.id
  status_transition_delay_seconds = 86500
  status_transition_time          = "04:20:00"

}

resource "genesyscloud_task_management_worktype_status_transition" "working" {
  worktype_id                     = genesyscloud_task_management_worktype.example_worktype.id
  status_id                       = genesyscloud_task_management_worktype_status.working.id
  destination_status_ids          = [genesyscloud_task_management_worktype_status.open.id, genesyscloud_task_management_worktype_status.waiting.id, genesyscloud_task_management_worktype_status.resolved.id, genesyscloud_task_management_worktype_status.closed.id]
  default_destination_status_id   = genesyscloud_task_management_worktype_status.waiting.id
  status_transition_delay_seconds = 86500
  status_transition_time          = "04:20:00"
}

resource "genesyscloud_task_management_worktype_status_transition" "waiting" {
  worktype_id                     = genesyscloud_task_management_worktype.example_worktype.id
  status_id                       = genesyscloud_task_management_worktype_status.waiting.id
  destination_status_ids          = [genesyscloud_task_management_worktype_status.working.id, genesyscloud_task_management_worktype_status.resolved.id, genesyscloud_task_management_worktype_status.closed.id]
  default_destination_status_id   = genesyscloud_task_management_worktype_status.working.id
  status_transition_delay_seconds = 86500
  status_transition_time          = "04:20:00"
}

resource "genesyscloud_task_management_worktype_status_transition" "resolved" {
  worktype_id            = genesyscloud_task_management_worktype.example_worktype.id
  status_id              = genesyscloud_task_management_worktype_status.resolved.id
  destination_status_ids = [genesyscloud_task_management_worktype_status.open.id, genesyscloud_task_management_worktype_status.backlog.id]
}

resource "genesyscloud_task_management_worktype_status_transition" "closed" {
  worktype_id            = genesyscloud_task_management_worktype.example_worktype.id
  status_id              = genesyscloud_task_management_worktype_status.closed.id
  destination_status_ids = [genesyscloud_task_management_worktype_status.open.id, genesyscloud_task_management_worktype_status.backlog.id]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `status_id` (String) Name of the status.
- `worktype_id` (String) The id of the worktype this status belongs to. Changing this attribute will cause the status to be dropped and recreated.

### Optional

- `default_destination_status_id` (String) Default destination status to which this Status will transition to if auto status transition enabled.
- `destination_status_ids` (List of String) A list of destination Statuses where a Workitem with this Status can transition to. If the list is empty Workitems with this Status can transition to all other Statuses defined on the Worktype. A Status can have a maximum of 24 destinations.
- `status_transition_delay_seconds` (Number) Delay in seconds for auto status transition. Required if default_destination_status_id is provided.
- `status_transition_time` (String) Time is represented as an ISO-8601 string without a timezone. For example: HH:mm:ss

### Read-Only

- `id` (String) The ID of this resource.

