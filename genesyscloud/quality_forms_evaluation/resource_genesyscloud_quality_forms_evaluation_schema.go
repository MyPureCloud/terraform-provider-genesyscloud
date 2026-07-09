package quality_forms_evaluation

// @team: PureCloud QM
// @chat: #genesys-cloud-qm-dev
// @pm: Jose Ruiz
// @jira: QM
// @description: Quality Management service for agent performance evaluation and customer feedback. Manages evaluation forms for supervisor assessments and survey forms for customer satisfaction measurement.

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
	sttTopic "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/speechandtextanalytics_topic"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/user"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ResourceType = "genesyscloud_quality_forms_evaluation"

var (
	evaluationFormQuestionGroup = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the question group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"context_id": {
				Description: "An identifier for this question group that stays the same across versions of the form.",
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
			"default_answers_to": {
				Description: "Default scoring settings for the questions within this question group.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        evaluationFormDefaultAnswersTo,
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
			"context_id": {
				Description: "An identifier for this question that stays the same across versions of the form.",
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
			"type": {
				Description:  "The type of question. Valid values: multipleChoiceQuestion, multipleSelectQuestion, freeTextQuestion, npsQuestion, readOnlyTextBlockQuestion.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"multipleChoiceQuestion", "multipleSelectQuestion", "freeTextQuestion", "npsQuestion", "readOnlyTextBlockQuestion"}, false),
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
				Description: "Options from which to choose an answer for this question. Required for multipleChoiceQuestion type.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        evaluationFormAnswerOptionsResource,
			},
			"multiple_select_option_questions": {
				Description: "Options for a multiple select question. Each option is itself a question with Selected/Unselected answer options. Required for multipleSelectQuestion type.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        evaluationFormMultipleSelectOptionQuestion,
			},
			"default_answer_id": {
				Description: "The default selected answer for the question.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"automated_scoring_focus": {
				Description:  "Focus setting for automated scoring. Valid values: FullInteraction, EvaluatedAgent.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"FullInteraction", "EvaluatedAgent"}, false),
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

	evaluationFormMultipleSelectOptionQuestion = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the question.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"context_id": {
				Description: "An identifier for this option that stays the same across versions of the form.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"text": {
				Description: "The text/label for the multiple select option.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"help_text": {
				Description: "Help text for the option.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description:  "The type of question. Valid values: multipleChoiceQuestion, freeTextQuestion, npsQuestion, readOnlyTextBlockQuestion.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"multipleChoiceQuestion", "freeTextQuestion", "npsQuestion", "readOnlyTextBlockQuestion"}, false),
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
				Description: "Defines conditions where the option would be visible",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        evaluationFormVisibilityCondition,
			},
			"answer_options": {
				Description: "Options from which to choose an answer for this option question. Required for multipleChoiceQuestion type options.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        evaluationFormAnswerOptionsResource,
			},
			"default_answer_id": {
				Description: "The default selected answer for the option question.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"automated_scoring_focus": {
				Description:  "Focus setting for automated scoring. Valid values: FullInteraction, EvaluatedAgent.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"FullInteraction", "EvaluatedAgent"}, false),
			},
			"is_kill": {
				Description: "True if the option is a fatal question",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"is_critical": {
				Description: "True if the option is a critical question",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
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
			"context_id": {
				Type:        schema.TypeString,
				Description: "An identifier for this answer that stays the same across versions of the form.",
				Computed:    true,
			},
			"text": {
				Description: "The text for the answer option. Required for regular answer options.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"value": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"built_in_type": {
				Description: "The built-in type of this answer option. Only used for Multiple Select answer options. Valid values: Selected, Unselected.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"assistance_conditions": {
				Description: "List of assistance conditions which are combined together with a logical AND operator.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        evaluationFormAssistanceCondition,
			},
		},
	}

	evaluationFormAssistanceCondition = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"operator": {
				Description:  "The operator for the assistance condition. Valid values: EXISTS, NOTEXISTS.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"EXISTS", "NOTEXISTS"}, false),
			},
			"topic_ids": {
				Description: "List of topic IDs which would be combined together using logical OR operator.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}

	evaluationFormDefaultAnswersTo = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"highest_score": {
				Description: "True, when answer should default to highest score.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"not_applicable": {
				Description: "True, when answer should default to N/A.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"lowest_score": {
				Description: "True, when answer should default to lowest score.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"user_defined": {
				Description: "True, when answer should default to user defined answer.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}

	evaluationFormDisputesAssignee = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "The ID of the user the dispute should be assigned to. Required when type is Individual.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description:  "The assignee type. Valid values: Original, Individual, None.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"Original", "Individual", "None"}, false),
			},
		},
	}

	evaluationFormEvaluationSettings = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"revisions_enabled": {
				Description: "Whether revisions are allowed for evaluations. When enabled, rescoring creates a new version of the evaluation and retracts the existing evaluation version. Does not apply for calibration evaluations.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"disputes_enabled": {
				Description: "Whether disputes are allowed for evaluations. Does not apply for calibration evaluations.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"disputes_allowed_per_evaluation": {
				Description: "The maximum number of disputes allowed for an evaluation.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"disputes_assignees": {
				Description: "A list of assignees responsible for handling each dispute. This list size needs to be equal to disputes_allowed_per_evaluation.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        evaluationFormDisputesAssignee,
			},
		},
	}

	evaluationFormAiScoringQuestionSetting = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"question_context_id": {
				Description: "The context id of the question in the group.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"enabled": {
				Description: "True if AI Scoring feature is configured for this question.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}

	evaluationFormAiScoringQuestionGroupSetting = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"question_group_context_id": {
				Description: "The context id of the question group in the form.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"question_settings": {
				Description: "AI scoring settings for the questions within this question group.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        evaluationFormAiScoringQuestionSetting,
			},
		},
	}

	evaluationFormAiScoring = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The globally unique identifier for the AI scoring settings object.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"question_group_settings": {
				Description: "AI scoring settings per question group.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        evaluationFormAiScoringQuestionGroupSetting,
			},
		},
	}
)

