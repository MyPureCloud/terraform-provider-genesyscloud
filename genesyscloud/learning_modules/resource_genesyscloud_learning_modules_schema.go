package learning_modules

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	registrar "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_register"
)

/*
ResourceName is defined in this file along with four functions:

1.  The registration code that registers the Datasource, Resource and Exporter for the package.
2.  The resource schema definitions for the learning_modules resource.
3.  The datasource schema definitions for the learning_modules datasource.
4.  The resource exporter configuration for the learning_modules exporter.
*/
const ResourceName = "genesyscloud_learning_modules"
const ResourceType = ResourceName

// SetRegistrar registers all the resources, datasources and exporters in the package
func SetRegistrar(regInstance registrar.Registrar) {
	regInstance.RegisterResource(ResourceName, ResourceLearningModules())
	regInstance.RegisterDataSource(ResourceName, DataSourceLearningModules())
	regInstance.RegisterExporter(ResourceName, LearningModulesExporter())
}

// ResourceLearningModules registers the genesyscloud_learning_modules resource with Terraform
func ResourceLearningModules() *schema.Resource {
	informStepsResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  `The learning module inform step type`,
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{`Url`, `Content`, `GenesysBuiltInCourse`, `RichText`, `Scorm`}, false),
			},
			"name": {
				Description: `The name of the inform step or content`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			"value": {
				Description: `The value for inform step`,
				Required:    true,
				Type:        schema.TypeString,
			},
			"sharing_uri": {
				Description: `The sharing uri for Content type inform step`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			"content_type": {
				Description: `The document type for Content type Inform step`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			"order": {
				Description: `The order of inform step in a learning module`,
				Required:    true,
				Type:        schema.TypeInt,
			},
			"display_name": {
				Description: `The display name for the inform step`,
				Optional:    true,
				Type:        schema.TypeString,
			},
			"description": {
				Description: `The description for the inform step`,
				Optional:    true,
				Type:        schema.TypeString,
			},
		},
	}

	assessmentAnswerOptions := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "The ID of the survey answer option.",
				Optional:    true,
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

	assessmentVisibilityCondition := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"combining_operation": {
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

	assessmentQuestion := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"text": {
				Description: "The question text",
				Type:        schema.TypeString,
				Required:    true,
			},
			"help_text": {
				Description: "Help text for the question.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
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
				Elem:        assessmentVisibilityCondition,
			},
			"answer_options": {
				Description: "Options from which to choose an answer for this question. Only used by Multiple Choice type questions.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        assessmentAnswerOptions,
			},
			"max_response_characters": {
				Description: "How many characters are allowed in the text response to this question. Used by Free Text question types.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"is_kill": {
				Description: "Does an incorrect answer to this question mark the form as having a failed kill question. Only used by Multiple Choice type questions.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"is_critical": {
				Description: "Does this question contribute to the critical score. Only used by Multiple Choice type questions.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}

	assessmentQuestionGroup := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the question group.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Description: "The question group name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "The question group type",
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
				Optional:    true,
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
				Elem:        assessmentQuestion,
			},
			"visibility_condition": {
				Description: "Defines conditions where question would be visible",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem:        assessmentVisibilityCondition,
			},
		},
	}

	assessmentFormResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"pass_percent": {
				Description: "The pass percent for the assessment form",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"id": {
				Description: "The ID of the assessment form",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"question_groups": {
				Description: "The question groups for the assessment",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        assessmentQuestionGroup,
			},
		},
	}

	reviewAssessmentResultsResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"by_assignees": {
				Description: "If true, learning assignment results can be seen in detail by assignees",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"by_viewers": {
				Description: "If true, learning assignment results can be seen in detail by people who are eligible to view",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}

	autoAssignResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Description: "Whether the rule is enabled for the module",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"rule_id": {
				Description: "The ID of the rule",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}

	return &schema.Resource{
		Description: `Genesys Cloud learning modules`,

		CreateContext: provider.CreateWithPooledClient(createLearningModule),
		ReadContext:   provider.ReadWithPooledClient(readLearningModule),
		UpdateContext: provider.UpdateWithPooledClient(updateLearningModule),
		DeleteContext: provider.DeleteWithPooledClient(deleteLearningModule),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of learning module",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The description of learning module",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"completion_time_in_days": {
				Description: "The completion time of learning module in days",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"inform_steps": {
				Description: "The list of inform steps in a learning module",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        informStepsResource,
			},
			"type": {
				Description:  "The type of the learning module. Informational, AssessedContent and Assessment are deprecated",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Informational", "AssessedContent", "Assessment", "External", "Native"}, false),
			},
			"assessment_form": {
				Description: "The assessment form for learning module",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        assessmentFormResource,
				MaxItems:    1,
			},
			"cover_art_id": {
				Description: "The cover art ID for the learning module",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"length_in_minutes": {
				Description: "The recommended time in minutes to complete the module",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"excluded_from_catalog": {
				Description: "If true, learning module is excluded when retrieving modules for manual assignment",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"external_id": {
				Description: "The external ID of the learning module. Maximum length: 50 characters.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enforce_content_order": {
				Description: "If true, learning module content should be viewed one by one in order",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"review_assessment_results": {
				Description: "Allows to view Assessment results in detail",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        reviewAssessmentResultsResource,
				MaxItems:    1,
			},
			"auto_assign": {
				Description: "The auto assign for the learning module",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        autoAssignResource,
				MaxItems:    1,
			},
			"is_published": {
				Description: "Specifies if the learning module is published.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

// LearningModulesExporter returns the resourceExporter object used to hold the genesyscloud_learning_modules exporter's config
func LearningModulesExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: provider.GetAllWithPooledClient(getAllAuthLearningModules),
	}
}

// DataSourceLearningModules registers the genesyscloud_learning_modules data source
func DataSourceLearningModules() *schema.Resource {
	return &schema.Resource{
		Description: `Data source for Genesys Cloud Learning Modules. Select a Learning Module by name.`,
		ReadContext: provider.ReadWithPooledClient(dataSourceLearningModulesRead),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: `Learning Module name.`,
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
