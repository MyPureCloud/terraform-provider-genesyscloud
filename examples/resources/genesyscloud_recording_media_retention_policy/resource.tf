resource "genesyscloud_recording_media_retention_policy" "example_media_retention_policy" {
  name        = "example-media-retention-policy"
  order       = 0
  description = "a media retention policy"
  enabled     = true
  media_policies {
    call_policy {
      actions {
        retain_recording = true
        delete_recording = false
        always_delete    = false
        assign_evaluations {
          evaluation_form_id = genesyscloud_quality_forms_evaluation.example_evaluation_form.id
          user_id            = genesyscloud_user.evaluator_user.id
        }
        assign_metered_evaluations {
          evaluator_ids          = [genesyscloud_user.evaluator_user.id]
          max_number_evaluations = 1
          evaluation_form_id     = genesyscloud_quality_forms_evaluation.example_evaluation_form.id
          assign_to_active_user  = true
          time_interval {
            days  = 1
            hours = 1
          }
        }
        assign_metered_assignment_by_agent {
          evaluator_ids          = [genesyscloud_user.evaluator_user.id]
          max_number_evaluations = 1
          evaluation_form_id     = genesyscloud_quality_forms_evaluation.example_evaluation_form.id
          time_interval {
            months = 1
            weeks  = 1
            days   = 1
          }
          time_zone = "EST"
        }
        assign_calibrations {
          calibrator_id       = genesyscloud_user.quality_admin.id
          evaluator_ids       = [genesyscloud_user.evaluator_user.id, genesyscloud_user.quality_admin.id]
          evaluation_form_id  = genesyscloud_quality_forms_evaluation.example_evaluation_form.id
          expert_evaluator_id = genesyscloud_user.quality_admin.id
        }
        assign_surveys {
          sending_domain   = genesyscloud_routing_email_domain.example_domain_com.domain_id
          survey_form_name = genesyscloud_quality_forms_survey.example_survey_form.name
          flow_id          = genesyscloud_flow.workflow_flow.id
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
          integration_id                  = genesyscloud_integration.example_rest_integration.id
          should_export_screen_recordings = true
        }
      }
      conditions {
        for_user_ids    = [genesyscloud_user.evaluator_user.id]
        date_ranges     = ["2022-05-12T04:00:00.000Z/2022-05-13T04:00:00.000Z"]
        for_queue_ids   = [genesyscloud_routing_queue.example_queue.id]
        wrapup_code_ids = [genesyscloud_routing_wrapupcode.win.id]
        language_ids    = [genesyscloud_routing_language.english.id]
        team_ids        = []
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
