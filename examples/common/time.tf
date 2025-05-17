resource "time_offset" "next_week" {
  offset_days = 7
}

resource "time_offset" "ten_days" {
  offset_months = 10
}

resource "time_offset" "tomorrow" {
  offset_days = 1
}

resource "time_static" "now" {}
