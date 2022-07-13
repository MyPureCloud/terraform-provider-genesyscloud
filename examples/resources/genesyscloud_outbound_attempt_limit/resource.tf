resource "genesyscloud_outbound_attempt_limit" "attempt_limit" {
  name                     = "Example Attempt Limit"
  reset_period             = "TODAY"
  time_zone_id             = "Etc/GMT"
  max_attempts_per_contact = 4
  max_attempts_per_number  = 3
  recall_entries {
    no_answer {
      minutes_between_attempts = 6
      nbr_attempts             = 2
    }
    answering_machine {
      minutes_between_attempts = 5
      nbr_attempts             = 1
    }
  }
}