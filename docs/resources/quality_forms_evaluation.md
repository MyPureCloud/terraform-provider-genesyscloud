---
page_title: "genesyscloud_quality_forms_evaluation Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Evaluation Forms
---
# genesyscloud_quality_forms_evaluation (Resource)

<!-- This document is automatically generated. Do not edit manually. Make changes to the schema, examples, or apis.md files in examples/resources/ and run 'make docs' to regenerate. -->

Genesys Cloud Evaluation Forms

## API Usage

The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/quality/forms/evaluations](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-quality-forms-evaluations)
* [POST /api/v2/quality/forms/evaluations](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-quality-forms-evaluations)
* [GET /api/v2/quality/forms/evaluations/bulk/contexts](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-quality-forms-evaluations-bulk-contexts)
* [DELETE /api/v2/quality/forms/evaluations/{formId}](https://developer.genesys.cloud/devapps/api-explorer#delete-api-v2-quality-forms-evaluations--formId-)
* [GET /api/v2/quality/forms/evaluations/{formId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-quality-forms-evaluations--formId-)
* [PUT /api/v2/quality/forms/evaluations/{formId}](https://developer.genesys.cloud/devapps/api-explorer#put-api-v2-quality-forms-evaluations--formId-)
* [GET /api/v2/quality/forms/evaluations/{formId}/versions](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-quality-forms-evaluations--formId--versions)
* [POST /api/v2/quality/publishedforms/evaluations](https://developer.genesys.cloud/devapps/api-explorer#post-api-v2-quality-publishedforms-evaluations)
* [GET /api/v2/quality/publishedforms/evaluations/{formId}](https://developer.genesys.cloud/devapps/api-explorer#get-api-v2-quality-publishedforms-evaluations--formId-)

## Permissions and Scopes

The following permissions are required to use this resource:

* `quality:evaluationForm:add`
* `quality:evaluationForm:delete`
* `quality:evaluationForm:edit`
* `quality:evaluationForm:view`

The following OAuth scopes are required to use this resource:

* `quality`
* `quality:readonly`


## Example Usage

```terraform
resource "genesyscloud_quality_forms_evaluation" "example_evaluation_form" {
  name      = "Example Evaluation Form"
  published = true
  question_groups {
    name                       = "Question Group 1"
    default_answers_to_highest = true
    default_answers_to_na      = true
    na_enabled                 = true
    weight                     = 1
    manual_weight              = true
    questions {
      text              = "Yes or no question."
      help_text         = "A simple question."
      na_enabled        = true
      comments_required = false
      is_kill           = true
      is_critical       = true
      answer_options {
        text  = "Yes"
        value = 1
      }
      answer_options {
        text  = "No"
        value = 0
      }
    }
    questions {
      text      = "Question with visibility condition"
      help_text = "A lot of fields are optional like this help text."
      answer_options {
        text  = "Yes"
        value = 1
      }
      answer_options {
        text  = "No"
        value = 0
      }
      visibility_condition {
        combining_operation = "AND"
        predicates          = ["/form/questionGroup/0/question/0/answer/0"]
      }
    }
  }
  question_groups {
    name   = "Question Group with Visibility Condition"
    weight = 1.5
    visibility_condition {
      combining_operation = "AND"
      predicates          = ["/form/questionGroup/0/question/0/answer/0"]
    }
    questions {
      text = "Multiple Choice Question."
      answer_options {
        text  = "1"
        value = 1
      }
      answer_options {
        text  = "2"
        value = 2
      }
      answer_options {
        text  = "3"
        value = 3
      }
    }
  }
  question_groups {
    name   = "Question Group with Multiple Select"
    weight = 2.0
    questions {
      text = "Multiple Select Question with options."
      type = "multipleSelectQuestion"
      multiple_select_option_questions {
        text              = "Option A - Basic"
        help_text         = "Help text for Option A"
        na_enabled        = true
        comments_required = true
        is_kill           = false
        is_critical       = true
        answer_options {
          built_in_type = "Selected"
          value         = 1
        }
        answer_options {
          built_in_type = "Unselected"
          value         = 0
        }
      }
      multiple_select_option_questions {
        text              = "Option B - With Visibility"
        help_text         = "This option has a visibility condition"
        na_enabled        = false
        comments_required = false
        is_kill           = true
        is_critical       = false
        visibility_condition {
          combining_operation = "OR"
          predicates          = ["../question/0/answer/0"]
        }
        answer_options {
          built_in_type = "Selected"
          value         = 2
        }
        answer_options {
          built_in_type = "Unselected"
          value         = 0
        }
      }
      multiple_select_option_questions {
        text = "Option C - Minimal"
        answer_options {
          built_in_type = "Selected"
          value         = 1
        }
        answer_options {
          built_in_type = "Unselected"
          value         = 0
        }
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the entity.
- `question_groups` (Block List, Min: 1) A list of question groups. (see [below for nested schema](#nestedblock--question_groups))

### Optional

- `ai_scoring` (Block List, Max: 1) AI scoring settings for the evaluation form. (see [below for nested schema](#nestedblock--ai_scoring))
- `dialect` (String) The language dialect for this evaluation form. Supported dialects: ar, cs, da, de, en-US, es, fi, fr, fr-CA, he, hi, it, ja, ko, nl, no, pl, pt-BR, pt-PT, ru, sv, th, tr, uk, zh-CN, zh-TW.
- `evaluation_settings` (Block List, Max: 1) Settings for evaluations associated with this form. (see [below for nested schema](#nestedblock--evaluation_settings))
- `published` (Boolean) Specifies if the evaluation form is published. **Note:** A form cannot be modified if published is set to true. Defaults to `false`.

### Read-Only

- `context_id` (String) ID of the context of the evaluation form. This provides access to all versions of forms.
- `id` (String) The ID of this resource.
- `modified_date` (String) Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z
- `published_id` (String) The ID of the published evaluation form.

<a id="nestedblock--question_groups"></a>
### Nested Schema for `question_groups`

Required:

- `name` (String) Name of display question in question group.
- `questions` (Block List, Min: 1) Questions inside the group (see [below for nested schema](#nestedblock--question_groups--questions))
- `weight` (Number) Points per question

Optional:

- `default_answers_to` (Block List, Max: 1) Default scoring settings for the questions within this question group. (see [below for nested schema](#nestedblock--question_groups--default_answers_to))
- `default_answers_to_highest` (Boolean) Specifies whether to default answers to highest score. Defaults to `false`.
- `default_answers_to_na` (Boolean) Specifies whether to default answers to not applicable. Defaults to `false`.
- `manual_weight` (Boolean) Specifies whether a manual weight is set. Defaults to `true`.
- `na_enabled` (Boolean) Specifies whether a not applicable answer is enabled. Defaults to `false`.
- `visibility_condition` (Block List, Max: 1) Defines conditions where question would be visible (see [below for nested schema](#nestedblock--question_groups--visibility_condition))

Read-Only:

- `context_id` (String) An identifier for this question group that stays the same across versions of the form.
- `id` (String) ID of the question group.

<a id="nestedblock--question_groups--questions"></a>
### Nested Schema for `question_groups.questions`

Required:

- `text` (String) Individual question

Optional:

- `answer_options` (Block List) Options from which to choose an answer for this question. Required for multipleChoiceQuestion type. (see [below for nested schema](#nestedblock--question_groups--questions--answer_options))
- `automated_scoring_focus` (String) Focus setting for automated scoring. Valid values: FullInteraction, EvaluatedAgent.
- `comments_required` (Boolean) Specifies whether comments are required. Defaults to `false`.
- `default_answer_id` (String) The default selected answer for the question.
- `help_text` (String) Help text for the question.
- `is_critical` (Boolean) True if the question is a critical question Defaults to `false`.
- `is_kill` (Boolean) True if the question is a fatal question Defaults to `false`.
- `multiple_select_option_questions` (Block List) Options for a multiple select question. Each option is itself a question with Selected/Unselected answer options. Required for multipleSelectQuestion type. (see [below for nested schema](#nestedblock--question_groups--questions--multiple_select_option_questions))
- `na_enabled` (Boolean) Specifies whether a not applicable answer is enabled. Defaults to `false`.
- `type` (String) The type of question. Valid values: multipleChoiceQuestion, multipleSelectQuestion, freeTextQuestion, npsQuestion, readOnlyTextBlockQuestion.
- `visibility_condition` (Block List, Max: 1) Defines conditions where question would be visible (see [below for nested schema](#nestedblock--question_groups--questions--visibility_condition))

Read-Only:

- `context_id` (String) An identifier for this question that stays the same across versions of the form.
- `id` (String) ID of the question.

<a id="nestedblock--question_groups--questions--answer_options"></a>
### Nested Schema for `question_groups.questions.answer_options`

Required:

- `value` (Number)

Optional:

- `assistance_conditions` (Block List) List of assistance conditions which are combined together with a logical AND operator. (see [below for nested schema](#nestedblock--question_groups--questions--answer_options--assistance_conditions))
- `built_in_type` (String) The built-in type of this answer option. Only used for Multiple Select answer options. Valid values: Selected, Unselected.
- `text` (String) The text for the answer option. Required for regular answer options.

Read-Only:

- `context_id` (String) An identifier for this answer that stays the same across versions of the form.
- `id` (String) The ID for the answer option.

<a id="nestedblock--question_groups--questions--answer_options--assistance_conditions"></a>
### Nested Schema for `question_groups.questions.answer_options.assistance_conditions`

Required:

- `operator` (String) The operator for the assistance condition. Valid values: EXISTS, NOTEXISTS.
- `topic_ids` (List of String) List of topic IDs which would be combined together using logical OR operator.



<a id="nestedblock--question_groups--questions--multiple_select_option_questions"></a>
### Nested Schema for `question_groups.questions.multiple_select_option_questions`

Required:

- `text` (String) The text/label for the multiple select option.

Optional:

- `answer_options` (Block List) Options from which to choose an answer for this option question. Required for multipleChoiceQuestion type options. (see [below for nested schema](#nestedblock--question_groups--questions--multiple_select_option_questions--answer_options))
- `automated_scoring_focus` (String) Focus setting for automated scoring. Valid values: FullInteraction, EvaluatedAgent.
- `comments_required` (Boolean) Specifies whether comments are required. Defaults to `false`.
- `default_answer_id` (String) The default selected answer for the option question.
- `help_text` (String) Help text for the option.
- `is_critical` (Boolean) True if the option is a critical question Defaults to `false`.
- `is_kill` (Boolean) True if the option is a fatal question Defaults to `false`.
- `na_enabled` (Boolean) Specifies whether a not applicable answer is enabled. Defaults to `false`.
- `type` (String) The type of question. Valid values: multipleChoiceQuestion, freeTextQuestion, npsQuestion, readOnlyTextBlockQuestion.
- `visibility_condition` (Block List, Max: 1) Defines conditions where the option would be visible (see [below for nested schema](#nestedblock--question_groups--questions--multiple_select_option_questions--visibility_condition))

Read-Only:

- `context_id` (String) An identifier for this option that stays the same across versions of the form.
- `id` (String) ID of the question.

<a id="nestedblock--question_groups--questions--multiple_select_option_questions--answer_options"></a>
### Nested Schema for `question_groups.questions.multiple_select_option_questions.answer_options`

Required:

- `value` (Number)

Optional:

- `assistance_conditions` (Block List) List of assistance conditions which are combined together with a logical AND operator. (see [below for nested schema](#nestedblock--question_groups--questions--multiple_select_option_questions--answer_options--assistance_conditions))
- `built_in_type` (String) The built-in type of this answer option. Only used for Multiple Select answer options. Valid values: Selected, Unselected.
- `text` (String) The text for the answer option. Required for regular answer options.

Read-Only:

- `context_id` (String) An identifier for this answer that stays the same across versions of the form.
- `id` (String) The ID for the answer option.

<a id="nestedblock--question_groups--questions--multiple_select_option_questions--answer_options--assistance_conditions"></a>
### Nested Schema for `question_groups.questions.multiple_select_option_questions.answer_options.assistance_conditions`

Required:

- `operator` (String) The operator for the assistance condition. Valid values: EXISTS, NOTEXISTS.
- `topic_ids` (List of String) List of topic IDs which would be combined together using logical OR operator.



<a id="nestedblock--question_groups--questions--multiple_select_option_questions--visibility_condition"></a>
### Nested Schema for `question_groups.questions.multiple_select_option_questions.visibility_condition`

Required:

- `combining_operation` (String) Valid Values: AND, OR
- `predicates` (List of String) A list of strings, each representing the location in the form of the Answer Option to depend on. In the format of "/form/questionGroup/{questionGroupIndex}/question/{questionIndex}/answer/{answerIndex}" or, to assume the current question group, "../question/{questionIndex}/answer/{answerIndex}". Note: Indexes are zero-based



<a id="nestedblock--question_groups--questions--visibility_condition"></a>
### Nested Schema for `question_groups.questions.visibility_condition`

Required:

- `combining_operation` (String) Valid Values: AND, OR
- `predicates` (List of String) A list of strings, each representing the location in the form of the Answer Option to depend on. In the format of "/form/questionGroup/{questionGroupIndex}/question/{questionIndex}/answer/{answerIndex}" or, to assume the current question group, "../question/{questionIndex}/answer/{answerIndex}". Note: Indexes are zero-based



<a id="nestedblock--question_groups--default_answers_to"></a>
### Nested Schema for `question_groups.default_answers_to`

Optional:

- `highest_score` (Boolean) True, when answer should default to highest score. Defaults to `false`.
- `lowest_score` (Boolean) True, when answer should default to lowest score. Defaults to `false`.
- `not_applicable` (Boolean) True, when answer should default to N/A. Defaults to `false`.
- `user_defined` (Boolean) True, when answer should default to user defined answer. Defaults to `false`.


<a id="nestedblock--question_groups--visibility_condition"></a>
### Nested Schema for `question_groups.visibility_condition`

Required:

- `combining_operation` (String) Valid Values: AND, OR
- `predicates` (List of String) A list of strings, each representing the location in the form of the Answer Option to depend on. In the format of "/form/questionGroup/{questionGroupIndex}/question/{questionIndex}/answer/{answerIndex}" or, to assume the current question group, "../question/{questionIndex}/answer/{answerIndex}". Note: Indexes are zero-based



<a id="nestedblock--ai_scoring"></a>
### Nested Schema for `ai_scoring`

Optional:

- `id` (String) The globally unique identifier for the AI scoring settings object.
- `question_group_settings` (Block List) AI scoring settings per question group. (see [below for nested schema](#nestedblock--ai_scoring--question_group_settings))

<a id="nestedblock--ai_scoring--question_group_settings"></a>
### Nested Schema for `ai_scoring.question_group_settings`

Optional:

- `question_group_context_id` (String) The context id of the question group in the form.
- `question_settings` (Block List) AI scoring settings for the questions within this question group. (see [below for nested schema](#nestedblock--ai_scoring--question_group_settings--question_settings))

<a id="nestedblock--ai_scoring--question_group_settings--question_settings"></a>
### Nested Schema for `ai_scoring.question_group_settings.question_settings`

Optional:

- `enabled` (Boolean) True if AI Scoring feature is configured for this question. Defaults to `false`.
- `question_context_id` (String) The context id of the question in the group.




<a id="nestedblock--evaluation_settings"></a>
### Nested Schema for `evaluation_settings`

Optional:

- `disputes_allowed_per_evaluation` (Number) The maximum number of disputes allowed for an evaluation.
- `disputes_assignees` (Block List) A list of assignees responsible for handling each dispute. This list size needs to be equal to disputes_allowed_per_evaluation. (see [below for nested schema](#nestedblock--evaluation_settings--disputes_assignees))
- `disputes_enabled` (Boolean) Whether disputes are allowed for evaluations. Does not apply for calibration evaluations. Defaults to `false`.
- `revisions_enabled` (Boolean) Whether revisions are allowed for evaluations. When enabled, rescoring creates a new version of the evaluation and retracts the existing evaluation version. Does not apply for calibration evaluations. Defaults to `false`.

<a id="nestedblock--evaluation_settings--disputes_assignees"></a>
### Nested Schema for `evaluation_settings.disputes_assignees`

Required:

- `type` (String) The assignee type. Valid values: Original, Individual, None.

Optional:

- `user_id` (String) The ID of the user the dispute should be assigned to. Required when type is Individual.

