package quality_forms_evaluation

import (
	"fmt"
	"log"
	"strconv"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v178/platformclientv2"
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
	sdkQuestions := make([]platformclientv2.Evaluationquestion, 0)
	for _, question := range questions {
		questionsMap := question.(map[string]interface{})

		text := questionsMap["text"].(string)
		helpText := questionsMap["help_text"].(string)
		naEnabled := questionsMap["na_enabled"].(bool)
		commentsRequired := questionsMap["comments_required"].(bool)
		isKill := questionsMap["is_kill"].(bool)
		isCritical := questionsMap["is_critical"].(bool)

		sdkQuestion := platformclientv2.Evaluationquestion{
			Text:             &text,
			HelpText:         &helpText,
			NaEnabled:        &naEnabled,
			CommentsRequired: &commentsRequired,
			IsKill:           &isKill,
			IsCritical:       &isCritical,
		}

		// Handle type field - use explicit type if provided, otherwise infer from content
		if questionType, ok := questionsMap["type"].(string); ok && questionType != "" {
			sdkQuestion.VarType = &questionType
		} else {
			// Infer type based on which options are provided
			multipleSelectOptionQuestions := questionsMap["multiple_select_option_questions"].([]interface{})
			if len(multipleSelectOptionQuestions) > 0 {
				multipleSelectType := "multipleSelectQuestion"
				sdkQuestion.VarType = &multipleSelectType
			} else {
				multipleChoiceType := "multipleChoiceQuestion"
				sdkQuestion.VarType = &multipleChoiceType
			}
		}

		// Set answer options or multiple select option questions based on type
		multipleSelectOptionQuestions := questionsMap["multiple_select_option_questions"].([]interface{})
		if len(multipleSelectOptionQuestions) > 0 {
			sdkQuestion.MultipleSelectOptionQuestions = buildSdkMultipleSelectOptionQuestions(multipleSelectOptionQuestions)
		}
		answerOptions := questionsMap["answer_options"].([]interface{})
		if len(answerOptions) > 0 {
			sdkQuestion.AnswerOptions = BuildSdkAnswerOptions(answerOptions)
		}

		visibilityCondition := questionsMap["visibility_condition"].([]interface{})
		sdkQuestion.VisibilityCondition = BuildSdkVisibilityCondition(visibilityCondition)

		sdkQuestions = append(sdkQuestions, sdkQuestion)
	}

	return &sdkQuestions
}

func buildSdkMultipleSelectOptionQuestions(optionQuestions []interface{}) *[]platformclientv2.Evaluationquestion {
	sdkQuestions := make([]platformclientv2.Evaluationquestion, 0)
	for _, optionQuestion := range optionQuestions {
		optionMap := optionQuestion.(map[string]interface{})

		text := optionMap["text"].(string)
		helpText := optionMap["help_text"].(string)
		naEnabled := optionMap["na_enabled"].(bool)
		commentsRequired := optionMap["comments_required"].(bool)
		isKill := optionMap["is_kill"].(bool)
		isCritical := optionMap["is_critical"].(bool)

		sdkQuestion := platformclientv2.Evaluationquestion{
			Text:             &text,
			HelpText:         &helpText,
			NaEnabled:        &naEnabled,
			CommentsRequired: &commentsRequired,
			IsKill:           &isKill,
			IsCritical:       &isCritical,
		}

		// Handle type field - use explicit type if provided, otherwise default to multipleChoiceQuestion
		if questionType, ok := optionMap["type"].(string); ok && questionType != "" {
			sdkQuestion.VarType = &questionType
		} else {
			// Default to multipleChoiceQuestion for option questions
			multipleChoiceType := "multipleChoiceQuestion"
			sdkQuestion.VarType = &multipleChoiceType
		}

		// Set answer options
		answerOptions := optionMap["answer_options"].([]interface{})
		if len(answerOptions) > 0 {
			sdkQuestion.AnswerOptions = BuildSdkAnswerOptions(answerOptions)
		}

		visibilityCondition := optionMap["visibility_condition"].([]interface{})
		sdkQuestion.VisibilityCondition = BuildSdkVisibilityCondition(visibilityCondition)

		sdkQuestions = append(sdkQuestions, sdkQuestion)
	}

	return &sdkQuestions
}

