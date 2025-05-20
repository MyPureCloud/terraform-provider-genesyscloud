resource "genesyscloud_outbound_callabletimeset" "example_callable_time_set" {
  name = "Example Callable time set"
  callable_times {
    time_zone_id = "Africa/Abidjan"
    time_slots {
      start_time = "07:00:00"
      stop_time  = "18:00:00"
      day        = 3
    }
    time_slots {
      start_time = "09:30:00"
      stop_time  = "22:30:00"
      day        = 5
    }
  }
  callable_times {
    time_zone_id = "Europe/Dublin"
    time_slots {
      start_time = "05:30:30"
      stop_time  = "14:45:00"
      day        = 1
    }
    time_slots {
      start_time = "10:15:45"
      stop_time  = "20:30:00"
      day        = 6
    }
  }
}
