resource "genesyscloud_user" "ExampleTestOrganization" {
  name  = "Example Test Organization"
  email = "example.test.organization@example.com"
}

resource "genesyscloud_greeting_organization" "Test_Greeting" {
  name       = "Example Test Organization Greeting"
  type       = "NAME"
  owner_type = "ORGANIZATION"
  audio_tts  = "This is a test greeting"
}