type EvaluationFormQuestionGroupStruct struct {
	Name                    string
	DefaultAnswersToHighest bool
	DefaultAnswersToNA      bool
	NaEnabled               bool
	Weight                  float32
	ManualWeight            bool
	Questions               []EvaluationFormQuestionStruct
	VisibilityCondition     VisibilityConditionStruct
	DefaultAnswersTo        *DefaultAnswersToStruct
}

type EvaluationFormStruct struct {
	Name               string
	Published          bool
	Dialect            string
	QuestionGroups     []EvaluationFormQuestionGroupStruct
	EvaluationSettings *EvaluationSettingsStruct
	DependsOn          []string
}

type EvaluationFormQuestionStruct struct {
	Text                          string
	HelpText                      string
	Type                          string
	NaEnabled                     bool
	CommentsRequired              bool
	IsKill                        bool
	IsCritical                    bool
	AutomatedScoringFocus         string
	VisibilityCondition           VisibilityConditionStruct
	AnswerOptions                 []AnswerOptionStruct
	MultipleSelectOptionQuestions []MultipleSelectOptionQuestionStruct
}

type DefaultAnswersToStruct struct {
	HighestScore  bool
	NotApplicable bool
	LowestScore   bool
	UserDefined   bool
}

type EvaluationSettingsStruct struct {
	RevisionsEnabled             bool
	DisputesEnabled              bool
	DisputesAllowedPerEvaluation int
	DisputesAssignees            []DisputesAssigneeStruct
}

type DisputesAssigneeStruct struct {
	Type   string
	UserId string
}

type MultipleSelectOptionQuestionStruct struct {
	Text                string
	HelpText            string
	Type                string
	NaEnabled           bool
	CommentsRequired    bool
	IsKill              bool
	IsCritical          bool
	VisibilityCondition VisibilityConditionStruct
	AnswerOptions       []AnswerOptionStruct
}

