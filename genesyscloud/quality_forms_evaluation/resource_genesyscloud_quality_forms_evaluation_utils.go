package quality_forms_evaluation

import (
	"fmt"
	"strconv"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
)

func buildSdkQuestionGroups(d *schema.ResourceData) *[]platformclientv2.Evaluationquestiongroup {
	questionGroupType := "questionGroup"

	var evalQuestionGroups []platformclientv2.Evaluationquestiongroup
	if questionGroupList, ok := d.Get("question_groups").([]interface{}); ok {
		for _, questionGroup := range questionGroupList {
			questionGroupsMap := questionGroup.(map[string]interface{})

			questionGroupName := questionGroupsMap["name"].(string)
			defaultAnswersToHighest := questionGroupsMap["default_answers_to_highest"].(bool)
			defaultAnswersToNA := questionGroupsMap["default_answers_to_na"].(bool)
			naEnabled := questionGroupsMap["na_enabled"].(bool)
			weight := float32(questionGroupsMap["weight"].(float64))
			manualWeight := questionGroupsMap["manual_weight"].(bool)
			questions := questionGroupsMap["questions"].([]interface{})

			sdkquestionGroup := platformclientv2.Evaluationquestiongroup{
				Name:                    &questionGroupName,
				VarType:                 &questionGroupType,
				DefaultAnswersToHighest: &defaultAnswersToHighest,
				DefaultAnswersToNA:      &defaultAnswersToNA,
				NaEnabled:               &naEnabled,
				Weight:                  &weight,
				ManualWeight:            &manualWeight,
				Questions:               buildSdkQuestions(questions),
			}

			visibilityCondition := questionGroupsMap["visibility_condition"].([]interface{})
			sdkquestionGroup.VisibilityCondition = BuildSdkVisibilityCondition(visibilityCondition)

			evalQuestionGroups = append(evalQuestionGroups, sdkquestionGroup)
		}
	}

	return &evalQuestionGroups
}

func buildSdkQuestions(questions []interface{}) *[]platformclientv2.Evaluationquestion {
	questionType := "multipleChoiceQuestion"

	sdkQuestions := make([]platformclientv2.Evaluationquestion, 0)
	for _, question := range questions {
		questionsMap := question.(map[string]interface{})

		text := questionsMap["text"].(string)
		helpText := questionsMap["help_text"].(string)
		naEnabled := questionsMap["na_enabled"].(bool)
		commentsRequired := questionsMap["comments_required"].(bool)
		answerQuestions := questionsMap["answer_options"].([]interface{})
		isKill := questionsMap["is_kill"].(bool)
		isCritical := questionsMap["is_critical"].(bool)

		sdkQuestion := platformclientv2.Evaluationquestion{
			Text:             &text,
			HelpText:         &helpText,
			VarType:          &questionType,
			NaEnabled:        &naEnabled,
			CommentsRequired: &commentsRequired,
			IsKill:           &isKill,
			IsCritical:       &isCritical,
			AnswerOptions:    BuildSdkAnswerOptions(answerQuestions),
		}

		visibilityCondition := questionsMap["visibility_condition"].([]interface{})
		sdkQuestion.VisibilityCondition = BuildSdkVisibilityCondition(visibilityCondition)

		sdkQuestions = append(sdkQuestions, sdkQuestion)
	}

	return &sdkQuestions
}

