resource "genesyscloud_architect_schedules" "open1" {
  name        = "CX as Code Schedule Open 1"
  description = "Sample Schedule by CX as Code"
  start       = "2021-08-04T08:00:00.000000"
  end         = "2021-08-04T17:00:00.000000"
  rrule       = "FREQ=DAILY;INTERVAL=1"
}

resource "genesyscloud_architect_schedules" "open2" {
  name        = "CX as Code Schedule Open 2"
  description = "Sample Schedule by CX as Code"
  start       = "2021-08-04T13:00:00.000000"
  end         = "2021-08-04T22:00:00.000000"
  rrule       = "FREQ=DAILY;INTERVAL=1"
}

resource "genesyscloud_architect_schedules" "closed" {
  name        = "CX as Code Schedule"
  description = "Sample Schedule by CX as Code"
  start       = "2021-08-04T22:00:00.000000"
  end         = "2021-08-05T08:00:00.000000"
  rrule       = "FREQ=DAILY;INTERVAL=1"
}

resource "genesyscloud_architect_schedules" "holiday" {
  name        = "CX as Code Schedule Holiday"
  description = "Sample Schedule by CX as Code"
  start       = "2021-08-04T22:00:00.000000"
  end         = "2021-08-05T08:00:00.000000"
  rrule       = "FREQ=MONTHLY;INTERVAL=1"
}
