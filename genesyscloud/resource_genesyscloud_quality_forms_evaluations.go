package genesyscloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v48/platformclientv2"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

func resourceEvaluation() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Evaluation Forms",
		CreateContext: createWithPooledClient(createEvaluation),
		ReadContext:   readWithPooledClient(readEvaluation),
		UpdateContext: updateWithPooledClient(updateEvaluation),
		DeleteContext: deleteWithPooledClient(deleteEvaluation),
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
				Description: "Specifies if the evalutaion form is published.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"question_groups": {
				Description: "A list of question groups.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "Name of display question in question group.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"type": {
							Description: "Type of display question. Valid value: questionGroup.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"default_answers_to_highest": {
							Description: "Specifies whether to default answers to highest score.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"default_answers_to_na": {
							Description: "Specifies whether to default answers to not applicable.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"na_enabled": {
							Description: "Specifies whether a not applicable answer is enabled.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"weight": {
							Description: "Points per question",
							Type:        schema.TypeInt,
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
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"text": {
										Description: "Individual question",
										Type:        schema.TypeString,
										Required:    true,
									},
									"help_text": {
										Description: "Help text for the question.",
										Type:        schema.TypeString,
										Required:    false,
									},
									"type": {
										Description: "Type of questions. Valid values: multipleChoiceQuestion, freeTextQuestion, npsQuestion, readOnlyTextBlockQuestion.",
										Type:        schema.TypeString,
										Required:    true,
									},
									"na_enabled": {
										Description: "Specifies whether a not applicable answer is enabled.",
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     true,
									},
									"comments_required": {
										Description: "Specifies whether comments are required.",
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     true,
									},
									"visibility_condition": {
										Description: "",
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										MinItems:    2,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"combining_operation": {
													Description: "Valid Values: AND, OR",
													Type:        schema.TypeString,
													Required:    true,
												},
												"predicates": {
													Description: "A list of strings, each representing the location in the form of the Answer Option to depend on. In the format of \"/form/questionGroup/{questionGroupIndex}/question/{questionIndex}/answer/{answerIndex}\" or, to assume the current question group, \"../question/{questionIndex}/answer/{answerIndex}\". Note: Indexes are zero-based",
													Type:        schema.TypeString,
													Required:    true,
												},
											},
										},
									},
									"answer_options": {
										Description: "Options from which to choose an answer for this question. Only used by Multiple Choice type questions.",
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										MinItems:    2,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"text": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"is_kill": {
										Description: "",
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
									},
									"is_critical": {
										Description: "",
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
									},
								},
							},
						},
						"visibility_condition": {
							Description: "",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							MinItems:    2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"combining_operation": {
										Description: "Valid Values: AND, OR",
										Type:        schema.TypeString,
										Required:    true,
									},
									"predicates": {
										Description: "A list of strings, each representing the location in the form of the Answer Option to depend on. In the format of \"/form/questionGroup/{questionGroupIndex}/question/{questionIndex}/answer/{answerIndex}\" or, to assume the current question group, \"../question/{questionIndex}/answer/{answerIndex}\". Note: Indexes are zero-based",
										Type:        schema.TypeString,
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createEvaluation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	name := d.Get("name").(string)
	questionGroups, err := buildSdkquestionGroups(d)
	if err != nil {
		return diag.FromErr(err)
	}

	sdkConfig := meta.(*providerMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	log.Printf("Creating Evaluation Form %s", name)
	form, _, err := qualityAPI.QualityFormsPostEvaluations(platformclientv2.Evaluationform{
		Name:           &name,
		Published:      &published,
		QuestionGroups: buildSdkquestionGroups(d),
	})
	if err != nil {
		return diag.Errorf("Failed to create evaluation form %s", name)
	}

	d.SetId(*form.Id)

	log.Printf("evaluation form %s %s", name, *form.Id)
	return readFormEvaluation(ctx, d, meta)

}

func readEvaluation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*providerMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	log.Printf("Reading form %s", d.Id())
	return withRetriesForRead(ctx, 30*time.Second, d, func() *resource.RetryError {
		currentEvaluation, resp, getErr := qualityAPI.GetQualityFormsEvaluations(d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read evaluation %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read evaluation %s: %s", d.Id(), getErr))
		}

		d.Set("name", *currentEvaluation.Name)
		d.Set("context_id", *currentEvaluation.contextId)
		d.Set("published", *currentSite.published)
		d.Set("description", nil)

		// Not done

		log.Printf("Read site %s %s", d.Id(), *currentSite.Name)
		return nil
	})

}

func buildSdkquestionGroups(d *schema.ResourceData) *[]platformclientv2.Evaluationquestiongroup {
	var evalQuestionGroups []platformclientv2.Evaluationquestiongroup
	if questionGroups, ok := d.GetOk("question_groups"); ok {
		for _, questionGroupsList := range questionGroups.([]interface{}) {
			questionGroupsMap := questionGroupsList.(map[string]interface{})

			var sdkquestionGroup platformclientv2.Evaluationquestiongroup
			sdkquestionGroup.Name = questionGroupsMap["name"].(string)
			sdkquestionGroup.VarType = questionGroupsMap["type"].(string)
			sdkquestionGroup.DefaultAnswersToHighest = questionGroupsMap["default_answer_to_highest"].(bool)
			sdkquestionGroup.DefaultAnswersToNA = questionGroupsMap["default_answers_to_na"].(bool)
			sdkquestionGroup.NaEnabled = questionGroupsMap["na_enabled"].(bool)
			sdkquestionGroup.Weight = questionGroupsMap["weight"].(int)
			sdkquestionGroup.ManualWeight = questionGroupsMap["manual_weight"].(bool)
			sdkquestionGroup.VisibilityCondition = buildSdkvisibilityCondition(d)

			if questions, ok := questionGroupsMap["questions"].([]interface{}); ok {
				sdkQuestions := make([]platformclientv2.Evaluationquestion, 0)
				for _, questionsList := range questions.([]interface{}) {
					questionsMap := questionsList.(map[string]interface{})

					sdkQuestion := platformclientv2.Evaluationquestion{}
					sdkQuestion.Text = questionsMap["text"].(string)
					sdkQuestion.HelpText = questionsMap["help_text"].(string)
					sdkQuestion.VarType = questionsMap["type"].(string)
					sdkQuestion.NaEnabled = questionsMap["na_enabled"].(bool)
					sdkQuestion.CommentsRequired = questionsMap["comments_required"].(bool)
					sdkQuestion.VisibilityCondition = buildSdkvisibilityCondition(d)

					if answerOptions, ok := questionsMap["answer_options"].([]interface{}); ok {
						sdkAnswerOptions := make([]platformclientv2.Answeroption, 0)
						for _, answerOptionsList := range answerOptions.([]interface{}) {
							answerOptionsMap := answerOptionsList.(map[string]interface{})

							sdkAnswerOption := platformclientv2.Answeroption{}
							sdkAnswerOption.Text = answerOptionsMap["text"].(string)
							sdkAnswerOption.Value = answerOptionsMap["value"].(string)

							sdkAnswerOptions = append(sdkAnswerOptions, sdkAnswerOption)
						}
						sdkQuestion.AnswerOptions = &sdkAnswerOptions
					}

					sdkQuestions = append(sdkQuestions, sdkQuestion)
				}
				sdkquestionGroup.Questions = &sdkQuestions
			}

			evalQuestionGroups = append(evalQuestionGroups, sdkquestionGroup)
		}
	}
	return &evalQuestionGroups
}

func buildSdkvisibilityCondition(d *schema.ResourceData) *platformclientv2.Visibilitycondition {
	if buildSdkvisibilityConditionOptions := d.Get("visibility_condition"); visibilityCondition != nil {
		if visibilityConditionList := visibilityCondition.([]interface{}); len(visibilityConditionList) > 0 {
			visibilityConditionMap := visibilityConditionList[0].(map[string]interface{})

			combiningOperation := visibilityConditionMap["combining_operation"].(string)
			predicates := visibilityConditionMap["predicates"].(string)

			return &platformclientv2.Visibilitycondition{
				CombiningOperation: &combiningOperation,
				Predicates:         &predicates,
			}
		}
	}

	return &platformclientv2.Visibilitycondition{}
}