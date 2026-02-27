resource "genesyscloud_greeting" "test_greeting" {
  name       = "Example Greeting"
  type       = "NAME"
  owner_type = "ORGANIZATION"
  audio_tts  = "This is a test greeting"
}
