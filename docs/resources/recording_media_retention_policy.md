---
page_title: "genesyscloud_recording_media_retention_policy Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Media Retention Policies
---
# genesyscloud_recording_media_retention_policy (Resource)

Genesys Cloud Media Retention Policies

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/recording/mediaretentionpolicies](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-recording-mediaretentionpolicies)
* [POST /api/v2/recording/mediaretentionpolicies](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-recording-mediaretentionpolicies)
* [GET /api/v2/recording/mediaretentionpolicies/{policyId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-recording-mediaretentionpolicies--policyId-)
* [PUT /api/v2/recording/mediaretentionpolicies/{policyId}](https://developer.genesys.cloud/devapps/api-explorer#put-api-v2-recording-mediaretentionpolicies--policyId-)
* [DELETE /api/v2/recording/mediaretentionpolicies/{policyId}](https://developer.genesys.cloud/devapps/api-explorer#delete-api-v2-recording-mediaretentionpolicies--policyId-)
* [GET /api/v2/quality/forms/evaluations/{formId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-quality-forms-evaluations--formId-)
* [GET /api/v2/quality/forms/evaluations/{formId}/versions](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-quality-forms-evaluations--formId--versions)
* [GET /api/v2/quality/forms/surveys](https://developer.genesys.cloud/api/rest/v2/quality/#get-api-v2-quality-forms-surveys)


## Example Usage

```terraform
resource "genesyscloud_recording_media_retention_policy" "example-media-retention-policy" {
  name        = "example-media-retention-policy"
  order       = 1
  description = "a media retention policy"
  enabled     = true
  media_policies {
    call_policy {
      actions {
        retain_recording = true
        delete_recording = false
        always_delete    = false
        assign_evaluations {
          evaluation_form_id = genesyscloud_quality_forms_evaluation.example-evaluation-form.id
          user_id            = genesyscloud_user.example-user.id
        }
        assign_metered_evaluations {
          evaluator_ids          = [genesyscloud_user.example-user.id]
          max_number_evaluations = 1
          evaluation_form_id     = genesyscloud_quality_forms_evaluation.example-evaluation-form.id
          assign_to_active_user  = true
          time_interval {
            months = 1
            weeks  = 1
            days   = 1
            hours  = 1
          }
        }
        assign_metered_assignment_by_agent {
          evaluator_ids          = [genesyscloud_user.example-user.id]
          max_number_evaluations = 1
          evaluation_form_id     = genesyscloud_quality_forms_evaluation.example-evaluation-form.id
          time_interval {
            months = 1
            weeks  = 1
            days   = 1
            hours  = 1
          }
          time_zone = "EST"
        }
        assign_calibrations {
          calibrator_id       = genesyscloud_user.example-user.id
          evaluator_ids       = [genesyscloud_user.example-user.id]
          evaluation_form_id  = genesyscloud_quality_forms_evaluation.example-evaluation-form.id
          expert_evaluator_id = genesyscloud_user.example-user.id
        }
        assign_surveys {
          sending_domain   = genesyscloud_routing_email_domain.routing-domain.domain_id
          survey_form_name = "survey-form-name"
          flow_id          = genesyscloud_flow.example-flow-resource.id
        }
        retention_duration {
          archive_retention {
            days           = 1
            storage_medium = "CLOUDARCHIVE"
          }
          delete_retention {
            days = 3
          }
        }
        initiate_screen_recording {
          record_acw = true
          archive_retention {
            days           = 1
            storage_medium = "CLOUDARCHIVE"
          }
          delete_retention {
            days = 3
          }

        }
        integration_export {
          integration_id                  = genesyscloud_integration.example-integration-resource.id
          should_export_screen_recordings = true
        }
      }
      conditions {
        for_user_ids    = [genesyscloud_user.example-user.id]
        date_ranges     = ["2022-05-12T04:00:00.000Z/2022-05-13T04:00:00.000Z"]
        for_queue_ids   = [genesyscloud_routing_queue.example-queue.id]
        wrapup_code_ids = [genesyscloud_routing_wrapupcode.example-wrapup-code.id]
        language_ids    = [genesyscloud_routing_language.example-language.id]
        time_allowed {
          time_slots {
            start_time = "10:10:10.010"
            stop_time  = "11:11:11.011"
            day        = 3
          }
          time_zone_id = "Europe/Paris"
          empty        = false
        }
        directions = ["INBOUND"]
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The policy name. Changing the policy_name attribute will cause the recording_media_retention_policy to be dropped and recreated with a new ID.

### Optional

- `actions` (Block List, Max: 1) Actions (see [below for nested schema](#nestedblock--actions))
- `conditions` (Block List, Max: 1) Conditions (see [below for nested schema](#nestedblock--conditions))
- `description` (String) The description for the policy
- `enabled` (Boolean) The policy will be enabled if true, otherwise it will be disabled
- `media_policies` (Block List, Max: 1) Conditions and actions per media type (see [below for nested schema](#nestedblock--media_policies))
- `order` (Number) The ordinal number for the policy
- `policy_errors` (Block List, Max: 1) A list of errors in the policy configuration (see [below for nested schema](#nestedblock--policy_errors))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--actions"></a>
### Nested Schema for `actions`

Optional:

- `always_delete` (Boolean) true to delete the recording associated with the conversation regardless of the values of retainRecording or deleteRecording.
- `assign_calibrations` (Block List) (see [below for nested schema](#nestedblock--actions--assign_calibrations))
- `assign_evaluations` (Block List) (see [below for nested schema](#nestedblock--actions--assign_evaluations))
- `assign_metered_assignment_by_agent` (Block List) (see [below for nested schema](#nestedblock--actions--assign_metered_assignment_by_agent))
- `assign_metered_evaluations` (Block List) (see [below for nested schema](#nestedblock--actions--assign_metered_evaluations))
- `assign_surveys` (Block List) (see [below for nested schema](#nestedblock--actions--assign_surveys))
- `delete_recording` (Boolean) true to delete the recording associated with the conversation. If retainRecording = true, this will be ignored.
- `initiate_screen_recording` (Block List, Max: 1) (see [below for nested schema](#nestedblock--actions--initiate_screen_recording))
- `integration_export` (Block List, Max: 1) Policy action for exporting recordings using an integration to 3rd party s3. (see [below for nested schema](#nestedblock--actions--integration_export))
- `media_transcriptions` (Block List) (see [below for nested schema](#nestedblock--actions--media_transcriptions))
- `retain_recording` (Boolean) true to retain the recording associated with the conversation.
- `retention_duration` (Block List, Max: 1) (see [below for nested schema](#nestedblock--actions--retention_duration))

<a id="nestedblock--actions--assign_calibrations"></a>
### Nested Schema for `actions.assign_calibrations`

Optional:

- `calibrator_id` (String)
- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `expert_evaluator_id` (String)


<a id="nestedblock--actions--assign_evaluations"></a>
### Nested Schema for `actions.assign_evaluations`

Optional:

- `evaluation_form_id` (String)
- `user_id` (String)


<a id="nestedblock--actions--assign_metered_assignment_by_agent"></a>
### Nested Schema for `actions.assign_metered_assignment_by_agent`

Optional:

- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `max_number_evaluations` (Number)
- `time_interval` (Block List, Max: 1) (see [below for nested schema](#nestedblock--actions--assign_metered_assignment_by_agent--time_interval))
- `time_zone` (String)

<a id="nestedblock--actions--assign_metered_assignment_by_agent--time_interval"></a>
### Nested Schema for `actions.assign_metered_assignment_by_agent.time_interval`

Optional:

- `days` (Number)
- `months` (Number)
- `weeks` (Number)



<a id="nestedblock--actions--assign_metered_evaluations"></a>
### Nested Schema for `actions.assign_metered_evaluations`

Optional:

- `assign_to_active_user` (Boolean)
- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `max_number_evaluations` (Number)
- `time_interval` (Block List, Max: 1) (see [below for nested schema](#nestedblock--actions--assign_metered_evaluations--time_interval))

<a id="nestedblock--actions--assign_metered_evaluations--time_interval"></a>
### Nested Schema for `actions.assign_metered_evaluations.time_interval`

Optional:

- `days` (Number)
- `hours` (Number)



<a id="nestedblock--actions--assign_surveys"></a>
### Nested Schema for `actions.assign_surveys`

Required:

- `sending_domain` (String) Validated email domain, required

Optional:

- `flow_id` (String) The UUID reference to the flow associated with this survey.
- `invite_time_interval` (String) An ISO 8601 repeated interval consisting of the number of repetitions, the start datetime, and the interval (e.g. R2/2018-03-01T13:00:00Z/P1M10DT2H30M). Total duration must not exceed 90 days. Defaults to `R1/P0M`.
- `sending_user` (String) User together with sendingDomain used to send email, null to use no-reply
- `survey_form_name` (String) The survey form used for this survey.


<a id="nestedblock--actions--initiate_screen_recording"></a>
### Nested Schema for `actions.initiate_screen_recording`

Optional:

- `archive_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--actions--initiate_screen_recording--archive_retention))
- `delete_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--actions--initiate_screen_recording--delete_retention))
- `record_acw` (Boolean)

<a id="nestedblock--actions--initiate_screen_recording--archive_retention"></a>
### Nested Schema for `actions.initiate_screen_recording.archive_retention`

Optional:

- `days` (Number)
- `storage_medium` (String)


<a id="nestedblock--actions--initiate_screen_recording--delete_retention"></a>
### Nested Schema for `actions.initiate_screen_recording.delete_retention`

Optional:

- `days` (Number)



<a id="nestedblock--actions--integration_export"></a>
### Nested Schema for `actions.integration_export`

Optional:

- `integration_id` (String) The aws-s3-recording-bulk-actions-integration that the policy uses for exports.
- `should_export_screen_recordings` (Boolean) True if the policy should export screen recordings in addition to the other conversation media. Defaults to `true`.


<a id="nestedblock--actions--media_transcriptions"></a>
### Nested Schema for `actions.media_transcriptions`

Optional:

- `display_name` (String)
- `integration_id` (String)
- `transcription_provider` (String)


<a id="nestedblock--actions--retention_duration"></a>
### Nested Schema for `actions.retention_duration`

Optional:

- `archive_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--actions--retention_duration--archive_retention))
- `delete_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--actions--retention_duration--delete_retention))

<a id="nestedblock--actions--retention_duration--archive_retention"></a>
### Nested Schema for `actions.retention_duration.archive_retention`

Optional:

- `days` (Number)
- `storage_medium` (String)


<a id="nestedblock--actions--retention_duration--delete_retention"></a>
### Nested Schema for `actions.retention_duration.delete_retention`

Optional:

- `days` (Number)




<a id="nestedblock--conditions"></a>
### Nested Schema for `conditions`

Optional:

- `date_ranges` (List of String)
- `directions` (List of String)
- `duration` (Block List, Max: 1) (see [below for nested schema](#nestedblock--conditions--duration))
- `for_queue_ids` (List of String)
- `for_user_ids` (List of String)
- `media_types` (List of String)
- `time_allowed` (Block List, Max: 1) (see [below for nested schema](#nestedblock--conditions--time_allowed))
- `wrapup_code_ids` (List of String)

<a id="nestedblock--conditions--duration"></a>
### Nested Schema for `conditions.duration`

Optional:

- `duration_mode` (String)
- `duration_operator` (String)
- `duration_range` (String)
- `duration_target` (String)


<a id="nestedblock--conditions--time_allowed"></a>
### Nested Schema for `conditions.time_allowed`

Optional:

- `empty` (Boolean)
- `time_slots` (Block List) (see [below for nested schema](#nestedblock--conditions--time_allowed--time_slots))
- `time_zone_id` (String)

<a id="nestedblock--conditions--time_allowed--time_slots"></a>
### Nested Schema for `conditions.time_allowed.time_slots`

Optional:

- `day` (Number) Day for this time slot, Monday = 1 ... Sunday = 7
- `start_time` (String) start time in xx:xx:xx.xxx format
- `stop_time` (String) stop time in xx:xx:xx.xxx format




<a id="nestedblock--media_policies"></a>
### Nested Schema for `media_policies`

Optional:

- `call_policy` (Block List, Max: 1) Conditions and actions for calls (see [below for nested schema](#nestedblock--media_policies--call_policy))
- `chat_policy` (Block List, Max: 1) Conditions and actions for calls (see [below for nested schema](#nestedblock--media_policies--chat_policy))
- `email_policy` (Block List, Max: 1) Conditions and actions for calls (see [below for nested schema](#nestedblock--media_policies--email_policy))
- `message_policy` (Block List, Max: 1) Conditions and actions for calls (see [below for nested schema](#nestedblock--media_policies--message_policy))

<a id="nestedblock--media_policies--call_policy"></a>
### Nested Schema for `media_policies.call_policy`

Optional:

- `actions` (Block List, Max: 1) Actions applied when specified conditions are met (see [below for nested schema](#nestedblock--media_policies--call_policy--actions))
- `conditions` (Block List, Max: 1) Conditions for when actions should be applied (see [below for nested schema](#nestedblock--media_policies--call_policy--conditions))

<a id="nestedblock--media_policies--call_policy--actions"></a>
### Nested Schema for `media_policies.call_policy.actions`

Optional:

- `always_delete` (Boolean) true to delete the recording associated with the conversation regardless of the values of retainRecording or deleteRecording.
- `assign_calibrations` (Block List) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--assign_calibrations))
- `assign_evaluations` (Block List) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--assign_evaluations))
- `assign_metered_assignment_by_agent` (Block List) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--assign_metered_assignment_by_agent))
- `assign_metered_evaluations` (Block List) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--assign_metered_evaluations))
- `assign_surveys` (Block List) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--assign_surveys))
- `delete_recording` (Boolean) true to delete the recording associated with the conversation. If retainRecording = true, this will be ignored.
- `initiate_screen_recording` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--initiate_screen_recording))
- `integration_export` (Block List, Max: 1) Policy action for exporting recordings using an integration to 3rd party s3. (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--integration_export))
- `media_transcriptions` (Block List) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--media_transcriptions))
- `retain_recording` (Boolean) true to retain the recording associated with the conversation.
- `retention_duration` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--retention_duration))

<a id="nestedblock--media_policies--call_policy--actions--assign_calibrations"></a>
### Nested Schema for `media_policies.call_policy.actions.assign_calibrations`

Optional:

- `calibrator_id` (String)
- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `expert_evaluator_id` (String)


<a id="nestedblock--media_policies--call_policy--actions--assign_evaluations"></a>
### Nested Schema for `media_policies.call_policy.actions.assign_evaluations`

Optional:

- `evaluation_form_id` (String)
- `user_id` (String)


<a id="nestedblock--media_policies--call_policy--actions--assign_metered_assignment_by_agent"></a>
### Nested Schema for `media_policies.call_policy.actions.assign_metered_assignment_by_agent`

Optional:

- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `max_number_evaluations` (Number)
- `time_interval` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--assign_metered_assignment_by_agent--time_interval))
- `time_zone` (String)

<a id="nestedblock--media_policies--call_policy--actions--assign_metered_assignment_by_agent--time_interval"></a>
### Nested Schema for `media_policies.call_policy.actions.assign_metered_assignment_by_agent.time_interval`

Optional:

- `days` (Number)
- `months` (Number)
- `weeks` (Number)



<a id="nestedblock--media_policies--call_policy--actions--assign_metered_evaluations"></a>
### Nested Schema for `media_policies.call_policy.actions.assign_metered_evaluations`

Optional:

- `assign_to_active_user` (Boolean)
- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `max_number_evaluations` (Number)
- `time_interval` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--assign_metered_evaluations--time_interval))

<a id="nestedblock--media_policies--call_policy--actions--assign_metered_evaluations--time_interval"></a>
### Nested Schema for `media_policies.call_policy.actions.assign_metered_evaluations.time_interval`

Optional:

- `days` (Number)
- `hours` (Number)



<a id="nestedblock--media_policies--call_policy--actions--assign_surveys"></a>
### Nested Schema for `media_policies.call_policy.actions.assign_surveys`

Required:

- `sending_domain` (String) Validated email domain, required

Optional:

- `flow_id` (String) The UUID reference to the flow associated with this survey.
- `invite_time_interval` (String) An ISO 8601 repeated interval consisting of the number of repetitions, the start datetime, and the interval (e.g. R2/2018-03-01T13:00:00Z/P1M10DT2H30M). Total duration must not exceed 90 days. Defaults to `R1/P0M`.
- `sending_user` (String) User together with sendingDomain used to send email, null to use no-reply
- `survey_form_name` (String) The survey form used for this survey.


<a id="nestedblock--media_policies--call_policy--actions--initiate_screen_recording"></a>
### Nested Schema for `media_policies.call_policy.actions.initiate_screen_recording`

Optional:

- `archive_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--initiate_screen_recording--archive_retention))
- `delete_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--initiate_screen_recording--delete_retention))
- `record_acw` (Boolean)

<a id="nestedblock--media_policies--call_policy--actions--initiate_screen_recording--archive_retention"></a>
### Nested Schema for `media_policies.call_policy.actions.initiate_screen_recording.archive_retention`

Optional:

- `days` (Number)
- `storage_medium` (String)


<a id="nestedblock--media_policies--call_policy--actions--initiate_screen_recording--delete_retention"></a>
### Nested Schema for `media_policies.call_policy.actions.initiate_screen_recording.delete_retention`

Optional:

- `days` (Number)



<a id="nestedblock--media_policies--call_policy--actions--integration_export"></a>
### Nested Schema for `media_policies.call_policy.actions.integration_export`

Optional:

- `integration_id` (String) The aws-s3-recording-bulk-actions-integration that the policy uses for exports.
- `should_export_screen_recordings` (Boolean) True if the policy should export screen recordings in addition to the other conversation media. Defaults to `true`.


<a id="nestedblock--media_policies--call_policy--actions--media_transcriptions"></a>
### Nested Schema for `media_policies.call_policy.actions.media_transcriptions`

Optional:

- `display_name` (String)
- `integration_id` (String)
- `transcription_provider` (String)


<a id="nestedblock--media_policies--call_policy--actions--retention_duration"></a>
### Nested Schema for `media_policies.call_policy.actions.retention_duration`

Optional:

- `archive_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--retention_duration--archive_retention))
- `delete_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--call_policy--actions--retention_duration--delete_retention))

<a id="nestedblock--media_policies--call_policy--actions--retention_duration--archive_retention"></a>
### Nested Schema for `media_policies.call_policy.actions.retention_duration.archive_retention`

Optional:

- `days` (Number)
- `storage_medium` (String)


<a id="nestedblock--media_policies--call_policy--actions--retention_duration--delete_retention"></a>
### Nested Schema for `media_policies.call_policy.actions.retention_duration.delete_retention`

Optional:

- `days` (Number)




<a id="nestedblock--media_policies--call_policy--conditions"></a>
### Nested Schema for `media_policies.call_policy.conditions`

Optional:

- `date_ranges` (List of String)
- `directions` (List of String)
- `duration` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--call_policy--conditions--duration))
- `for_queue_ids` (List of String)
- `for_user_ids` (List of String)
- `language_ids` (List of String)
- `time_allowed` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--call_policy--conditions--time_allowed))
- `wrapup_code_ids` (List of String)

<a id="nestedblock--media_policies--call_policy--conditions--duration"></a>
### Nested Schema for `media_policies.call_policy.conditions.duration`

Optional:

- `duration_mode` (String)
- `duration_operator` (String)
- `duration_range` (String)
- `duration_target` (String)


<a id="nestedblock--media_policies--call_policy--conditions--time_allowed"></a>
### Nested Schema for `media_policies.call_policy.conditions.time_allowed`

Optional:

- `empty` (Boolean)
- `time_slots` (Block List) (see [below for nested schema](#nestedblock--media_policies--call_policy--conditions--time_allowed--time_slots))
- `time_zone_id` (String)

<a id="nestedblock--media_policies--call_policy--conditions--time_allowed--time_slots"></a>
### Nested Schema for `media_policies.call_policy.conditions.time_allowed.time_slots`

Optional:

- `day` (Number) Day for this time slot, Monday = 1 ... Sunday = 7
- `start_time` (String) start time in xx:xx:xx.xxx format
- `stop_time` (String) stop time in xx:xx:xx.xxx format





<a id="nestedblock--media_policies--chat_policy"></a>
### Nested Schema for `media_policies.chat_policy`

Optional:

- `actions` (Block List, Max: 1) Actions applied when specified conditions are met (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions))
- `conditions` (Block List, Max: 1) Conditions for when actions should be applied (see [below for nested schema](#nestedblock--media_policies--chat_policy--conditions))

<a id="nestedblock--media_policies--chat_policy--actions"></a>
### Nested Schema for `media_policies.chat_policy.actions`

Optional:

- `always_delete` (Boolean) true to delete the recording associated with the conversation regardless of the values of retainRecording or deleteRecording.
- `assign_calibrations` (Block List) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--assign_calibrations))
- `assign_evaluations` (Block List) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--assign_evaluations))
- `assign_metered_assignment_by_agent` (Block List) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--assign_metered_assignment_by_agent))
- `assign_metered_evaluations` (Block List) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--assign_metered_evaluations))
- `assign_surveys` (Block List) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--assign_surveys))
- `delete_recording` (Boolean) true to delete the recording associated with the conversation. If retainRecording = true, this will be ignored.
- `initiate_screen_recording` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--initiate_screen_recording))
- `integration_export` (Block List, Max: 1) Policy action for exporting recordings using an integration to 3rd party s3. (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--integration_export))
- `media_transcriptions` (Block List) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--media_transcriptions))
- `retain_recording` (Boolean) true to retain the recording associated with the conversation.
- `retention_duration` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--retention_duration))

<a id="nestedblock--media_policies--chat_policy--actions--assign_calibrations"></a>
### Nested Schema for `media_policies.chat_policy.actions.assign_calibrations`

Optional:

- `calibrator_id` (String)
- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `expert_evaluator_id` (String)


<a id="nestedblock--media_policies--chat_policy--actions--assign_evaluations"></a>
### Nested Schema for `media_policies.chat_policy.actions.assign_evaluations`

Optional:

- `evaluation_form_id` (String)
- `user_id` (String)


<a id="nestedblock--media_policies--chat_policy--actions--assign_metered_assignment_by_agent"></a>
### Nested Schema for `media_policies.chat_policy.actions.assign_metered_assignment_by_agent`

Optional:

- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `max_number_evaluations` (Number)
- `time_interval` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--assign_metered_assignment_by_agent--time_interval))
- `time_zone` (String)

<a id="nestedblock--media_policies--chat_policy--actions--assign_metered_assignment_by_agent--time_interval"></a>
### Nested Schema for `media_policies.chat_policy.actions.assign_metered_assignment_by_agent.time_interval`

Optional:

- `days` (Number)
- `months` (Number)
- `weeks` (Number)



<a id="nestedblock--media_policies--chat_policy--actions--assign_metered_evaluations"></a>
### Nested Schema for `media_policies.chat_policy.actions.assign_metered_evaluations`

Optional:

- `assign_to_active_user` (Boolean)
- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `max_number_evaluations` (Number)
- `time_interval` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--assign_metered_evaluations--time_interval))

<a id="nestedblock--media_policies--chat_policy--actions--assign_metered_evaluations--time_interval"></a>
### Nested Schema for `media_policies.chat_policy.actions.assign_metered_evaluations.time_interval`

Optional:

- `days` (Number)
- `hours` (Number)



<a id="nestedblock--media_policies--chat_policy--actions--assign_surveys"></a>
### Nested Schema for `media_policies.chat_policy.actions.assign_surveys`

Required:

- `sending_domain` (String) Validated email domain, required

Optional:

- `flow_id` (String) The UUID reference to the flow associated with this survey.
- `invite_time_interval` (String) An ISO 8601 repeated interval consisting of the number of repetitions, the start datetime, and the interval (e.g. R2/2018-03-01T13:00:00Z/P1M10DT2H30M). Total duration must not exceed 90 days. Defaults to `R1/P0M`.
- `sending_user` (String) User together with sendingDomain used to send email, null to use no-reply
- `survey_form_name` (String) The survey form used for this survey.


<a id="nestedblock--media_policies--chat_policy--actions--initiate_screen_recording"></a>
### Nested Schema for `media_policies.chat_policy.actions.initiate_screen_recording`

Optional:

- `archive_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--initiate_screen_recording--archive_retention))
- `delete_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--initiate_screen_recording--delete_retention))
- `record_acw` (Boolean)

<a id="nestedblock--media_policies--chat_policy--actions--initiate_screen_recording--archive_retention"></a>
### Nested Schema for `media_policies.chat_policy.actions.initiate_screen_recording.archive_retention`

Optional:

- `days` (Number)
- `storage_medium` (String)


<a id="nestedblock--media_policies--chat_policy--actions--initiate_screen_recording--delete_retention"></a>
### Nested Schema for `media_policies.chat_policy.actions.initiate_screen_recording.delete_retention`

Optional:

- `days` (Number)



<a id="nestedblock--media_policies--chat_policy--actions--integration_export"></a>
### Nested Schema for `media_policies.chat_policy.actions.integration_export`

Optional:

- `integration_id` (String) The aws-s3-recording-bulk-actions-integration that the policy uses for exports.
- `should_export_screen_recordings` (Boolean) True if the policy should export screen recordings in addition to the other conversation media. Defaults to `true`.


<a id="nestedblock--media_policies--chat_policy--actions--media_transcriptions"></a>
### Nested Schema for `media_policies.chat_policy.actions.media_transcriptions`

Optional:

- `display_name` (String)
- `integration_id` (String)
- `transcription_provider` (String)


<a id="nestedblock--media_policies--chat_policy--actions--retention_duration"></a>
### Nested Schema for `media_policies.chat_policy.actions.retention_duration`

Optional:

- `archive_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--retention_duration--archive_retention))
- `delete_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--chat_policy--actions--retention_duration--delete_retention))

<a id="nestedblock--media_policies--chat_policy--actions--retention_duration--archive_retention"></a>
### Nested Schema for `media_policies.chat_policy.actions.retention_duration.archive_retention`

Optional:

- `days` (Number)
- `storage_medium` (String)


<a id="nestedblock--media_policies--chat_policy--actions--retention_duration--delete_retention"></a>
### Nested Schema for `media_policies.chat_policy.actions.retention_duration.delete_retention`

Optional:

- `days` (Number)




<a id="nestedblock--media_policies--chat_policy--conditions"></a>
### Nested Schema for `media_policies.chat_policy.conditions`

Optional:

- `date_ranges` (List of String)
- `duration` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--chat_policy--conditions--duration))
- `for_queue_ids` (List of String)
- `for_user_ids` (List of String)
- `language_ids` (List of String)
- `time_allowed` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--chat_policy--conditions--time_allowed))
- `wrapup_code_ids` (List of String)

<a id="nestedblock--media_policies--chat_policy--conditions--duration"></a>
### Nested Schema for `media_policies.chat_policy.conditions.duration`

Optional:

- `duration_mode` (String)
- `duration_operator` (String)
- `duration_range` (String)
- `duration_target` (String)


<a id="nestedblock--media_policies--chat_policy--conditions--time_allowed"></a>
### Nested Schema for `media_policies.chat_policy.conditions.time_allowed`

Optional:

- `empty` (Boolean)
- `time_slots` (Block List) (see [below for nested schema](#nestedblock--media_policies--chat_policy--conditions--time_allowed--time_slots))
- `time_zone_id` (String)

<a id="nestedblock--media_policies--chat_policy--conditions--time_allowed--time_slots"></a>
### Nested Schema for `media_policies.chat_policy.conditions.time_allowed.time_slots`

Optional:

- `day` (Number) Day for this time slot, Monday = 1 ... Sunday = 7
- `start_time` (String) start time in xx:xx:xx.xxx format
- `stop_time` (String) stop time in xx:xx:xx.xxx format





<a id="nestedblock--media_policies--email_policy"></a>
### Nested Schema for `media_policies.email_policy`

Optional:

- `actions` (Block List, Max: 1) Actions applied when specified conditions are met (see [below for nested schema](#nestedblock--media_policies--email_policy--actions))
- `conditions` (Block List, Max: 1) Conditions for when actions should be applied (see [below for nested schema](#nestedblock--media_policies--email_policy--conditions))

<a id="nestedblock--media_policies--email_policy--actions"></a>
### Nested Schema for `media_policies.email_policy.actions`

Optional:

- `always_delete` (Boolean) true to delete the recording associated with the conversation regardless of the values of retainRecording or deleteRecording.
- `assign_calibrations` (Block List) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--assign_calibrations))
- `assign_evaluations` (Block List) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--assign_evaluations))
- `assign_metered_assignment_by_agent` (Block List) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--assign_metered_assignment_by_agent))
- `assign_metered_evaluations` (Block List) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--assign_metered_evaluations))
- `assign_surveys` (Block List) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--assign_surveys))
- `delete_recording` (Boolean) true to delete the recording associated with the conversation. If retainRecording = true, this will be ignored.
- `initiate_screen_recording` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--initiate_screen_recording))
- `integration_export` (Block List, Max: 1) Policy action for exporting recordings using an integration to 3rd party s3. (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--integration_export))
- `media_transcriptions` (Block List) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--media_transcriptions))
- `retain_recording` (Boolean) true to retain the recording associated with the conversation.
- `retention_duration` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--retention_duration))

<a id="nestedblock--media_policies--email_policy--actions--assign_calibrations"></a>
### Nested Schema for `media_policies.email_policy.actions.assign_calibrations`

Optional:

- `calibrator_id` (String)
- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `expert_evaluator_id` (String)


<a id="nestedblock--media_policies--email_policy--actions--assign_evaluations"></a>
### Nested Schema for `media_policies.email_policy.actions.assign_evaluations`

Optional:

- `evaluation_form_id` (String)
- `user_id` (String)


<a id="nestedblock--media_policies--email_policy--actions--assign_metered_assignment_by_agent"></a>
### Nested Schema for `media_policies.email_policy.actions.assign_metered_assignment_by_agent`

Optional:

- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `max_number_evaluations` (Number)
- `time_interval` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--assign_metered_assignment_by_agent--time_interval))
- `time_zone` (String)

<a id="nestedblock--media_policies--email_policy--actions--assign_metered_assignment_by_agent--time_interval"></a>
### Nested Schema for `media_policies.email_policy.actions.assign_metered_assignment_by_agent.time_interval`

Optional:

- `days` (Number)
- `months` (Number)
- `weeks` (Number)



<a id="nestedblock--media_policies--email_policy--actions--assign_metered_evaluations"></a>
### Nested Schema for `media_policies.email_policy.actions.assign_metered_evaluations`

Optional:

- `assign_to_active_user` (Boolean)
- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `max_number_evaluations` (Number)
- `time_interval` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--assign_metered_evaluations--time_interval))

<a id="nestedblock--media_policies--email_policy--actions--assign_metered_evaluations--time_interval"></a>
### Nested Schema for `media_policies.email_policy.actions.assign_metered_evaluations.time_interval`

Optional:

- `days` (Number)
- `hours` (Number)



<a id="nestedblock--media_policies--email_policy--actions--assign_surveys"></a>
### Nested Schema for `media_policies.email_policy.actions.assign_surveys`

Required:

- `sending_domain` (String) Validated email domain, required

Optional:

- `flow_id` (String) The UUID reference to the flow associated with this survey.
- `invite_time_interval` (String) An ISO 8601 repeated interval consisting of the number of repetitions, the start datetime, and the interval (e.g. R2/2018-03-01T13:00:00Z/P1M10DT2H30M). Total duration must not exceed 90 days. Defaults to `R1/P0M`.
- `sending_user` (String) User together with sendingDomain used to send email, null to use no-reply
- `survey_form_name` (String) The survey form used for this survey.


<a id="nestedblock--media_policies--email_policy--actions--initiate_screen_recording"></a>
### Nested Schema for `media_policies.email_policy.actions.initiate_screen_recording`

Optional:

- `archive_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--initiate_screen_recording--archive_retention))
- `delete_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--initiate_screen_recording--delete_retention))
- `record_acw` (Boolean)

<a id="nestedblock--media_policies--email_policy--actions--initiate_screen_recording--archive_retention"></a>
### Nested Schema for `media_policies.email_policy.actions.initiate_screen_recording.archive_retention`

Optional:

- `days` (Number)
- `storage_medium` (String)


<a id="nestedblock--media_policies--email_policy--actions--initiate_screen_recording--delete_retention"></a>
### Nested Schema for `media_policies.email_policy.actions.initiate_screen_recording.delete_retention`

Optional:

- `days` (Number)



<a id="nestedblock--media_policies--email_policy--actions--integration_export"></a>
### Nested Schema for `media_policies.email_policy.actions.integration_export`

Optional:

- `integration_id` (String) The aws-s3-recording-bulk-actions-integration that the policy uses for exports.
- `should_export_screen_recordings` (Boolean) True if the policy should export screen recordings in addition to the other conversation media. Defaults to `true`.


<a id="nestedblock--media_policies--email_policy--actions--media_transcriptions"></a>
### Nested Schema for `media_policies.email_policy.actions.media_transcriptions`

Optional:

- `display_name` (String)
- `integration_id` (String)
- `transcription_provider` (String)


<a id="nestedblock--media_policies--email_policy--actions--retention_duration"></a>
### Nested Schema for `media_policies.email_policy.actions.retention_duration`

Optional:

- `archive_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--retention_duration--archive_retention))
- `delete_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--email_policy--actions--retention_duration--delete_retention))

<a id="nestedblock--media_policies--email_policy--actions--retention_duration--archive_retention"></a>
### Nested Schema for `media_policies.email_policy.actions.retention_duration.archive_retention`

Optional:

- `days` (Number)
- `storage_medium` (String)


<a id="nestedblock--media_policies--email_policy--actions--retention_duration--delete_retention"></a>
### Nested Schema for `media_policies.email_policy.actions.retention_duration.delete_retention`

Optional:

- `days` (Number)




<a id="nestedblock--media_policies--email_policy--conditions"></a>
### Nested Schema for `media_policies.email_policy.conditions`

Optional:

- `date_ranges` (List of String)
- `for_queue_ids` (List of String)
- `for_user_ids` (List of String)
- `language_ids` (List of String)
- `time_allowed` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--email_policy--conditions--time_allowed))
- `wrapup_code_ids` (List of String)

<a id="nestedblock--media_policies--email_policy--conditions--time_allowed"></a>
### Nested Schema for `media_policies.email_policy.conditions.time_allowed`

Optional:

- `empty` (Boolean)
- `time_slots` (Block List) (see [below for nested schema](#nestedblock--media_policies--email_policy--conditions--time_allowed--time_slots))
- `time_zone_id` (String)

<a id="nestedblock--media_policies--email_policy--conditions--time_allowed--time_slots"></a>
### Nested Schema for `media_policies.email_policy.conditions.time_allowed.time_slots`

Optional:

- `day` (Number) Day for this time slot, Monday = 1 ... Sunday = 7
- `start_time` (String) start time in xx:xx:xx.xxx format
- `stop_time` (String) stop time in xx:xx:xx.xxx format





<a id="nestedblock--media_policies--message_policy"></a>
### Nested Schema for `media_policies.message_policy`

Optional:

- `actions` (Block List, Max: 1) Actions applied when specified conditions are met (see [below for nested schema](#nestedblock--media_policies--message_policy--actions))
- `conditions` (Block List, Max: 1) Conditions for when actions should be applied (see [below for nested schema](#nestedblock--media_policies--message_policy--conditions))

<a id="nestedblock--media_policies--message_policy--actions"></a>
### Nested Schema for `media_policies.message_policy.actions`

Optional:

- `always_delete` (Boolean) true to delete the recording associated with the conversation regardless of the values of retainRecording or deleteRecording.
- `assign_calibrations` (Block List) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--assign_calibrations))
- `assign_evaluations` (Block List) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--assign_evaluations))
- `assign_metered_assignment_by_agent` (Block List) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--assign_metered_assignment_by_agent))
- `assign_metered_evaluations` (Block List) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--assign_metered_evaluations))
- `assign_surveys` (Block List) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--assign_surveys))
- `delete_recording` (Boolean) true to delete the recording associated with the conversation. If retainRecording = true, this will be ignored.
- `initiate_screen_recording` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--initiate_screen_recording))
- `integration_export` (Block List, Max: 1) Policy action for exporting recordings using an integration to 3rd party s3. (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--integration_export))
- `media_transcriptions` (Block List) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--media_transcriptions))
- `retain_recording` (Boolean) true to retain the recording associated with the conversation.
- `retention_duration` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--retention_duration))

<a id="nestedblock--media_policies--message_policy--actions--assign_calibrations"></a>
### Nested Schema for `media_policies.message_policy.actions.assign_calibrations`

Optional:

- `calibrator_id` (String)
- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `expert_evaluator_id` (String)


<a id="nestedblock--media_policies--message_policy--actions--assign_evaluations"></a>
### Nested Schema for `media_policies.message_policy.actions.assign_evaluations`

Optional:

- `evaluation_form_id` (String)
- `user_id` (String)


<a id="nestedblock--media_policies--message_policy--actions--assign_metered_assignment_by_agent"></a>
### Nested Schema for `media_policies.message_policy.actions.assign_metered_assignment_by_agent`

Optional:

- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `max_number_evaluations` (Number)
- `time_interval` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--assign_metered_assignment_by_agent--time_interval))
- `time_zone` (String)

<a id="nestedblock--media_policies--message_policy--actions--assign_metered_assignment_by_agent--time_interval"></a>
### Nested Schema for `media_policies.message_policy.actions.assign_metered_assignment_by_agent.time_interval`

Optional:

- `days` (Number)
- `months` (Number)
- `weeks` (Number)



<a id="nestedblock--media_policies--message_policy--actions--assign_metered_evaluations"></a>
### Nested Schema for `media_policies.message_policy.actions.assign_metered_evaluations`

Optional:

- `assign_to_active_user` (Boolean)
- `evaluation_form_id` (String)
- `evaluator_ids` (List of String)
- `max_number_evaluations` (Number)
- `time_interval` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--assign_metered_evaluations--time_interval))

<a id="nestedblock--media_policies--message_policy--actions--assign_metered_evaluations--time_interval"></a>
### Nested Schema for `media_policies.message_policy.actions.assign_metered_evaluations.time_interval`

Optional:

- `days` (Number)
- `hours` (Number)



<a id="nestedblock--media_policies--message_policy--actions--assign_surveys"></a>
### Nested Schema for `media_policies.message_policy.actions.assign_surveys`

Required:

- `sending_domain` (String) Validated email domain, required

Optional:

- `flow_id` (String) The UUID reference to the flow associated with this survey.
- `invite_time_interval` (String) An ISO 8601 repeated interval consisting of the number of repetitions, the start datetime, and the interval (e.g. R2/2018-03-01T13:00:00Z/P1M10DT2H30M). Total duration must not exceed 90 days. Defaults to `R1/P0M`.
- `sending_user` (String) User together with sendingDomain used to send email, null to use no-reply
- `survey_form_name` (String) The survey form used for this survey.


<a id="nestedblock--media_policies--message_policy--actions--initiate_screen_recording"></a>
### Nested Schema for `media_policies.message_policy.actions.initiate_screen_recording`

Optional:

- `archive_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--initiate_screen_recording--archive_retention))
- `delete_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--initiate_screen_recording--delete_retention))
- `record_acw` (Boolean)

<a id="nestedblock--media_policies--message_policy--actions--initiate_screen_recording--archive_retention"></a>
### Nested Schema for `media_policies.message_policy.actions.initiate_screen_recording.archive_retention`

Optional:

- `days` (Number)
- `storage_medium` (String)


<a id="nestedblock--media_policies--message_policy--actions--initiate_screen_recording--delete_retention"></a>
### Nested Schema for `media_policies.message_policy.actions.initiate_screen_recording.delete_retention`

Optional:

- `days` (Number)



<a id="nestedblock--media_policies--message_policy--actions--integration_export"></a>
### Nested Schema for `media_policies.message_policy.actions.integration_export`

Optional:

- `integration_id` (String) The aws-s3-recording-bulk-actions-integration that the policy uses for exports.
- `should_export_screen_recordings` (Boolean) True if the policy should export screen recordings in addition to the other conversation media. Defaults to `true`.


<a id="nestedblock--media_policies--message_policy--actions--media_transcriptions"></a>
### Nested Schema for `media_policies.message_policy.actions.media_transcriptions`

Optional:

- `display_name` (String)
- `integration_id` (String)
- `transcription_provider` (String)


<a id="nestedblock--media_policies--message_policy--actions--retention_duration"></a>
### Nested Schema for `media_policies.message_policy.actions.retention_duration`

Optional:

- `archive_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--retention_duration--archive_retention))
- `delete_retention` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--message_policy--actions--retention_duration--delete_retention))