type AnswerOptionStruct struct {
	Text                 string
	Value                int
	BuiltInType          string
	AssistanceConditions []AssistanceConditionStruct
}

type AssistanceConditionStruct struct {
	Operator string
	TopicIds []string
}

type VisibilityConditionStruct struct {
	CombiningOperation string
	Predicates         []string
}

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceEvaluationForm())
	regInstance.RegisterDataSource(ResourceType, DataSourceQualityFormsEvaluations())
	regInstance.RegisterExporter(ResourceType, EvaluationFormExporter())
}

// ResourceEvaluationForm registers the genesyscloud_quality_forms_evaluation resource with Terraform
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
			"context_id": {
				Description: "ID of the context of the evaluation form. This provides access to all versions of forms.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The name of the entity.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"published": {
				Description: "Specifies if the evaluation form is published. **Note:** A form cannot be modified if published is set to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"published_id": {
				Description: "The ID of the published evaluation form.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"question_groups": {
				Description: "A list of question groups.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem:        evaluationFormQuestionGroup,
			},
			"evaluation_settings": {
				Description: "Settings for evaluations associated with this form.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        evaluationFormEvaluationSettings,
			},
			"ai_scoring": {
				Description: "AI scoring settings for the evaluation form.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        evaluationFormAiScoring,
			},
			"dialect": {
				Description: "The language dialect for this evaluation form. Supported dialects: ar, cs, da, de, en-US, es, fi, fr, fr-CA, he, hi, it, ja, ko, nl, no, pl, pt-BR, pt-PT, ru, sv, th, tr, uk, zh-CN, zh-TW.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"modified_date": {
				Description: "Date time is represented as an ISO-8601 string. For example: yyyy-MM-ddTHH:mm:ss[.mmm]Z",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// EvaluationFormExporter returns the resourceExporter object used to hold the genesyscloud_quality_forms_evaluation exporter's config
func EvaluationFormExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllEvaluationForms),
		RefAttrs: map[string]*resourceExporter.RefAttrSettings{
			"question_groups.questions.answer_options.assistance_conditions.topic_ids": {
				RefType: sttTopic.ResourceType,
			},
			"question_groups.questions.multiple_select_option_questions.answer_options.assistance_conditions.topic_ids": {
				RefType: sttTopic.ResourceType,
			},
			"evaluation_settings.disputes_assignees.user_id": {
				RefType: user.ResourceType,
			},
		},
		CustomAttributeResolver: map[string]*resourceExporter.RefAttrCustomResolver{
			"question_groups.questions.answer_options.assistance_conditions.topic_ids": {
				ResolveToDataSourceFunc: resourceExporter.SpeechAndTextAnalyticsTopicIdResolver,
			},
			"question_groups.questions.multiple_select_option_questions.answer_options.assistance_conditions.topic_ids": {
				ResolveToDataSourceFunc: resourceExporter.SpeechAndTextAnalyticsTopicIdResolver,
			},
		},
		AllowZeroValues: []string{
			"question_groups.questions.answer_options.value",
			"question_groups.questions.multiple_select_option_questions.answer_options.value",
			"question_groups.weight",
		},
		ExcludedAttributes: []string{
			"question_groups.id",
			"question_groups.context_id",
			"question_groups.questions.id",
			"question_groups.questions.context_id",
			"question_groups.questions.answer_options.id",
			"question_groups.questions.answer_options.context_id",
			"question_groups.questions.multiple_select_option_questions.id",
			"question_groups.questions.multiple_select_option_questions.context_id",
			"question_groups.questions.multiple_select_option_questions.answer_options.id",
			"question_groups.questions.multiple_select_option_questions.answer_options.context_id",
			"modified_date",
		},
	}
}

// DataSourceQualityFormsEvaluations registers the genesyscloud_quality_forms_evaluation data source
func DataSourceQualityFormsEvaluations() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Evaluation Forms. Select an evaluations form by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceQualityFormsEvaluationsRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Evaluation Form name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