func BuildSdkAnswerOptions(answerOptions []interface{}) *[]platformclientv2.Answeroption {
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

func BuildSdkVisibilityCondition(visibilityCondition []interface{}) *platformclientv2.Visibilitycondition {
	if len(visibilityCondition) <= 0 {
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

func flattenQuestionGroups(questionGroups *[]platformclientv2.Evaluationquestiongroup) []interface{} {
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
		if questionGroup.DefaultAnswersToHighest != nil {
			questionGroupMap["default_answers_to_highest"] = *questionGroup.DefaultAnswersToHighest
		}
		if questionGroup.DefaultAnswersToNA != nil {
			questionGroupMap["default_answers_to_na"] = *questionGroup.DefaultAnswersToNA
		}
		if questionGroup.NaEnabled != nil {
			questionGroupMap["na_enabled"] = *questionGroup.NaEnabled
		}
		if questionGroup.Weight != nil {
			questionGroupMap["weight"] = *questionGroup.Weight
		}
		if questionGroup.ManualWeight != nil {
			questionGroupMap["manual_weight"] = *questionGroup.ManualWeight
		}
		if questionGroup.Questions != nil {
			questionGroupMap["questions"] = flattenQuestions(questionGroup.Questions)
		}
		if questionGroup.VisibilityCondition != nil {
			questionGroupMap["visibility_condition"] = FlattenVisibilityCondition(questionGroup.VisibilityCondition)
		}

		questionGroupList = append(questionGroupList, questionGroupMap)
	}
	return questionGroupList
}

func flattenQuestions(questions *[]platformclientv2.Evaluationquestion) []interface{} {
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
		if question.NaEnabled != nil {
			questionMap["na_enabled"] = *question.NaEnabled
		}
		if question.CommentsRequired != nil {
			questionMap["comments_required"] = *question.CommentsRequired
		}
		if question.IsKill != nil {
			questionMap["is_kill"] = *question.IsKill
		}
		if question.IsCritical != nil {
			questionMap["is_critical"] = *question.IsCritical
		}
		if question.VisibilityCondition != nil {
			questionMap["visibility_condition"] = FlattenVisibilityCondition(question.VisibilityCondition)
		}
		if question.AnswerOptions != nil {
			questionMap["answer_options"] = FlattenAnswerOptions(question.AnswerOptions)
		}

		questionList = append(questionList, questionMap)
	}
	return questionList
}

func FlattenAnswerOptions(answerOptions *[]platformclientv2.Answeroption) []interface{} {
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

func FlattenVisibilityCondition(visibilityCondition *platformclientv2.Visibilitycondition) []interface{} {
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

func GenerateEvaluationFormResource(resourceLabel string, evaluationForm *EvaluationFormStruct) string {
	return fmt.Sprintf(`resource "%s" "%s" {
		name = "%s"
		published = %v
		%s
	}
	`, ResourceType, resourceLabel,
		evaluationForm.Name,
		strconv.FormatBool(evaluationForm.Published),
		GenerateEvaluationFormQuestionGroups(&evaluationForm.QuestionGroups),
	)
}

func GenerateEvaluationFormQuestionGroups(questionGroups *[]EvaluationFormQuestionGroupStruct) string {
	if questionGroups == nil {
		return ""
	}

	questionGroupsString := ""

	for _, questionGroup := range *questionGroups {
		questionGroupString := fmt.Sprintf(`
        question_groups {
            name = "%s"
            default_answers_to_highest = %v
            default_answers_to_na  = %v
            na_enabled = %v
            weight = %v
            manual_weight = %v
            %s
            %s
        }
        `, questionGroup.Name,
			questionGroup.DefaultAnswersToHighest,
			questionGroup.DefaultAnswersToNA,
			questionGroup.NaEnabled,
			questionGroup.Weight,
			questionGroup.ManualWeight,
			GenerateEvaluationFormQuestions(&questionGroup.Questions),
			GenerateFormVisibilityCondition(&questionGroup.VisibilityCondition),
		)

		questionGroupsString += questionGroupString
	}

	return questionGroupsString
}

func GenerateEvaluationFormQuestions(questions *[]EvaluationFormQuestionStruct) string {
	if questions == nil {
		return ""
	}

	questionsString := ""

	for _, question := range *questions {
		questionString := fmt.Sprintf(`
        questions {
            text              = "%s"
            help_text         = "%s"
            na_enabled        = %v
            comments_required = %v
            is_kill           = %v
            is_critical       = %v
            %s
            %s
        }
        `, question.Text,
			question.HelpText,
			question.NaEnabled,
			question.CommentsRequired,
			question.IsKill,
			question.IsCritical,
			GenerateFormVisibilityCondition(&question.VisibilityCondition),
			GenerateFormAnswerOptions(&question.AnswerOptions),
		)

		questionsString += questionString
	}

	return questionsString
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

	return answerOptionsString
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

// formIsPublishedRemotely observes the state of the schema resource data to determine if the forum is published remotely
func formIsPublishedRemotely(d *schema.ResourceData) bool {
	return (d.Get("published").(bool) && !d.HasChange("published")) ||
		!d.Get("published").(bool) && d.HasChange("published")
}
