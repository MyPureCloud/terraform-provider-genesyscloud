resource "genesyscloud_architect_ivr" "test_ivr" {
  name                  = "Sample IVR"
  description           = "A sample IVR configuration"
  dnis                  = ["+13175550000", "+13175550001"]
  open_hours_flow_id    = "0d9e9ae3-28ed-476b-9364-585f896f4f9d"
  closed_hours_flow_id  = "c49e3d66-4014-49e6-9c6e-13e3e94e5700"
  holiday_hours_flow_id = "736e05d0-6b13-41f8-a2d0-cfba8e826ea3"
  schedule_group_id     = "decb6506-e534-4707-9329-293a3aca38d8"
}
