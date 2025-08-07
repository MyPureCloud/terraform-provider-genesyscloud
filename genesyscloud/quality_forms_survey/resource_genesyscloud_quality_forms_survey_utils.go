package quality_forms_survey

import (
	"fmt"
	"strconv"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func buildSurveyQuestionGroups(d *schema.ResourceData) (*[]platformclientv2.Surveyquestiongroup, diag.Diagnostics) {
	questionGroupType := "questionGroup"

	var surveyQuestionGroups []platformclientv2.Surveyquestiongroup
	if questionGroups, ok := d.GetOk("question_groups"); ok {
		questionGroupList := questionGroups.([]interface{})
		for _, questionGroup := range questionGroupList {
			questionGroupsMap := questionGroup.(map[string]interface{})

			questionGroupName := questionGroupsMap["name"].(string)
			naEnabled := questionGroupsMap["na_enabled"].(bool)
			questions := questionGroupsMap["questions"].([]interface{})

			sdkquestionGroup := platformclientv2.Surveyquestiongroup{
				Name:      &questionGroupName,
				VarType:   &questionGroupType,
				NaEnabled: &naEnabled,
				Questions: buildSurveyQuestions(questions),
			}

			visibilityCondition := questionGroupsMap["visibility_condition"].([]interface{})
			sdkquestionGroup.VisibilityCondition = buildSdkVisibilityCondition(visibilityCondition)

			surveyQuestionGroups = append(surveyQuestionGroups, sdkquestionGroup)
		}
	}

	return &surveyQuestionGroups, nil
}

func buildSurveyQuestions(questions []interface{}) *[]platformclientv2.Surveyquestion {
	sdkQuestions := make([]platformclientv2.Surveyquestion, 0)
	for _, question := range questions {
		questionsMap := question.(map[string]interface{})
		text := questionsMap["text"].(string)
		helpText := questionsMap["help_text"].(string)
		questionType := questionsMap["type"].(string)
		naEnabled := questionsMap["na_enabled"].(bool)
		answerQuestions := questionsMap["answer_options"].([]interface{})
		maxResponseCharacters := questionsMap["max_response_characters"].(int)
		sdkAnswerOptions := buildSdkAnswerOptions(answerQuestions)

		sdkQuestion := platformclientv2.Surveyquestion{
			Text:                  &text,
			HelpText:              &helpText,
			VarType:               &questionType,
			NaEnabled:             &naEnabled,
			AnswerOptions:         sdkAnswerOptions,
			MaxResponseCharacters: &maxResponseCharacters,
		}

		explanationPrompt := questionsMap["explanation_prompt"].(string)
		if explanationPrompt != "" {
			sdkQuestion.ExplanationPrompt = &explanationPrompt
		}

		visibilityCondition := questionsMap["visibility_condition"].([]interface{})
		sdkQuestion.VisibilityCondition = buildSdkVisibilityCondition(visibilityCondition)

		sdkQuestions = append(sdkQuestions, sdkQuestion)
	}

	return &sdkQuestions
}

func buildSdkAnswerOptions(answerOptions []interface{}) *[]platformclientv2.Answeroption {
	sdkAnswerOptions := make([]platformclientv2.Answeroption, 0)
	for _, answerOptionsList := range answerOptions {
		answerOptionsMap := answerOptionsList.(map[string]interface{})

		answerText := answerOptionsMap["text"].(string)
		answerValue := answerOptionsMap["value"].(int)

		sdkAnswerOption := platformclientv2.Answeroption{
			Text:  &answerText,
			Value: &answerValue,
		}

		sdkAnswerOptions = append(sdkAnswerOptions, sdkAnswerOption)
	}

	return &sdkAnswerOptions
}

func buildSdkVisibilityCondition(visibilityCondition []interface{}) *platformclientv2.Visibilitycondition {
	if visibilityCondition == nil || len(visibilityCondition) <= 0 {
		return nil
	}

	visibilityConditionMap, ok := visibilityCondition[0].(map[string]interface{})
	if !ok {
		return nil
	}

	combiningOperation := visibilityConditionMap["combining_operation"].(string)
	predicates := visibilityConditionMap["predicates"].([]interface{})

	return &platformclientv2.Visibilitycondition{
		CombiningOperation: &combiningOperation,
		Predicates:         &predicates,
	}
}

func flattenSurveyQuestionGroups(questionGroups *[]platformclientv2.Surveyquestiongroup) []interface{} {
	if questionGroups == nil {
		return nil
	}

	var questionGroupList []interface{}

	for _, questionGroup := range *questionGroups {
		questionGroupMap := make(map[string]interface{})
		if questionGroup.Id != nil {
			questionGroupMap["id"] = *questionGroup.Id
		}
		if questionGroup.Name != nil {
			questionGroupMap["name"] = *questionGroup.Name
		}
		if questionGroup.NaEnabled != nil {
			questionGroupMap["na_enabled"] = *questionGroup.NaEnabled
		}
		if questionGroup.Questions != nil {
			questionGroupMap["questions"] = flattenSurveyQuestions(questionGroup.Questions)
		}
		if questionGroup.VisibilityCondition != nil {
			questionGroupMap["visibility_condition"] = flattenVisibilityCondition(questionGroup.VisibilityCondition)
		}

		questionGroupList = append(questionGroupList, questionGroupMap)
	}
	return questionGroupList
}

func flattenSurveyQuestions(questions *[]platformclientv2.Surveyquestion) []interface{} {
	if questions == nil {
		return nil
	}

	var questionList []interface{}

	for _, question := range *questions {
		questionMap := make(map[string]interface{})
		if question.Id != nil {
			questionMap["id"] = *question.Id
		}
		if question.Text != nil {
			questionMap["text"] = *question.Text
		}
		if question.HelpText != nil {
			questionMap["help_text"] = *question.HelpText
		}
		if question.VarType != nil {
			questionMap["type"] = *question.VarType
		}
		if question.NaEnabled != nil {
			questionMap["na_enabled"] = *question.NaEnabled
		}
		if question.VisibilityCondition != nil {
			questionMap["visibility_condition"] = flattenVisibilityCondition(question.VisibilityCondition)
		}
		if question.AnswerOptions != nil {
			questionMap["answer_options"] = flattenAnswerOptions(question.AnswerOptions)
		}
		if question.MaxResponseCharacters != nil {
			questionMap["max_response_characters"] = *question.MaxResponseCharacters
		}
		if question.ExplanationPrompt != nil {
			questionMap["explanation_prompt"] = *question.ExplanationPrompt
		}

		questionList = append(questionList, questionMap)
	}
	return questionList
}

func flattenVisibilityCondition(visibilityCondition *platformclientv2.Visibilitycondition) []interface{} {
	if visibilityCondition == nil {
		return nil
	}

	visibilityConditionMap := make(map[string]interface{})
	if visibilityCondition.CombiningOperation != nil {
		visibilityConditionMap["combining_operation"] = *visibilityCondition.CombiningOperation
	}
	if visibilityCondition.Predicates != nil {
		visibilityConditionMap["predicates"] = lists.InterfaceListToStrings(*visibilityCondition.Predicates)
	}

	return []interface{}{visibilityConditionMap}
}

func flattenAnswerOptions(answerOptions *[]platformclientv2.Answeroption) []interface{} {
	if answerOptions == nil {
		return nil
	}

	var answerOptionsList []interface{}

	for _, answerOption := range *answerOptions {
		answerOptionMap := make(map[string]interface{})
		if answerOption.Id != nil {
			answerOptionMap["id"] = *answerOption.Id
		}
		if answerOption.Text != nil {
			answerOptionMap["text"] = *answerOption.Text
		}
		if answerOption.Value != nil {
			answerOptionMap["value"] = *answerOption.Value
		}
		answerOptionsList = append(answerOptionsList, answerOptionMap)
	}
	return answerOptionsList
}

func GenerateSurveyFormResource(resourceLabel string, surveyForm *SurveyFormStruct) string {
	form := fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		published = %v
		disabled = %v
        language = "%s"
        header = "%s"
        footer = "%s"
		%s
        %s
	}
	`, ResourceType, resourceLabel,
		surveyForm.Name,
		surveyForm.Published,
		surveyForm.Disabled,
		surveyForm.Language,
		surveyForm.Header,
		surveyForm.Footer,
		generateSurveyFormQuestionGroups(&surveyForm.QuestionGroups),
		generateLifeCycle(),
	)

	return form
}

func generateLifeCycle() string {
	return `
	lifecycle {
		ignore_changes = [
			question_groups[0].questions[0].type,
			question_groups[0].questions[1].type,
			question_groups[0].questions[2].type,
			question_groups[1].questions[0].type,
			question_groups[1].questions[1].type,
			question_groups[1].questions[2].type,
			question_groups[2].questions[0].type,
			question_groups[2].questions[1].type,
			question_groups[2].questions[2].type,
		]
	}
	`
}

func generateSurveyFormQuestions(questions *[]SurveyFormQuestionStruct) string {
	if questions == nil {
		return ""
	}

	questionsString := ""

	for _, question := range *questions {
		questionString := fmt.Sprintf(`
        questions {
            text = "%s"
            help_text = "%s"
            type = "%s"
            na_enabled = %v
            %s
            %s
            max_response_characters = %v
            explanation_prompt = "%s"
        }
        `, question.Text,
			question.HelpText,
			question.VarType,
			question.NaEnabled,
			GenerateFormVisibilityCondition(&question.VisibilityCondition),
			GenerateFormAnswerOptions(&question.AnswerOptions),
			question.MaxResponseCharacters,
			question.ExplanationPrompt,
		)

		questionsString += questionString
	}

	return questionsString
}

func generateSurveyFormQuestionGroups(questionGroups *[]SurveyFormQuestionGroupStruct) string {
	if questionGroups == nil {
		return ""
	}

	questionGroupsString := ""

	for _, questionGroup := range *questionGroups {
		questionGroupString := fmt.Sprintf(`
        question_groups {
            name = "%s"
            na_enabled = %v
            %s
            %s
        }
        `, questionGroup.Name,
			questionGroup.NaEnabled,
			generateSurveyFormQuestions(&questionGroup.Questions),
			GenerateFormVisibilityCondition(&questionGroup.VisibilityCondition),
		)

		questionGroupsString += questionGroupString
	}

	return questionGroupsString
}

func GenerateFormVisibilityCondition(condition *VisibilityConditionStruct) string {
	if condition == nil || len(condition.CombiningOperation) == 0 {
		return ""
	}

	predicateString := ""

	for i, predicate := range condition.Predicates {
		if i > 0 {
			predicateString += ", "
		}

		predicateString += strconv.Quote(predicate)
	}

	return fmt.Sprintf(`
	visibility_condition {
        combining_operation = "%s"
        predicates = [%s]
    }
	`, condition.CombiningOperation,
		predicateString,
	)
}

func GenerateFormAnswerOptions(answerOptions *[]AnswerOptionStruct) string {
	if answerOptions == nil {
		return ""
	}

	answerOptionsString := ""

	for _, answerOption := range *answerOptions {
		answerOptionString := fmt.Sprintf(`
        answer_options {
            text  = "%s"
            value = %v
        }
        `, answerOption.Text,
			answerOption.Value,
		)

		answerOptionsString += answerOptionString
	}

	return fmt.Sprintf(`%s`, answerOptionsString)
}
