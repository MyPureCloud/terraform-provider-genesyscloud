---
page_title: "genesyscloud_routing_queue_conditional_group_routing Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud routing queue conditional group routing rules
---
# genesyscloud_routing_queue_conditional_group_routing (Resource)

Genesys Cloud routing queue conditional group routing rules

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/routing/queues/{queueId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-routing-queues--queueId-)
* [PUT /api/v2/routing/queues/{queueId}](https://developer.genesys.cloud/devapps/api-explorer#put-api-v2-routing-queues--queueId-)

## Example Usage

```terraform
// To enable this resource, set ENABLE_STANDALONE_CGR as an environment variable
// WARNING: This resource will overwrite any conditional group routing rules that already on the queue
// For this reason, all conditional group routing rules for a queue should be managed solely by this resource
resource "genesyscloud_routing_queue_conditional_group_routing" "example_queue_cgr" {
  queue_id = genesyscloud_routing_queue.example_queue.id
  rules {
    operator        = "LessThanOrEqualTo"
    metric          = "EstimatedWaitTime"
    condition_value = 0
    wait_seconds    = 20
    groups {
      member_group_id   = genesyscloud_group.example_group.id
      member_group_type = "GROUP"
    }
  }
  rules {
    evaluated_queue_id = genesyscloud_routing_queue.example_queue2.id
    operator           = "GreaterThanOrEqualTo"
    metric             = "EstimatedWaitTime"
    condition_value    = 5
    wait_seconds       = 15
    groups {
      member_group_id   = genesyscloud_group.example_group2.id
      member_group_type = "GROUP"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `queue_id` (String) Id of the routing queue to which the rules belong
- `rules` (Block List, Min: 1, Max: 5) The Conditional Group Routing settings for the queue. (see [below for nested schema](#nestedblock--rules))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--rules"></a>
### Nested Schema for `rules`

Required:

- `condition_value` (Number) The limit value, beyond which a rule evaluates as true.
- `groups` (Block Set, Min: 1) The group(s) to activate if the rule evaluates as true. (see [below for nested schema](#nestedblock--rules--groups))
- `operator` (String) The operator that compares the actual value against the condition value. Valid values: GreaterThan, GreaterThanOrEqualTo, LessThan, LessThanOrEqualTo.

Optional:

- `evaluated_queue_id` (String) The queue being evaluated for this rule. For rule 1, this is always the current queue, so should not be specified.
- `metric` (String) The queue metric being evaluated. Valid values: EstimatedWaitTime, ServiceLevel. Defaults to `EstimatedWaitTime`.
- `wait_seconds` (Number) The number of seconds to wait in this rule, if it evaluates as true, before evaluating the next rule. For the final rule, this is ignored, so need not be specified. Defaults to `2`.

<a id="nestedblock--rules--groups"></a>
### Nested Schema for `rules.groups`

Required:

- `member_group_id` (String) ID (GUID) for Group, SkillGroup, Team
- `member_group_type` (String) The type of the member group. Accepted values: TEAM, GROUP, SKILLGROUP

