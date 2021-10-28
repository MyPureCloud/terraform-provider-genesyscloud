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
		Description:   "Genesys Cloud Evaluations",
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

			"context_id": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    false,
			},

			"published": {
				Description: "",
				Type:        schema.TypeBool,
				Required:    true,
			},

			"question_groups": {
				Description: " A list of question groups.",
				Type:        schema.TypeList,
				Optional:    false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "Id of display question in question group. ",
							Type:        schema.TypeString,
							Required:    false,
						},
						"name": {
							Description: "Name of display question in question group. ",
							Type:        schema.TypeString,
							Required:    true,
						},
						"type": {
							Description: "Type of display question. Valid value: questionGroup",
							Type:        schema.TypeString,
							Required:    true,
						},
						"default_answers_to_highest": {
							Description: "",
							Type:        schema.TypeBool,
							Required:    false,
						},
						"default_answers_to_na": {
							Description: "",
							Type:        schema.TypeBool,
							Required:    false,
						},
						"na_enabled": {
							Description: "",
							Type:        schema.TypeBool,
							Required:    false,
						},
						"weight": {
							Description: "Points per question",
							Type:        schema.TypeInt,
							Required:    true,
						},
						"manual_weight": {
							Description: "",
							Type:        schema.TypeBool,
							Required:    true,
						},
						"questions": {
							Description: "Questions inside the group",
							Type:        schema.TypeList,
							Required:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Description: "",
										Type:        schema.TypeString,
										Optional:    false,
									},
									"text": {
										Description: "Individual question",
										Type:        schema.TypeString,
										Optional:    false,
									},
									"help_text": {
										Description: "",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"type": {
										Description: "Type of questions. Valid values: multipleChoiceQuestion, freeTextQuestion, npsQuestion, readOnlyTextBlockQuestion ",
										Type:        schema.TypeString,
										Optional:    false,
									},
									"na_enabled": {
										Description: "Not Applicable Enabled ",
										Type:        schema.TypeBool,
										Optional:    false,
									},
									"comments_required": {
										Description: "",
										Type:        schema.TypeBool,
										Optional:    false,
									},
									"visibility_condition": {
										Description: "",
										Type:        schema.TypeBool,
										Optional:    false,
										MinItems:    2,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"combining_operation": {
													Description: "Valid Values: AND, OR",
													Type:     schema.TypeString,
													Optional: false,
												},
												"predicates": {
													Description: "A list of strings, each representing the location in the form of the Answer Option to depend on. In the format of "/form/questionGroup/{questionGroupIndex}/question/{questionIndex}/answer/{answerIndex}" or, to assume the current question group, "../question/{questionIndex}/answer/{answerIndex}". Note: Indexes are zero-based",
													Type:     schema.TypeString,
													Optional: false,
												},
											},
										},
									},
									"answer_options": {
										Description: "Choices for questions",
										Type:        schema.TypeBool,
										Optional:    false,
										MinItems:    2,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"text": {
													Type:     schema.TypeString,
													Optional: false,
												},
												"value": {
													Type:     schema.TypeString,
													Optional: false,
												},
											},
										},
									},
									"is_kill": {
										Description: "",
										Type:        schema.TypeBool,
										Optional:    false,
									},
									"is_critical": {
										Description: "",
										Type:        schema.TypeBool,
										Optional:    false,
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
		Name:         	 &name,
		ContextId:       &contextId,
		Published:       &published,
		QuestionGroups:  &buildSdkquestionGroups(d),
		isKill:       	 &isKill,
		isCritical:      &isCritical,

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

func buildSdkquestionGroups(d *schema.ResourceData) (*platformclientv2.QualityApi, error) {
	if questionGroups := d.Get("question_groups"); questionGroups != nil {
		if questionGroupsList := questionGroups.([]interface{}); len(questionGroupsList) > 0 {
			questionGroupsMap := questionGroupsList[0].(map[string]interface{})

			id := questionGroupsMap["id"].(string)
			name := questionGroupsMap["name"].(string)
			type := questionGroupsMap["type"].(string)
			defaultAnswersToHighest := questionGroupsMap["default_answer_to_highest"].(bool)
			defaultAnswersToNA := questionGroupsMap["default_answers_to_na"].(bool)
			naEnabled := questionGroupsMap["na_enabled"].(bool)
			weight := questionGroupsMap["weight"].(string)
			manualWeight := questionGroupsMap["manual_weight"].(string)
			questions, err := buildSdkquestions(d)
	if err != nil {
		return diag.FromErr(err)
	}

			return &platformclientv2.QualityApi{
				name:  &name,
				type: &type,
				weight: &weight,
			}, nil
		}
	}

	return nil, nil
}

func buildSdkquestions(d *schema.ResourceData) (*platformclientv2.QualityApi, error) {
	if buildSdkquestions := d.Get("questions"); questions != nil {
		if questionsList := questions.([]interface{}); len(questionsList) > 0 {
			questionsMap := questionsList[0].(map[string]interface{})
			id := questionsMap["id"].(string)
			text := questionsMap["text"].(string)
			helpText := questionsMap["help_text"].(string)
			type := questionsMap["type"].(string)
			naEnabled := questionsMap["type"].(bool)
			commentsRequired := questionsMap["comments_required"].(bool)
			visibilitycondition, err := buildSdkvisibilityCondition(d)
			answerOptions, err := buildSdkanswerOptions(d)
	if err != nil {
		return diag.FromErr(err)
	}

			return &platformclientv2.QualityApi{
				name:  &name,
				type: &type,
				weight: &weight,
			}, nil
		}
	}

	return nil, nil
}

func buildSdkvisibilityCondition(d *schema.ResourceData) (*platformclientv2.QualityApi, error) {
	if buildSdkvisibilityConditionOptions := d.Get("visibility_condition"); visibilityCondition != nil {
		if visibilityConditionList := visibilityCondition.([]interface{}); len(visibilityConditionList) > 0 {
			visibilityConditionMap := visibilityConditionList[0].(map[string]interface{})

			combiningOperation := visibilityConditionMap["combining_operation"].(string)
			predicates := visibilityConditionMap["predicates"].(string)

			return &platformclientv2.QualityApi{
				text:  &text,
				value: &value,
			}, nil
		}
	}

	return nil, nil
}

func buildSdkanswerOptions(d *schema.ResourceData) (*platformclientv2.QualityApi, error) {
	if buildSdkanswerOptions := d.Get("answer_options"); answerOptions != nil {
		if answerOptionsList := answerOptions.([]interface{}); len(answerOptionsList) > 0 {
			answerOptionsMap := answerOptionsList[0].(map[string]interface{})

			text := answerOptionsMap["text"].(string)
			value := answerOptionsMap["value"].(string)


			return &platformclientv2.QualityApi{
				text:  &text,
				value: &value,
			}, nil
		}
	}

	return nil, nil
}