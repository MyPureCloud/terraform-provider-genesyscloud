resource "genesyscloud_speechandtextanalytics_sentimentfeedback" "positive_feedback" {
  phrase         = "thank you for your help"
  dialect        = "en-US"
  feedback_value = "Positive"
}

resource "genesyscloud_speechandtextanalytics_sentimentfeedback" "negative_feedback" {
  phrase         = "this is unacceptable"
  dialect        = "en-US"
  feedback_value = "Negative"
}
