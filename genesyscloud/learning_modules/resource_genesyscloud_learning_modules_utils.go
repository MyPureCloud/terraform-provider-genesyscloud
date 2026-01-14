package learning_modules

import (
	"fmt"
	"strconv"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"
)

type InformStepStruct struct {
	Type        string
	Name        string
	Value       string
	SharingUri  string
	ContentType string
	Order       int
	DisplayName string
	Description string
}

type ReviewAssessmentResultsStruct struct {
	ByAssignees bool
	ByViewers   bool
}

type AssessmentFormStruct struct {
	Id             string
	PassPercent    int
	QuestionGroups []QuestionGroupStruct
}

type QuestionGroupStruct struct {
	Id                      string
	Name                    string
	Type                    string
	DefaultAnswersToHighest bool
	DefaultAnswersToNA      bool
	NaEnabled               bool
	Weight                  float32
	ManualWeight            bool
	Questions               []QuestionStruct
	VisibilityCondition     *VisibilityConditionStruct
}

type QuestionStruct struct {
	Id                    string
	Type                  string
	Text                  string
	HelpText              string
	NaEnabled             bool
	CommentsRequired      bool
	IsKill                bool
	IsCritical            bool
	MaxResponseCharacters int
	AnswerOptions         []AnswerOptionStruct
	VisibilityCondition   *VisibilityConditionStruct
}

type AnswerOptionStruct struct {
	Id    string
	Text  string
	Value int
}

type VisibilityConditionStruct struct {
	CombiningOperation string
	Predicates         []string
}

func buildSdkInformSteps(resourceInformStepList []interface{}) *[]platformclientv2.Learningmoduleinformsteprequest {
	informSteps := make([]platformclientv2.Learningmoduleinformsteprequest, 0)

	for _, informStep := range resourceInformStepList {
		informStepMap := informStep.(map[string]interface{})

		var sdkInformStep platformclientv2.Learningmoduleinformsteprequest
		sdkInformStep.VarType = platformclientv2.String(informStepMap["type"].(string))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInformStep.Name, informStepMap, "name")
		sdkInformStep.Value = platformclientv2.String(informStepMap["value"].(string))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInformStep.SharingUri, informStepMap, "sharing_uri")
		sdkInformStep.ContentType = platformclientv2.String(informStepMap["content_type"].(string))
		sdkInformStep.Order = platformclientv2.Int(informStepMap["order"].(int))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInformStep.DisplayName, informStepMap, "display_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInformStep.Description, informStepMap, "description")

		informSteps = append(informSteps, sdkInformStep)
	}

	return &informSteps
}

func buildSdkCoverArt(resourceCoverArtId string) *platformclientv2.Learningmodulecoverartrequest {
	if resourceCoverArtId == "" {
		return nil
	}

	return &platformclientv2.Learningmodulecoverartrequest{
		Id: &resourceCoverArtId,
	}
}

func buildSdkReviewAssessmentResults(resourceReviewAssessmentResultsList []interface{}) *platformclientv2.Reviewassessmentresults {
	if len(resourceReviewAssessmentResultsList) <= 0 {
		return nil
	}

	resourceReviewAssessmentResults, ok := resourceReviewAssessmentResultsList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Reviewassessmentresults{
		ByAssignees: platformclientv2.Bool(resourceReviewAssessmentResults["by_assignees"].(bool)),
		ByViewers:   platformclientv2.Bool(resourceReviewAssessmentResults["by_viewers"].(bool)),
	}
}

func buildSdkAssessmentForm(resourceAssessmentFormList []interface{}) *platformclientv2.Assessmentform {
	if len(resourceAssessmentFormList) <= 0 {
		return nil
	}

	resourceAssessmentForm, ok := resourceAssessmentFormList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var assessmentForm platformclientv2.Assessmentform
	resourcedata.BuildSDKStringValueIfNotNil(&assessmentForm.Id, resourceAssessmentForm, "id")
	assessmentForm.PassPercent = platformclientv2.Int(resourceAssessmentForm["pass_percent"].(int))
	assessmentForm.QuestionGroups = buildSdkQuestionGroups(resourceAssessmentForm["question_groups"].([]interface{}))

	return &assessmentForm
}

