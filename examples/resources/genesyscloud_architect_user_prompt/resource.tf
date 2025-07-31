resource "genesyscloud_architect_user_prompt" "welcome_greeting" {
  name        = "Welcome_Greeting"
  description = "Welcome greeting for all callers"
  resources {
    language   = "en-us"
    text       = "Good day. Thank you for calling."
    tts_string = "Good day. Thank you for calling."
  }
  resources {
    language          = "ja-jp"
    text              = "良い一日。お電話ありがとうございます。"
    filename          = "${local.working_dir.architect_user_prompt}/jp-welcome-greeting.wav"
    file_content_hash = filesha256("${local.working_dir.architect_user_prompt}/jp-welcome-greeting.wav")
  }
}