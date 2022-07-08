resource "genesyscloud_quality_forms_survey" "test-survey-form-1" {
  name      = "terraform-form-surveys-9dc27a6b-0e06-4814-a46e-ba7b56e95d16"
  published = false
  disabled  = false
  language  = "en-US"
  header    = ""
  footer    = ""
  question_groups {
    name       = "Test Question Group 1"
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
    name       = "Test Question Group 2"
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