func buildSdkQuestionGroups(resourceQuestionGroupsList []interface{}) *[]platformclientv2.Assessmentformquestiongroup {
	assessmentFormQuestionGroups := make([]platformclientv2.Assessmentformquestiongroup, 0)
	for _, questionGroup := range resourceQuestionGroupsList {
		questionGroupsMap := questionGroup.(map[string]interface{})
		var sdkQuestionGroup platformclientv2.Assessmentformquestiongroup

		resourcedata.BuildSDKStringValueIfNotNil(&sdkQuestionGroup.Id, questionGroupsMap, "id")
		sdkQuestionGroup.Name = platformclientv2.String(questionGroupsMap["name"].(string))
		sdkQuestionGroup.VarType = platformclientv2.String(questionGroupsMap["type"].(string))
		sdkQuestionGroup.DefaultAnswersToHighest = platformclientv2.Bool(questionGroupsMap["default_answers_to_highest"].(bool))
		sdkQuestionGroup.DefaultAnswersToNA = platformclientv2.Bool(questionGroupsMap["default_answers_to_na"].(bool))
		sdkQuestionGroup.NaEnabled = platformclientv2.Bool(questionGroupsMap["na_enabled"].(bool))
		sdkQuestionGroup.Weight = platformclientv2.Float32(float32(questionGroupsMap["weight"].(float64)))
		sdkQuestionGroup.ManualWeight = platformclientv2.Bool(questionGroupsMap["manual_weight"].(bool))
		sdkQuestionGroup.Questions = buildSdkQuestions(questionGroupsMap["questions"].([]interface{}))
		sdkQuestionGroup.VisibilityCondition = buildSdkVisibilityCondition(questionGroupsMap["visibility_condition"].([]interface{}))

		assessmentFormQuestionGroups = append(assessmentFormQuestionGroups, sdkQuestionGroup)
	}

	return &assessmentFormQuestionGroups
}

func buildSdkQuestions(questions []interface{}) *[]platformclientv2.Assessmentformquestion {
	assessmentFormQuestions := make([]platformclientv2.Assessmentformquestion, 0)
	for _, question := range questions {
		questionsMap := question.(map[string]interface{})

		var sdkQuestion platformclientv2.Assessmentformquestion

		resourcedata.BuildSDKStringValueIfNotNil(&sdkQuestion.Id, questionsMap, "id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkQuestion.VarType, questionsMap, "type")
		sdkQuestion.Text = platformclientv2.String(questionsMap["text"].(string))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkQuestion.HelpText, questionsMap, "help_text")
		sdkQuestion.NaEnabled = platformclientv2.Bool(questionsMap["na_enabled"].(bool))
		sdkQuestion.CommentsRequired = platformclientv2.Bool(questionsMap["comments_required"].(bool))
		sdkQuestion.IsKill = platformclientv2.Bool(questionsMap["is_kill"].(bool))
		sdkQuestion.IsCritical = platformclientv2.Bool(questionsMap["is_critical"].(bool))
		maxResponseCharacters := questionsMap["max_response_characters"].(int)
		if maxResponseCharacters > 0 {
			sdkQuestion.MaxResponseCharacters = platformclientv2.Int(maxResponseCharacters)
		}
		sdkQuestion.AnswerOptions = buildSdkAnswerOptions(questionsMap["answer_options"].([]interface{}))
		sdkQuestion.VisibilityCondition = buildSdkVisibilityCondition(questionsMap["visibility_condition"].([]interface{}))

		assessmentFormQuestions = append(assessmentFormQuestions, sdkQuestion)
	}

	return &assessmentFormQuestions
}

