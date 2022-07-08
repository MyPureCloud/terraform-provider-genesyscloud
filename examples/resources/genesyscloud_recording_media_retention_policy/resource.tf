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