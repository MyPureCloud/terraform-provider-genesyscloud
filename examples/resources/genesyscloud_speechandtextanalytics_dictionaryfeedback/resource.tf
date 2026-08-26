resource "genesyscloud_speechandtextanalytics_dictionaryfeedback" "Genesys" {
  term                 = "Genesys"
  dialect              = "en-AU"
  transcription_engine = "Genesys"
  sounds_like          = ["Genesis"]
  boost_value          = 2.0
  source               = "Manual"
  example_phrases {
    phrase = "Welcome to Genesys"
  }
  example_phrases {
    phrase = "Thanks for calling Genesys"
  }
  example_phrases {
    phrase = "Goodbye from Genesys"
  }
}

resource "genesyscloud_speechandtextanalytics_dictionaryfeedback" "Extended" {
  term                 = "covid"
  dialect              = "en-US"
  transcription_engine = "GenesysExtended"
  display_as           = "COVID-19"
}
