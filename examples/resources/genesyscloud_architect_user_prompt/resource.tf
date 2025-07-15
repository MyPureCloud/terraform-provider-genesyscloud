# Basic User Prompt with TTS only
resource "genesyscloud_architect_user_prompt" "basic_prompt" {
  name        = "Basic_TTS_Prompt"
  description = "Basic prompt with text-to-speech only"
  resources {
    language   = "en-us"
    text       = "Welcome to our service. How can I help you today?"
    tts_string = "Welcome to our service. How can I help you today?"
  }
}

# User Prompt with Local Audio File
resource "genesyscloud_architect_user_prompt" "local_audio_prompt" {
  name        = "Local_Audio_Prompt"
  description = "Prompt with local audio file"
  resources {
    language          = "en-us"
    text              = "Welcome message"
    filename          = "./audio/welcome-en.wav"
    file_content_hash = filesha256("./audio/welcome-en.wav")
  }
  resources {
    language          = "es-es"
    text              = "Mensaje de bienvenida"
    filename          = "./audio/welcome-es.wav"
    file_content_hash = filesha256("./audio/welcome-es.wav")
  }
}

# User Prompt with S3 Audio Files
resource "genesyscloud_architect_user_prompt" "s3_audio_prompt" {
  name        = "S3_Audio_Prompt"
  description = "Prompt with audio files from S3"
  resources {
    language          = "en-us"
    text              = "Welcome to our service"
    filename          = "s3://my-audio-bucket/prompts/welcome-en.wav"
    file_content_hash = filesha256("s3://my-audio-bucket/prompts/welcome-en.wav")
  }
  resources {
    language          = "fr-fr"
    text              = "Bienvenue à notre service"
    filename          = "s3://my-audio-bucket/prompts/welcome-fr.wav"
    file_content_hash = filesha256("s3://my-audio-bucket/prompts/welcome-fr.wav")
  }
}

# Mixed Local and S3 Files
resource "genesyscloud_architect_user_prompt" "mixed_prompt" {
  name        = "Mixed_Prompt"
  description = "Prompt with both local and S3 files"
  resources {
    language          = "en-us"
    text              = "Local file greeting"
    filename          = "./local-audio/greeting.wav"
    file_content_hash = filesha256("./local-audio/greeting.wav")
  }
  resources {
    language          = "de-de"
    text              = "S3 file greeting"
    filename          = "s3://audio-bucket/german/greeting.wav"
    file_content_hash = filesha256("s3://audio-bucket/german/greeting.wav")
  }
}
