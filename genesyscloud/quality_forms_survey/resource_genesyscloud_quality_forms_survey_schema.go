package quality_forms_survey

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const ResourceType = "genesyscloud_quality_forms_survey"

// SetRegistrar registers all of the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceType, ResourceQualityFormsSurvey())
	regInstance.RegisterDataSource(ResourceType, DataSourceQualityFormsSurvey())
	regInstance.RegisterExporter(ResourceType, QualityFormsSurveyExporter())
}

var (
	surveyQuestionGroup = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the survey question group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of display question in question group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"na_enabled": {
				Description: "Specifies whether a not applicable answer is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"questions": {
				Description: "Questions inside the group",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem:        surveyQuestion,
			},
			"visibility_condition": {
				Description: "Defines conditions where question would be visible",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        surveyFormVisibilityCondition,
			},
		},
	}

	surveyQuestion = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the survey question.",
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
				Description:  "Valid Values: multipleChoiceQuestion, freeTextQuestion, npsQuestion, readOnlyTextBlockQuestion",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "multipleChoiceQuestion",
				ValidateFunc: validation.StringInSlice([]string{"multipleChoiceQuestion", "freeTextQuestion", "npsQuestion", "readOnlyTextBlockQuestion"}, false),
			},
			"na_enabled": {
				Description: "Specifies whether a not applicable answer is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"visibility_condition": {
				Description: "Defines conditions where question would be visible",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        surveyFormVisibilityCondition,
			},
			"answer_options": {
				Description: "Options from which to choose an answer for this question.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        surveyFormAnswerOptions,
			},
			"max_response_characters": {
				Description: "How many characters are allowed in the text response to this question. Used by NPS and Free Text question types.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"explanation_prompt": {
				Description: "Prompt for details explaining the chosen NPS score. Used by NPS questions.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}

	surveyFormVisibilityCondition = &schema.Resource{
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

	surveyFormAnswerOptions = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "The ID of the survey answer option.",
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
			"assistance_conditions": {
				Description: "Options from which to choose an answer for this question.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        assistanceConditionsResource,
			},
		},
	}

	assistanceConditionsResource = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"operator": {
				Description:  "List of assistance conditions which are combined together with a logical AND operator. Eg ( assistanceCondtion1 && assistanceCondition2 ) wherein assistanceCondition could be ( EXISTS topic1 || topic2 || ... ) or (NOTEXISTS topic3 || topic4 || ...).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"EXISTS", "NOTEXISTS"}, true),
			},
			"topic_ids": {
				Description: "List of topicIds within the assistance condition which would be combined together using logical OR operator. Eg ( topicId_1 || topicId_2 ) .",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
)

// ResourceQualityFormsSurvey registers the genesyscloud_quality_forms_survey resource with Terraform
func ResourceQualityFormsSurvey() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Survey Forms",
		CreateContext: provider.CreateWithPooledClient(createSurveyForm),
		ReadContext:   provider.ReadWithPooledClient(readSurveyForm),
		UpdateContext: provider.UpdateWithPooledClient(updateSurveyForm),
		DeleteContext: provider.DeleteWithPooledClient(deleteSurveyForm),
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
				Description: "Specifies if the survey form is published.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"disabled": {
				Description: "Is this form disabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"language": {
				Description:  "Language for survey viewer localization. Currently localized languages: da, de, en-US, es, fi, fr, it, ja, ko, nl, no, pl, pt-BR, sv, th, tr, zh-CH, zh-TW",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"da", "de", "en-US", "es", "fi", "fr", "it", "ja", "ko", "nl", "no", "pl", "pt-BR", "sv", "th", "tr", "zh-CH", "zh-TW"}, false),
			},
			"header": {
				Description: "Markdown text for the top of the form.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"footer": {
				Description: "Markdown text for the bottom of the form.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"question_groups": {
				Description: "A list of question groups.",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem:        surveyQuestionGroup,
			},
		},
	}
}

// QualityFormsSurveyExporter returns the resourceExporter object used to hold the genesyscloud_quality_forms_survey exporter's config
func QualityFormsSurveyExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllSurveyForms),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
		AllowZeroValues:  []string{"question_groups.questions.answer_options.value"},
		ExcludedAttributes: []string{
			"question_groups.id",
			"question_groups.questions.id",
			"question_groups.questions.answer_options.id",
		},
	}
}

// DataSourceQualityFormsSurvey registers the genesyscloud_quality_forms_survey data source
func DataSourceQualityFormsSurvey() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Survey Forms. Select a form by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceQualityFormsSurveyRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Survey form name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
