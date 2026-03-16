resource "genesyscloud_user" "ExampleTestUser" {
  name  = "Example Test User"
  email = "example.test.user@example.com"
}

resource "genesyscloud_greeting_user" "Test_Greeting" {
  name       = "Example Test Greeting"
  type       = "NAME"
  owner_type = "USER"
  user_id    = genesyscloud_user.ExampleTestUser.id
  audio_tts  = "This is a test greeting"
}