<a id="nestedblock--media_policies--message_policy--actions--retention_duration--archive_retention"></a>
### Nested Schema for `media_policies.message_policy.actions.retention_duration.archive_retention`

Optional:

- `days` (Number)
- `storage_medium` (String)


<a id="nestedblock--media_policies--message_policy--actions--retention_duration--delete_retention"></a>
### Nested Schema for `media_policies.message_policy.actions.retention_duration.delete_retention`

Optional:

- `days` (Number)




<a id="nestedblock--media_policies--message_policy--conditions"></a>
### Nested Schema for `media_policies.message_policy.conditions`

Optional:

- `date_ranges` (List of String)
- `for_queue_ids` (List of String)
- `for_user_ids` (List of String)
- `language_ids` (List of String)
- `time_allowed` (Block List, Max: 1) (see [below for nested schema](#nestedblock--media_policies--message_policy--conditions--time_allowed))
- `wrapup_code_ids` (List of String)

<a id="nestedblock--media_policies--message_policy--conditions--time_allowed"></a>
### Nested Schema for `media_policies.message_policy.conditions.time_allowed`

Optional:

- `empty` (Boolean)
- `time_slots` (Block List) (see [below for nested schema](#nestedblock--media_policies--message_policy--conditions--time_allowed--time_slots))
- `time_zone_id` (String)

<a id="nestedblock--media_policies--message_policy--conditions--time_allowed--time_slots"></a>
### Nested Schema for `media_policies.message_policy.conditions.time_allowed.time_slots`

Optional:

- `day` (Number) Day for this time slot, Monday = 1 ... Sunday = 7
- `start_time` (String) start time in xx:xx:xx.xxx format
- `stop_time` (String) stop time in xx:xx:xx.xxx format






<a id="nestedblock--policy_errors"></a>
### Nested Schema for `policy_errors`

Optional:

- `policy_error_messages` (Block List) (see [below for nested schema](#nestedblock--policy_errors--policy_error_messages))

<a id="nestedblock--policy_errors--policy_error_messages"></a>
### Nested Schema for `policy_errors.policy_error_messages`

Optional:

- `correlation_id` (String)
- `error_code` (String)
- `insert_date` (String) Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z
- `status_code` (Number)
- `user_message` (Map of String)
- `user_params` (Block List) (see [below for nested schema](#nestedblock--policy_errors--policy_error_messages--user_params))
- `user_params_message` (String)

<a id="nestedblock--policy_errors--policy_error_messages--user_params"></a>
### Nested Schema for `policy_errors.policy_error_messages.user_params`

Optional:

- `key` (String)
- `value` (String)