func BuildSdkAnswerOptions(answerOptions []interface{}) *[]platformclientv2.Answeroption {
	sdkAnswerOptions := make([]platformclientv2.Answeroption, 0)
	for _, answerOptionsList := range answerOptions {
		answerOptionsMap := answerOptionsList.(map[string]interface{})

		answerValue := answerOptionsMap["value"].(int)

		sdkAnswerOption := platformclientv2.Answeroption{
			Value: &answerValue,
		}

		// Handle text field (optional for built-in types)
		if answerText, ok := answerOptionsMap["text"].(string); ok && answerText != "" {
			sdkAnswerOption.Text = &answerText
		}

		// Handle built_in_type field for multiple select answer options
		if builtInType, ok := answerOptionsMap["built_in_type"].(string); ok && builtInType != "" {
			sdkAnswerOption.BuiltInType = &builtInType
		}

		// Handle assistance_conditions
		if assistanceConditions, ok := answerOptionsMap["assistance_conditions"].([]interface{}); ok && len(assistanceConditions) > 0 {
			sdkAnswerOption.AssistanceConditions = buildSdkAssistanceConditions(assistanceConditions)
		}

		sdkAnswerOptions = append(sdkAnswerOptions, sdkAnswerOption)
	}

	return &sdkAnswerOptions
}

func buildSdkAssistanceConditions(assistanceConditions []interface{}) *[]platformclientv2.Assistancecondition {
	sdkConditions := make([]platformclientv2.Assistancecondition, 0)
	for _, condition := range assistanceConditions {
		conditionMap := condition.(map[string]interface{})

		operator := conditionMap["operator"].(string)
		topicIdsList := conditionMap["topic_ids"].([]interface{})

		topicIds := make([]string, len(topicIdsList))
		for i, id := range topicIdsList {
			topicIds[i] = id.(string)
		}

		sdkCondition := platformclientv2.Assistancecondition{
			Operator: &operator,
			TopicIds: &topicIds,
		}

		sdkConditions = append(sdkConditions, sdkCondition)
	}

	return &sdkConditions
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

		questionText := ""
		if question.Text != nil {
			questionText = *question.Text
			questionMap["text"] = *question.Text
		}

		if question.Id != nil {
			questionMap["id"] = *question.Id
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
		if question.AnswerOptions != nil && len(*question.AnswerOptions) > 0 {
			questionMap["answer_options"] = FlattenAnswerOptions(question.AnswerOptions)
		}
		if question.MultipleSelectOptionQuestions != nil && len(*question.MultipleSelectOptionQuestions) > 0 {
			questionMap["multiple_select_option_questions"] = flattenMultipleSelectOptionQuestions(question.MultipleSelectOptionQuestions)
			log.Printf("flattenQuestions: Question '%s' has %d multiple_select_option_questions", questionText, len(*question.MultipleSelectOptionQuestions))
		}

		questionList = append(questionList, questionMap)
	}
	return questionList
}

func flattenMultipleSelectOptionQuestions(optionQuestions *[]platformclientv2.Evaluationquestion) []interface{} {
	if optionQuestions == nil {
		return nil
	}

	var optionList []interface{}

	for _, option := range *optionQuestions {
		optionMap := make(map[string]interface{})
		if option.Id != nil {
			optionMap["id"] = *option.Id
		}
		if option.Text != nil {
			optionMap["text"] = *option.Text
		}
		if option.HelpText != nil {
			optionMap["help_text"] = *option.HelpText
		}
		if option.VarType != nil {
			optionMap["type"] = *option.VarType
		}
		if option.NaEnabled != nil {
			optionMap["na_enabled"] = *option.NaEnabled
		}
		if option.CommentsRequired != nil {
			optionMap["comments_required"] = *option.CommentsRequired
		}
		if option.IsKill != nil {
			optionMap["is_kill"] = *option.IsKill
		}
		if option.IsCritical != nil {
			optionMap["is_critical"] = *option.IsCritical
		}
		if option.VisibilityCondition != nil {
			optionMap["visibility_condition"] = FlattenVisibilityCondition(option.VisibilityCondition)
		}
		if option.AnswerOptions != nil {
			optionMap["answer_options"] = FlattenAnswerOptions(option.AnswerOptions)
		}

		optionList = append(optionList, optionMap)
	}
	return optionList
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
		if answerOption.BuiltInType != nil {
			answerOptionMap["built_in_type"] = *answerOption.BuiltInType
		}
		if answerOption.AssistanceConditions != nil {
			answerOptionMap["assistance_conditions"] = flattenAssistanceConditions(answerOption.AssistanceConditions)
		}
		answerOptionsList = append(answerOptionsList, answerOptionMap)
	}
	return answerOptionsList
}

