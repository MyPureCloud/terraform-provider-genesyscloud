data "genesyscloud_speechandtextanalytics_dictionaryfeedback" "myfeedback" {
  term                 = "my_dictionaryfeedback"
  dialect              = "en-US"
  transcription_engine = "Genesys"
}
