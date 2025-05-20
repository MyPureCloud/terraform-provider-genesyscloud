resource "genesyscloud_organization_presence_definition" "Away_From_Keyboard" {
  system_presence = "Away"
  language_labels = {
    en_US = "From Keyboard"
    es    = "del teclado"
  }
}