func flattenAssistanceConditions(assistanceConditions *[]platformclientv2.Assistancecondition) []interface{} {
	if assistanceConditions == nil {
		return nil
	}

	var conditionList []interface{}
	for _, condition := range *assistanceConditions {
		conditionMap := make(map[string]interface{})
		if condition.Operator != nil {
			conditionMap["operator"] = *condition.Operator
		}
		if condition.TopicIds != nil {
			conditionMap["topic_ids"] = *condition.TopicIds
		}
		conditionList = append(conditionList, conditionMap)
	}
	return conditionList
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
		var optionsString string
		if len(question.MultipleSelectOptionQuestions) > 0 {
			optionsString = GenerateMultipleSelectOptionQuestions(&question.MultipleSelectOptionQuestions)
		} else {
			optionsString = GenerateFormAnswerOptions(&question.AnswerOptions)
		}

		// Generate type field if specified
		typeString := ""
		if question.Type != "" {
			typeString = fmt.Sprintf(`type = "%s"`, question.Type)
		}

		questionString := fmt.Sprintf(`
        questions {
            text              = "%s"
            help_text         = "%s"
            %s
            na_enabled        = %v
            comments_required = %v
            is_kill           = %v
            is_critical       = %v
            %s
            %s
        }
        `, question.Text,
			question.HelpText,
			typeString,
			question.NaEnabled,
			question.CommentsRequired,
			question.IsKill,
			question.IsCritical,
			GenerateFormVisibilityCondition(&question.VisibilityCondition),
			optionsString,
		)

		questionsString += questionString
	}

	return questionsString
}

func GenerateMultipleSelectOptionQuestions(optionQuestions *[]MultipleSelectOptionQuestionStruct) string {
	if optionQuestions == nil {
		return ""
	}

	optionsString := ""

	for _, option := range *optionQuestions {
		// Generate type field if specified
		typeString := ""
		if option.Type != "" {
			typeString = fmt.Sprintf(`type = "%s"`, option.Type)
		}

		optionString := fmt.Sprintf(`
        multiple_select_option_questions {
            text              = "%s"
            help_text         = "%s"
            %s
            na_enabled        = %v
            comments_required = %v
            is_kill           = %v
            is_critical       = %v
            %s
            %s
        }
        `, option.Text,
			option.HelpText,
			typeString,
			option.NaEnabled,
			option.CommentsRequired,
			option.IsKill,
			option.IsCritical,
			GenerateFormVisibilityCondition(&option.VisibilityCondition),
			GenerateFormAnswerOptions(&option.AnswerOptions),
		)

		optionsString += optionString
	}

	return optionsString
}

func GenerateFormAnswerOptions(answerOptions *[]AnswerOptionStruct) string {
	if answerOptions == nil {
		return ""
	}

	answerOptionsString := ""

	for _, answerOption := range *answerOptions {
		var answerOptionString string
		assistanceConditionsString := GenerateAssistanceConditions(&answerOption.AssistanceConditions)

		if answerOption.BuiltInType != "" {
			answerOptionString = fmt.Sprintf(`
        answer_options {
            built_in_type = "%s"
            value         = %v
            %s
        }
        `, answerOption.BuiltInType,
				answerOption.Value,
				assistanceConditionsString,
			)
		} else {
			answerOptionString = fmt.Sprintf(`
        answer_options {
            text  = "%s"
            value = %v
            %s
        }
        `, answerOption.Text,
				answerOption.Value,
				assistanceConditionsString,
			)
		}

		answerOptionsString += answerOptionString
	}

	return answerOptionsString
}

func GenerateAssistanceConditions(assistanceConditions *[]AssistanceConditionStruct) string {
	if assistanceConditions == nil || len(*assistanceConditions) == 0 {
		return ""
	}

	conditionsString := ""

	for _, condition := range *assistanceConditions {
		topicIdsString := ""
		for i, topicId := range condition.TopicIds {
			if i > 0 {
				topicIdsString += ", "
			}
			topicIdsString += strconv.Quote(topicId)
		}

		conditionString := fmt.Sprintf(`
        assistance_conditions {
            operator  = "%s"
            topic_ids = [%s]
        }
        `, condition.Operator,
			topicIdsString,
		)

		conditionsString += conditionString
	}

	return conditionsString
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
