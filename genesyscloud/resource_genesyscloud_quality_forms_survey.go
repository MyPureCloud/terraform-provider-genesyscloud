package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"

	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mypurecloud/platform-client-sdk-go/v115/platformclientv2"
)

type SurveyFormStruct struct {
	Name           string
	Published      bool
	Disabled       bool
	ContextId      int
	Language       string
	Header         string
	Footer         string
	QuestionGroups []SurveyFormQuestionGroupStruct
}

type SurveyFormQuestionGroupStruct struct {
	Name                string
	NaEnabled           bool
	Questions           []SurveyFormQuestionStruct
	VisibilityCondition VisibilityConditionStruct
}

type SurveyFormQuestionStruct struct {
	Text                  string
	HelpText              string
	VarType               string
	NaEnabled             bool
	VisibilityCondition   VisibilityConditionStruct
	AnswerOptions         []AnswerOptionStruct
	MaxResponseCharacters int
	ExplanationPrompt     string
}

var (
	surveyQuestionGroup = &schema.Resource{
		Schema: map[string]*schema.Schema{
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
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        assistanceConditions,
			},
		},
	}

	assistanceConditions = &schema.Resource{
		Schema: map[string]*schema.Schema{
			"operator": {
				Description: "List of assistance conditions which are combined together with a logical AND operator. Eg ( assistanceCondtion1 && assistanceCondition2 ) wherein assistanceCondition could be ( EXISTS topic1 || topic2 || ... ) or (NOTEXISTS topic3 || topic4 || ...).",
				Type:        schema.TypeString,
				Optional:    true,
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

func getAllSurveyForms(_ context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	qualityAPI := platformclientv2.NewQualityApiWithConfig(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		surveyForms, _, getErr := qualityAPI.GetQualityFormsSurveys(pageSize, pageNum, "", "", "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of survey forms %v", getErr)
		}

		if surveyForms.Entities == nil || len(*surveyForms.Entities) == 0 {
			break
		}

		for _, surveyForm := range *surveyForms.Entities {
			resources[*surveyForm.Id] = &resourceExporter.ResourceMeta{Name: *surveyForm.Name}
		}
	}

	return resources, nil
}

func SurveyFormExporter() *resourceExporter.ResourceExporter {
	return &resourceExporter.ResourceExporter{
		GetResourcesFunc: GetAllWithPooledClient(getAllSurveyForms),
		RefAttrs:         map[string]*resourceExporter.RefAttrSettings{}, // No references
		AllowZeroValues:  []string{"question_groups.questions.answer_options.value"},
	}
}

func ResourceSurveyForm() *schema.Resource {
	return &schema.Resource{
		Description:   "Genesys Cloud Survey Forms",
		CreateContext: CreateWithPooledClient(createSurveyForm),
		ReadContext:   ReadWithPooledClient(readSurveyForm),
		UpdateContext: UpdateWithPooledClient(updateSurveyForm),
		DeleteContext: DeleteWithPooledClient(deleteSurveyForm),
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

func createSurveyForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	language := d.Get("language").(string)
	header := d.Get("header").(string)
	footer := d.Get("footer").(string)
	disabled := d.Get("disabled").(bool)
	published := d.Get("published").(bool)

	questionGroups, qgErr := buildSurveyQuestionGroups(d)
	if qgErr != nil {
		return qgErr
	}

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	log.Printf("Creating Survey Form %s", name)
	form, _, err := qualityAPI.PostQualityFormsSurveys(platformclientv2.Surveyform{
		Name:           &name,
		Disabled:       &disabled,
		Language:       &language,
		Header:         &header,
		Footer:         &footer,
		QuestionGroups: questionGroups,
	})
	if err != nil {
		return diag.Errorf("Failed to create survey form %s: %s", name, err)
	}

	// Make sure form is properly created
	time.Sleep(2 * time.Second)

	formId := form.Id

	// Publishing
	if published {
		_, _, err := qualityAPI.PostQualityPublishedformsSurveys(platformclientv2.Publishform{
			Id:        formId,
			Published: &published,
		})
		if err != nil {
			return diag.Errorf("Failed to publish survey form %s", name)
		}
	}

	d.SetId(*formId)

	log.Printf("Created survey form %s %s", name, *form.Id)
	return readSurveyForm(ctx, d, meta)
}

func readSurveyForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)
	log.Printf("Reading survey form %s", d.Id())

	return WithRetriesForRead(ctx, d, func() *retry.RetryError {
		surveyForm, resp, getErr := qualityAPI.GetQualityFormsSurvey(d.Id())
		if getErr != nil {
			if IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("Failed to read survey form %s: %s", d.Id(), getErr))
			}
			return retry.NonRetryableError(fmt.Errorf("Failed to read survey form %s: %s", d.Id(), getErr))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceSurveyForm())
		if surveyForm.Name != nil {
			d.Set("name", *surveyForm.Name)
		}
		if surveyForm.Disabled != nil {
			d.Set("disabled", *surveyForm.Disabled)
		}
		if surveyForm.Language != nil {
			d.Set("language", *surveyForm.Language)
		}
		if surveyForm.Header != nil {
			d.Set("header", *surveyForm.Header)
		}
		if surveyForm.Footer != nil {
			d.Set("footer", *surveyForm.Footer)
		}
		if surveyForm.Published != nil {
			d.Set("published", *surveyForm.Published)
		}
		if surveyForm.QuestionGroups != nil {
			d.Set("question_groups", flattenSurveyQuestionGroups(surveyForm.QuestionGroups))
		}

		return cc.CheckState()
	})
}

func updateSurveyForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	language := d.Get("language").(string)
	header := d.Get("header").(string)
	footer := d.Get("footer").(string)
	disabled := d.Get("disabled").(bool)
	published := d.Get("published").(bool)

	questionGroups, qgErr := buildSurveyQuestionGroups(d)
	if qgErr != nil {
		return qgErr
	}

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	diagErr := RetryWhen(IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {

		// Get the latest unpublished version of the form
		formVersions, getResp, err := qualityAPI.GetQualityFormsSurveyVersions(d.Id(), 25, 1)
		if err != nil {
			return getResp, diag.Errorf("Failed to get survey form versions %s", name)
		}

		versions := *formVersions.Entities
		latestUnpublishedVersion := ""
		for _, v := range versions {
			if !*v.Published {
				latestUnpublishedVersion = *v.Id
			}
		}

		log.Printf("Updating Survey Form %s", name)
		form, putResp, err := qualityAPI.PutQualityFormsSurvey(latestUnpublishedVersion, platformclientv2.Surveyform{
			Name:           &name,
			Disabled:       &disabled,
			Language:       &language,
			Header:         &header,
			Footer:         &footer,
			QuestionGroups: questionGroups,
		})
		if err != nil {
			return putResp, diag.Errorf("Failed to update survey form %s: %v", name, err)
		}
		log.Printf("Updated survey form %s %s", name, *form.Id)

		// Set published property on survey form update.
		if published {
			_, postResp, err := qualityAPI.PostQualityPublishedformsSurveys(platformclientv2.Publishform{
				Id:        form.Id,
				Published: &published,
			})
			if err != nil {
				return postResp, diag.Errorf("Failed to publish survey form %s", name)
			}
		} else {
			// If published property is reset to false, set the resource Id to the latest unpublished form
			d.SetId(*form.Id)
		}
		return putResp, nil
	})
	if diagErr != nil {
		return diagErr
	}
	return readSurveyForm(ctx, d, meta)
}

func deleteSurveyForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	// Get the latest unpublished version of the form
	formVersions, _, err := qualityAPI.GetQualityFormsSurveyVersions(d.Id(), 25, 1)
	if err != nil {
		return diag.Errorf("Failed to get survey form versions %s", name)
	}
	versions := *formVersions.Entities
	latestUnpublishedVersion := ""
	for _, v := range versions {
		if !*v.Published {
			latestUnpublishedVersion = *v.Id
		}
	}
	d.SetId(latestUnpublishedVersion)

	log.Printf("Deleting survey form %s", name)
	if _, err := qualityAPI.DeleteQualityFormsSurvey(d.Id()); err != nil {
		return diag.Errorf("Failed to delete survey form %s: %s", name, err)
	}

	return WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := qualityAPI.GetQualityFormsSurvey(d.Id())
		if err != nil {
			if IsStatus404(resp) {
				// survey form deleted
				log.Printf("Deleted survey form %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("Error deleting survey form %s: %s", d.Id(), err))
		}

		return retry.RetryableError(fmt.Errorf("Survey form %s still exists", d.Id()))
	})
}

func buildSurveyQuestionGroups(d *schema.ResourceData) (*[]platformclientv2.Surveyquestiongroup, diag.Diagnostics) {
	questionGroupType := "questionGroup"

	var surveyQuestionGroups []platformclientv2.Surveyquestiongroup
	if questionGroups, ok := d.GetOk("question_groups"); ok {
		questionGroupList := questionGroups.([]interface{})
		for _, questionGroup := range questionGroupList {
			questionGroupsMap := questionGroup.(map[string]interface{})

			questionGroupName := questionGroupsMap["name"].(string)
			naEnabled := questionGroupsMap["na_enabled"].(bool)
			questions := questionGroupsMap["questions"].([]interface{})

			sdkquestionGroup := platformclientv2.Surveyquestiongroup{
				Name:      &questionGroupName,
				VarType:   &questionGroupType,
				NaEnabled: &naEnabled,
				Questions: buildSurveyQuestions(questions),
			}

			visibilityCondition := questionGroupsMap["visibility_condition"].([]interface{})
			sdkquestionGroup.VisibilityCondition = buildSdkVisibilityCondition(visibilityCondition)

			surveyQuestionGroups = append(surveyQuestionGroups, sdkquestionGroup)
		}
	}

	return &surveyQuestionGroups, nil
}

func buildSurveyQuestions(questions []interface{}) *[]platformclientv2.Surveyquestion {
	sdkQuestions := make([]platformclientv2.Surveyquestion, 0)
	for _, question := range questions {
		questionsMap := question.(map[string]interface{})
		text := questionsMap["text"].(string)
		helpText := questionsMap["help_text"].(string)
		questionType := questionsMap["type"].(string)
		naEnabled := questionsMap["na_enabled"].(bool)
		answerQuestions := questionsMap["answer_options"].([]interface{})
		maxResponseCharacters := questionsMap["max_response_characters"].(int)

		sdkQuestion := platformclientv2.Surveyquestion{
			Text:                  &text,
			HelpText:              &helpText,
			VarType:               &questionType,
			NaEnabled:             &naEnabled,
			AnswerOptions:         buildSdkAnswerOptions(answerQuestions),
			MaxResponseCharacters: &maxResponseCharacters,
		}

		explanationPrompt := questionsMap["explanation_prompt"].(string)
		if explanationPrompt != "" {
			sdkQuestion.ExplanationPrompt = &explanationPrompt
		}

		visibilityCondition := questionsMap["visibility_condition"].([]interface{})
		sdkQuestion.VisibilityCondition = buildSdkVisibilityCondition(visibilityCondition)

		sdkQuestions = append(sdkQuestions, sdkQuestion)
	}

	return &sdkQuestions
}

func flattenSurveyQuestionGroups(questionGroups *[]platformclientv2.Surveyquestiongroup) []interface{} {
	if questionGroups == nil {
		return nil
	}

	questionGroupList := []interface{}{}

	for _, questionGroup := range *questionGroups {
		questionGroupMap := make(map[string]interface{})
		if questionGroup.Name != nil {
			questionGroupMap["name"] = *questionGroup.Name
		}
		if questionGroup.NaEnabled != nil {
			questionGroupMap["na_enabled"] = *questionGroup.NaEnabled
		}
		if questionGroup.Questions != nil {
			questionGroupMap["questions"] = flattenSurveyQuestions(questionGroup.Questions)
		}
		if questionGroup.VisibilityCondition != nil {
			questionGroupMap["visibility_condition"] = flattenVisibilityCondition(questionGroup.VisibilityCondition)
		}

		questionGroupList = append(questionGroupList, questionGroupMap)
	}
	return questionGroupList
}

