resource "genesyscloud_quality_forms_survey" "example_survey_form" {
  name      = "Example survey form"
  published = true
  disabled  = false
  language  = "en-US"
  header    = "example header"
  footer    = "example footer"
  question_groups {
    name       = "Example Question Group 1"
    na_enabled = false
    questions {
      text                    = "Would you recommend our services?"
      help_text               = ""
      type                    = "npsQuestion"
      na_enabled              = false
      max_response_characters = 100
      explanation_prompt      = "explanation-prompt"
    }
    questions {
      text                    = "Are you satisifed with your experience?"
      help_text               = "Help text here"
      type                    = "freeTextQuestion"
      na_enabled              = true
      max_response_characters = 100
      explanation_prompt      = ""
    }
    questions {
      text       = "Would you recommend our services?"
      help_text  = ""
      type       = "multipleChoiceQuestion"
      na_enabled = false
      answer_options {
        text  = "Yes"
        value = 1
      }
      answer_options {
        text  = "No"
        value = 0
      }
      max_response_characters = 0
    }
  }
  question_groups {
    name       = "Example Question Group 2"
    na_enabled = false
    questions {
      text       = "Did the agent offer to sell product?"
      help_text  = ""
      type       = "multipleChoiceQuestion"
      na_enabled = false
      visibility_condition {
        combining_operation = "AND"
        predicates          = ["/form/questionGroup/0/question/2/answer/1"]
      }
      answer_options {
        text  = "Yes"
        value = 1
      }
      answer_options {
        text  = "No"
        value = 0
      }
      max_response_characters = 0
      explanation_prompt      = ""
    }
    visibility_condition {
      combining_operation = "AND"
      predicates          = ["/form/questionGroup/0/question/2/answer/1"]
    }
  }
}
