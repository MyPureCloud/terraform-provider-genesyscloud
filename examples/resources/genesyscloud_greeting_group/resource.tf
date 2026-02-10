resource "genesyscloud_group" "ExampleTestGroup" {
  name  = "Example Test Group"
}

resource "genesyscloud_greeting_group" "Test_Greeting" {
  name       = "Example Test Group Greeting"
  type       = "VOICEMAIL"
  owner_type = "GROUP"
  group_id   = genesyscloud_group.ExampleTestGroup.id
  audio_tts  = "This is a test greeting"
}