func flattenSurveyQuestions(questions *[]platformclientv2.Surveyquestion) []interface{} {
	if questions == nil {
		return nil
	}

	questionList := []interface{}{}

	for _, question := range *questions {
		questionMap := make(map[string]interface{})
		if question.Text != nil {
			questionMap["text"] = *question.Text
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
		if question.VisibilityCondition != nil {
			questionMap["visibility_condition"] = flattenVisibilityCondition(question.VisibilityCondition)
		}
		if question.AnswerOptions != nil {
			questionMap["answer_options"] = flattenAnswerOptions(question.AnswerOptions)
		}
		if question.MaxResponseCharacters != nil {
			questionMap["max_response_characters"] = *question.MaxResponseCharacters
		}
		if question.ExplanationPrompt != nil {
			questionMap["explanation_prompt"] = *question.ExplanationPrompt
		}

		questionList = append(questionList, questionMap)
	}
	return questionList
}

func GenerateSurveyFormResource(resourceID string, surveyForm *SurveyFormStruct) string {
	form := fmt.Sprintf(`resource "genesyscloud_quality_forms_survey" "%s" {
		name = "%s"
		published = %v
		disabled = %v
        language = "%s"
        header = "%s"
        footer = "%s"
		%s
        %s
	}
	`, resourceID,
		surveyForm.Name,
		surveyForm.Published,
		surveyForm.Disabled,
		surveyForm.Language,
		surveyForm.Header,
		surveyForm.Footer,
		generateSurveyFormQuestionGroups(&surveyForm.QuestionGroups),
		generateLifeCycle(),
	)

	return form
}

func generateLifeCycle() string {
	return `
	lifecycle {
		ignore_changes = [
			question_groups[0].questions[0].type,
			question_groups[0].questions[1].type,
			question_groups[0].questions[2].type,
			question_groups[1].questions[0].type,
			question_groups[1].questions[1].type,
			question_groups[1].questions[2].type,
			question_groups[2].questions[0].type,
			question_groups[2].questions[1].type,
			question_groups[2].questions[2].type,
		]
	}
	`
}

func generateSurveyFormQuestions(questions *[]SurveyFormQuestionStruct) string {
	if questions == nil {
		return ""
	}

	questionsString := ""

	for _, question := range *questions {
		questionString := fmt.Sprintf(`
        questions {
            text = "%s"
            help_text = "%s"
            type = "%s"
            na_enabled = %v
            %s
            %s
            max_response_characters = %v
            explanation_prompt = "%s"
        }
        `, question.Text,
			question.HelpText,
			question.VarType,
			question.NaEnabled,
			GenerateFormVisibilityCondition(&question.VisibilityCondition),
			GenerateFormAnswerOptions(&question.AnswerOptions),
			question.MaxResponseCharacters,
			question.ExplanationPrompt,
		)

		questionsString += questionString
	}

	return questionsString
}

func generateSurveyFormQuestionGroups(questionGroups *[]SurveyFormQuestionGroupStruct) string {
	if questionGroups == nil {
		return ""
	}

	questionGroupsString := ""

	for _, questionGroup := range *questionGroups {
		questionGroupString := fmt.Sprintf(`
        question_groups {
            name = "%s"
            na_enabled = %v
            %s
            %s
        }
        `, questionGroup.Name,
			questionGroup.NaEnabled,
			generateSurveyFormQuestions(&questionGroup.Questions),
			GenerateFormVisibilityCondition(&questionGroup.VisibilityCondition),
		)

		questionGroupsString += questionGroupString
	}

	return questionGroupsString
}
