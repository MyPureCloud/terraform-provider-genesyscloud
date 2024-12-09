package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"terraform-provider-genesyscloud/genesyscloud/util/constants"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"terraform-provider-genesyscloud/genesyscloud/util/lists"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v146/platformclientv2"
)

var (
	evaluationFormQuestionGroup = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the question group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of display question in question group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"default_answers_to_highest": {
				Description: "Specifies whether to default answers to highest score.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"default_answers_to_na": {
				Description: "Specifies whether to default answers to not applicable.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"na_enabled": {
				Description: "Specifies whether a not applicable answer is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"weight": {
				Description: "Points per question",
				Type:        schema.TypeFloat,
				Required:    true,
			},
			"manual_weight": {
				Description: "Specifies whether a manual weight is set.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"questions": {
				Description: "Questions inside the group",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem:        evaluationFormQuestion,
			},
			"visibility_condition": {
				Description: "Defines conditions where question would be visible",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        evaluationFormVisibilityCondition,
			},
		},
	}

	evaluationFormQuestion = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the question.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"text": {
				Description: "Individual question",
				Type:        schema.TypeString,
				Required:    true,
			},
			"help_text": {
				Description: "Help text for the question.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"na_enabled": {
				Description: "Specifies whether a not applicable answer is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"comments_required": {
				Description: "Specifies whether comments are required.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"visibility_condition": {
				Description: "Defines conditions where question would be visible",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        evaluationFormVisibilityCondition,
			},
			"answer_options": {
				Description: "Options from which to choose an answer for this question.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    2,
				Elem:        evaluationFormAnswerOptionsResource,
			},
			"is_kill": {
				Description: "True if the question is a fatal question",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"is_critical": {
				Description: "True if the question is a critical question",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}

	evaluationFormVisibilityCondition = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"combining_operation": {
				Description:  "Valid Values: AND, OR",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"AND", "OR"}, false),
			},
			"predicates": {
				Description: "A list of strings, each representing the location in the form of the Answer Option to depend on. In the format of \"/form/questionGroup/{questionGroupIndex}/question/{questionIndex}/answer/{answerIndex}\" or, to assume the current question group, \"../question/{questionIndex}/answer/{answerIndex}\". Note: Indexes are zero-based",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	evaluationFormAnswerOptionsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "The ID for the answer option.",
				Computed:    true,
			},
			"text": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
)

func getAllEvaluationForms(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	qualityAPI := platformclientv2.NewQualityApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		evaluationForms, resp, getErr := qualityAPI.GetQualityFormsEvaluations(pageSize, pageNum, "", "", "", "", "", "")
		if getErr != nil {
			return nil, util.BuildAPIDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to get page of evaluation forms error: %s", getErr), resp)
		}

		if evaluationForms.Entities == nil || len(*evaluationForms.Entities) == 0 {
			break
		}

		for _, evaluationForm := range *evaluationForms.Entities {
			resources[*evaluationForm.Id] = &resourceExporter.ResourceMeta{BlockLabel: *evaluationForm.Name}
		}
	}

	return resources, nil
}

func EvaluationFormExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllEvaluationForms),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
		AllowZeroValues:  []string{"question_groups.questions.answer_options.value", "question_groups.weight"},
		ExcludedAttributes: []string{
			"question_groups.id",
			"question_groups.questions.id",
			"question_groups.questions.answer_options.id",
		},
	}
}

func ResourceEvaluationForm() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Evaluation Forms",
		CreateContext: provider.CreateWithPooledClient(createEvaluationForm),
		ReadContext:   provider.ReadWithPooledClient(readEvaluationForm),
		UpdateContext: provider.UpdateWithPooledClient(updateEvaluationForm),
		DeleteContext: provider.DeleteWithPooledClient(deleteEvaluationForm),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"published": {
				Description: "Specifies if the evaluation form is published.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"question_groups": {
				Description: "A list of question groups.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem:        evaluationFormQuestionGroup,
			},
		},
	}
}

func createEvaluationForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	published := d.Get("published").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	log.Printf("Creating Evaluation Form %s", name)
	form, resp, err := qualityAPI.PostQualityFormsEvaluations(platformclientv2.Evaluationform{
		Name:           &name,
		QuestionGroups: buildSdkQuestionGroups(d),
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to create evaluation form %s error: %s", name, err), resp)
	}

	// Make sure form is properly created
	time.Sleep(2 * time.Second)

	formId := form.Id

	// Publishing
	if published {
		_, resp, err := qualityAPI.PostQualityPublishedformsEvaluations(platformclientv2.Publishform{
			Id:        formId,
			Published: &published,
		})
		if err != nil {
			return util.BuildAPIDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to publish evaluation form %s error: %s", name, err), resp)
		}
	}

	d.SetId(*formId)

	log.Printf("Created evaluation form %s %s", name, *form.Id)
	return readEvaluationForm(ctx, d, meta)
}

func readEvaluationForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceEvaluationForm(), constants.ConsistencyChecks(), "genesyscloud_quality_forms_evaluation")

	log.Printf("Reading evaluation form %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		evaluationForm, resp, getErr := qualityAPI.GetQualityFormsEvaluation(d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to read evaluation form %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to read evaluation form %s | error: %s", d.Id(), getErr), resp))
		}

		// During an export, Retrieve a list of any published versions of the evaluation form
		// If there are published versions, published will be set to true
		if tfexporter_state.IsExporterActive() {
			publishedVersions, resp, err := qualityAPI.GetQualityFormsEvaluationsBulkContexts([]string{*evaluationForm.ContextId})
			if err != nil {
				if util.IsStatus404(resp) {
					return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to retrieve a list of the latest published evaluation form versions"), resp))
				}
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to retrieve a list of the latest published evaluation form versions"), resp))
			}

			if len(publishedVersions) > 0 {
				_ = d.Set("published", true)
			} else {
				_ = d.Set("published", false)
			}
		} else {
			_ = d.Set("published", *evaluationForm.Published)
		}

		if evaluationForm.Name != nil {
			d.Set("name", *evaluationForm.Name)
		}
		if evaluationForm.QuestionGroups != nil {
			d.Set("question_groups", flattenQuestionGroups(evaluationForm.QuestionGroups))
		}

		return cc.CheckState(d)
	})
}

func updateEvaluationForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	published := d.Get("published").(bool)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	// Get the latest unpublished version of the form
	formVersions, resp, err := qualityAPI.GetQualityFormsEvaluationVersions(d.Id(), 25, 1, "desc")
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to get evaluation form versions %s error: %s", name, err), resp)
	}

	unpublishedForm := (*formVersions.Entities)[0]

	log.Printf("Updating Evaluation Form %s", name)
	form, resp, err := qualityAPI.PutQualityFormsEvaluation(*unpublishedForm.Id, platformclientv2.Evaluationform{
		Name:           &name,
		QuestionGroups: buildSdkQuestionGroups(d),
	})
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to update evaluation form %s error: %s", name, err), resp)
	}

	// Set published property on evaluation form update.
	if published {
		_, resp, err := qualityAPI.PostQualityPublishedformsEvaluations(platformclientv2.Publishform{
			Id:        form.Id,
			Published: &published,
		})
		if err != nil {
			return util.BuildAPIDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to publish evaluation form %s error: %s", name, err), resp)
		}
	} else {
		// If published property is reset to false, set the resource Id to the latest unpublished form
		d.SetId(*form.Id)
	}

	log.Printf("Updated evaluation form %s %s", name, *form.Id)
	return readEvaluationForm(ctx, d, meta)
}

func deleteEvaluationForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	// Get the latest unpublished version of the form
	formVersions, resp, err := qualityAPI.GetQualityFormsEvaluationVersions(d.Id(), 25, 1, "desc")
	if err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to get evaluation form versions %s error: %s", name, err), resp)
	}

	latestFormVersion := (*formVersions.Entities)[0]
	d.SetId(*latestFormVersion.Id)

	log.Printf("Deleting evaluation form %s", name)
	if resp, err := qualityAPI.DeleteQualityFormsEvaluation(d.Id()); err != nil {
		return util.BuildAPIDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Failed to delete evaluation form %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := qualityAPI.GetQualityFormsEvaluation(d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				// Evaluation form deleted
				log.Printf("Deleted evaluation form %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Error deleting evaluation form %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_quality_forms_evaluationn", fmt.Sprintf("Evaluation form %s still exists", d.Id()), resp))
	})
}

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
			sdkquestionGroup.VisibilityCondition = buildSdkVisibilityCondition(visibilityCondition)

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
			AnswerOptions:    buildSdkAnswerOptions(answerQuestions),
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
			questionGroupMap["visibility_condition"] = flattenVisibilityCondition(questionGroup.VisibilityCondition)
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
			questionMap["visibility_condition"] = flattenVisibilityCondition(question.VisibilityCondition)
		}
		if question.AnswerOptions != nil {
			questionMap["answer_options"] = flattenAnswerOptions(question.AnswerOptions)
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

func GenerateEvaluationFormResource(resourceLabel string, evaluationForm *EvaluationFormStruct) string {
	return fmt.Sprintf(`resource "genesyscloud_quality_forms_evaluation" "%s" {
		name = "%s"
		published = %v
		%s
	}
	`, resourceLabel,
		evaluationForm.Name,
		evaluationForm.Published,
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

	return fmt.Sprintf(`%s`, answerOptionsString)
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
