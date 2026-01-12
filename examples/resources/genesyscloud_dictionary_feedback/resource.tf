resource "genesyscloud_dictionary_feedback" "Genesys" {
  term        = "Genesys"
  dialect     = "en-AU"
  sounds_like = ["Genesis"]
  boost_value = 2.0
  source      = "Manual"
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
