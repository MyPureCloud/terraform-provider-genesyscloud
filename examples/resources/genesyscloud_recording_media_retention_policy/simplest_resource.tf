resource "genesyscloud_recording_media_retention_policy" "example_media_retention_policy" {
  name        = "example-media-retention-policy"
  order       = 0
  description = "a media retention policy"
  enabled     = false
  media_policies {
    call_policy {
      actions {
        retain_recording = true
        delete_recording = false
        always_delete    = false

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

      }
      conditions {
        date_ranges = ["2022-05-12T04:00:00.000Z/2022-05-13T04:00:00.000Z"]
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
