package quality_forms_evaluation

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

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

type EvaluationFormQuestionGroupStruct struct {
	Name                    string
	DefaultAnswersToHighest bool
	DefaultAnswersToNA      bool
	NaEnabled               bool
	Weight                  float32
	ManualWeight            bool
	Questions               []EvaluationFormQuestionStruct
	VisibilityCondition     VisibilityConditionStruct
}

type EvaluationFormStruct struct {
	Name           string
	Published      bool
	QuestionGroups []EvaluationFormQuestionGroupStruct
}

type EvaluationFormQuestionStruct struct {
	Text                string
	HelpText            string
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

// EvaluationFormExporter returns the resourceExporter object used to hold the genesyscloud_quality_forms_evaluation exporter's config
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
