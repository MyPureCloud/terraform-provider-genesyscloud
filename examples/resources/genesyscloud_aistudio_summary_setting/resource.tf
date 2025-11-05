resource "genesyscloud_aistudio_summary_setting" "summary" {
  name         = "user10"
  language     = "en-au"
  summary_type = "Concise"
  setting_type = "Basic"
  format       = "BulletPoints"
  mask_p_i_i {
    all = true
  }
  participant_labels {
    internal = "Advisor"
    external = "Member"
  }
  predefined_insights = []
  prompt              = ""
}