func buildSdkAnswerOptions(answerOptions []interface{}) *[]platformclientv2.Answeroption {
	sdkAnswerOptions := make([]platformclientv2.Answeroption, 0)
	for _, answerOptionsList := range answerOptions {
		answerOptionsMap := answerOptionsList.(map[string]interface{})

		var sdkAnswerOption platformclientv2.Answeroption

		resourcedata.BuildSDKStringValueIfNotNil(&sdkAnswerOption.Id, answerOptionsMap, "id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkAnswerOption.Text, answerOptionsMap, "text")
		answerValue := answerOptionsMap["value"].(int)
		if answerValue > 0 {
			sdkAnswerOption.Value = platformclientv2.Int(answerValue)
		}

		sdkAnswerOptions = append(sdkAnswerOptions, sdkAnswerOption)
	}

	return &sdkAnswerOptions
}

func buildSdkVisibilityCondition(visibilityCondition []interface{}) *platformclientv2.Visibilitycondition {
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

func flattenInformSteps(informSteps *[]platformclientv2.Learningmoduleinformstep) []interface{} {
	if informSteps == nil {
		return nil
	}

	var informStepList []interface{}
	for _, informStep := range *informSteps {
		informStepMap := make(map[string]interface{})
		if informStep.VarType != nil {
			informStepMap["type"] = *informStep.VarType
		}
		if informStep.Name != nil {
			informStepMap["name"] = *informStep.Name
		}
		if informStep.Value != nil {
			informStepMap["value"] = *informStep.Value
		}
		if informStep.SharingUri != nil {
			informStepMap["sharing_uri"] = *informStep.SharingUri
		}
		if informStep.ContentType != nil {
			informStepMap["content_type"] = *informStep.ContentType
		}
		if informStep.Order != nil {
			informStepMap["order"] = *informStep.Order
		}
		if informStep.DisplayName != nil {
			informStepMap["display_name"] = *informStep.DisplayName
		}
		if informStep.Description != nil {
			informStepMap["description"] = *informStep.Description
		}
		informStepList = append(informStepList, informStepMap)
	}

	return informStepList
}

func flattenCoverArt(coverArt *platformclientv2.Learningmodulecoverartresponse) *string {
	if coverArt == nil {
		return nil
	}

	return coverArt.Id
}

func flattenReviewAssessmentResults(reviewAssessmentResults *platformclientv2.Reviewassessmentresults) []interface{} {
	if reviewAssessmentResults == nil {
		return nil
	}

	reviewAssessmentResultsMap := make(map[string]interface{})
	if reviewAssessmentResults.ByAssignees != nil {
		reviewAssessmentResultsMap["by_assignees"] = *reviewAssessmentResults.ByAssignees
	}
	if reviewAssessmentResults.ByViewers != nil {
		reviewAssessmentResultsMap["by_viewers"] = *reviewAssessmentResults.ByViewers
	}
	return []interface{}{reviewAssessmentResultsMap}
}

func flattenAssessmentForm(assessmentForm *platformclientv2.Assessmentform) []interface{} {
	if assessmentForm == nil {
		return nil
	}

	assessmentFormMap := make(map[string]interface{})
	if assessmentForm.Id != nil {
		assessmentFormMap["id"] = *assessmentForm.Id
	}
	if assessmentForm.PassPercent != nil {
		assessmentFormMap["pass_percent"] = *assessmentForm.PassPercent
	}
	if assessmentForm.QuestionGroups != nil {
		assessmentFormMap["question_groups"] = flattenQuestionGroups(assessmentForm.QuestionGroups)
	}
	return []interface{}{assessmentFormMap}
}

func flattenQuestionGroups(questionGroups *[]platformclientv2.Assessmentformquestiongroup) []interface{} {
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
		if questionGroup.VarType != nil {
			questionGroupMap["type"] = *questionGroup.VarType
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
			questionGroupMap["visibility_condition"] = flattenVisibilityCondition(questionGroup.VisibilityCondition)
		}

		questionGroupList = append(questionGroupList, questionGroupMap)
	}
	return questionGroupList
}

func flattenQuestions(questions *[]platformclientv2.Assessmentformquestion) []interface{} {
	if questions == nil {
		return nil
	}

	var questionList []interface{}

	for _, question := range *questions {
		questionMap := make(map[string]interface{})
		if question.Id != nil {
			questionMap["id"] = *question.Id
		}
		if question.VarType != nil {
			questionMap["type"] = *question.VarType
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
			questionMap["visibility_condition"] = flattenVisibilityCondition(question.VisibilityCondition)
		}
		if question.AnswerOptions != nil {
			questionMap["answer_options"] = flattenAnswerOptions(question.AnswerOptions)
		}
		if question.MaxResponseCharacters != nil {
			questionMap["max_response_characters"] = *question.MaxResponseCharacters
		}

		questionList = append(questionList, questionMap)
	}
	return questionList
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

func GenerateLearningModuleResource(
	resourceLabel string,
	name string,
	description string,
	completionTimeInDays int,
	informSteps []InformStepStruct,
	moduleType string,
	coverArtId string,
	lengthInMinutes int,
	excludedFromCatalog bool,
	externalId string,
	enforceContentOrder bool,
	reviewAssessmentResults *ReviewAssessmentResultsStruct,
	assessmentForm *AssessmentFormStruct,
	isPublished bool,
) string {
	return fmt.Sprintf(`
		resource "genesyscloud_learning_modules" "%s" {
			name = "%s"
			description = "%s"
			completion_time_in_days = %d
			type = "%s"
			cover_art_id = "%s"
			length_in_minutes = %d
			excluded_from_catalog = %v
			external_id = "%s"
			enforce_content_order = %v
			is_published = %v
			%s
			%s
			%s
		}
		`,
		resourceLabel,
		name,
		description,
		completionTimeInDays,
		moduleType,
		coverArtId,
		lengthInMinutes,
		excludedFromCatalog,
		externalId,
		enforceContentOrder,
		isPublished,
		generateInformSteps(informSteps),
		generateReviewAssessmentResults(reviewAssessmentResults),
		generateAssessmentForm(assessmentForm),
	)
}

func generateInformSteps(informSteps []InformStepStruct) string {
	if informSteps == nil {
		return ""
	}

	var informStepsString string
	for _, informStep := range informSteps {
		informStepsString += fmt.Sprintf(`
			inform_steps {
				type = "%s"
				name = "%s"
				value = "%s"
				sharing_uri = "%s"
				content_type = "%s"
				order = %d
				display_name = "%s"
				description = "%s"
			}
			`,
			informStep.Type,
			informStep.Name,
			informStep.Value,
			informStep.SharingUri,
			informStep.ContentType,
			informStep.Order,
			informStep.DisplayName,
			informStep.Description,
		)
	}
	return informStepsString
}

func generateReviewAssessmentResults(reviewAssessmentResults *ReviewAssessmentResultsStruct) string {
	if reviewAssessmentResults == nil {
		return ""
	}

	return fmt.Sprintf(`
		review_assessment_results {
			by_assignees = %v
			by_viewers = %v
		}
		`, reviewAssessmentResults.ByAssignees, reviewAssessmentResults.ByViewers)
}

func generateAssessmentForm(assessmentForm *AssessmentFormStruct) string {
	if assessmentForm == nil {
		return ""
	}

	return fmt.Sprintf(`
		assessment_form {
			id = "%s"
			pass_percent = %v
			%s
		}
		`,
		assessmentForm.Id,
		assessmentForm.PassPercent,
		generateQuestionGroups(assessmentForm.QuestionGroups),
	)
}

func generateQuestionGroups(questionGroups []QuestionGroupStruct) string {
	if questionGroups == nil {
		return ""
	}

	var questionGroupsString string
	for _, questionGroup := range questionGroups {
		questionGroupsString += fmt.Sprintf(`
			question_groups {
				id = "%s"
				name = "%s"
				type = "%s"
				default_answers_to_highest = %v
				default_answers_to_na = %v
				na_enabled = %v
				weight = %v
				manual_weight = %v
				%s
				%s
			}
			`,
			questionGroup.Id,
			questionGroup.Name,
			questionGroup.Type,
			questionGroup.DefaultAnswersToHighest,
			questionGroup.DefaultAnswersToNA,
			questionGroup.NaEnabled,
			questionGroup.Weight,
			questionGroup.ManualWeight,
			generateQuestions(questionGroup.Questions),
			generateVisibilityCondition(questionGroup.VisibilityCondition),
		)
	}
	return questionGroupsString
}

func generateQuestions(questions []QuestionStruct) string {
	if questions == nil {
		return ""
	}

	var questionsString string
	for _, question := range questions {
		questionsString += fmt.Sprintf(`
			questions {
				id = "%s"
				text = "%s"
				help_text = "%s"
				type = "%s"
				na_enabled = %v
				comments_required = %v
				max_response_characters = %v
				is_kill = %v
				is_critical = %v
				%s
				%s
			}
			`,
			question.Id,
			question.Text,
			question.HelpText,
			question.Type,
			question.NaEnabled,
			question.CommentsRequired,
			question.MaxResponseCharacters,
			question.IsKill,
			question.IsCritical,
			generateVisibilityCondition(question.VisibilityCondition),
			generateAnswerOptions(question.AnswerOptions),
		)
	}
	return questionsString
}

func generateVisibilityCondition(visibilityCondition *VisibilityConditionStruct) string {
	if visibilityCondition == nil {
		return ""
	}

	predicateString := ""

	for i, predicate := range visibilityCondition.Predicates {
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
		`, visibilityCondition.CombiningOperation, predicateString)
}

func generateAnswerOptions(answerOptions []AnswerOptionStruct) string {
	if answerOptions == nil {
		return ""
	}

	var answerOptionsString string
	for _, answerOption := range answerOptions {
		answerOptionsString += fmt.Sprintf(`
			answer_options {
				id = "%s"
				text = "%s"
				value = %v
			}
			`,
			answerOption.Id,
			answerOption.Text,
			answerOption.Value,
		)
	}
	return answerOptionsString
}
