package quality_forms_survey

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/constants"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v165/platformclientv2"
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

type VisibilityConditionStruct struct {
	CombiningOperation string
	Predicates         []string
}

type SurveyFormQuestionGroupStruct struct {
	Name                string
	NaEnabled           bool
	Questions           []SurveyFormQuestionStruct
	VisibilityCondition VisibilityConditionStruct
}

type AssistanceConditionStruct struct {
	Operator string
	TopicIds []string
}

type AnswerOptionStruct struct {
	Text                 string
	Value                int
	AssistanceConditions []AssistanceConditionStruct
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

// getAllSurveyForms retrieves all survey forms and is used for the exporter
func getAllSurveyForms(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	proxy := getQualityFormsSurveyProxy(clientConfig)
	resources := make(resourceExporter.ResourceIDMetaMap)

	forms, resp, err := proxy.getAllQualityFormsSurvey(ctx, "")
	if err != nil {
		return nil, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get survey forms error: %s", err), resp)
	}

	for _, form := range *forms {
		resources[*form.Id] = &resourceExporter.ResourceMeta{BlockLabel: *form.Name}
	}

	return resources, nil
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityFormsSurveyProxy(sdkConfig)

	log.Printf("Creating Survey Form %s", name)
	form, resp, err := proxy.createQualityFormsSurvey(ctx, &platformclientv2.Surveyform{
		Name:           &name,
		Disabled:       &disabled,
		Language:       &language,
		Header:         &header,
		Footer:         &footer,
		QuestionGroups: questionGroups,
	})
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to create survey form %s error: %s", name, err), resp)
	}

	// Make sure form is properly created
	time.Sleep(2 * time.Second)

	// Add nil check for form and form.Id
	if form == nil || form.Id == nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Created form or form ID is nil for survey form %s", name), resp)
	}
	formId := form.Id

	// Publishing
	if published {
		if _, err := proxy.publishQualityFormsSurvey(ctx, *formId, published); err != nil {
			return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish survey form %s error: %s", name, err), resp)
		}
	}

	d.SetId(*formId)

	log.Printf("Created survey form %s %s", name, *form.Id)
	return readSurveyForm(ctx, d, meta)
}

func readSurveyForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityFormsSurveyProxy(sdkConfig)
	cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceQualityFormsSurvey(), constants.ConsistencyChecks(), ResourceType)

	log.Printf("Reading survey form %s", d.Id())
	return util.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		surveyForm, resp, getErr := proxy.getQualityFormsSurveyById(ctx, d.Id())
		if getErr != nil {
			if util.IsStatus404(resp) {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read survey form %s | error: %s", d.Id(), getErr), resp))
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read survey form %s | error: %s", d.Id(), getErr), resp))
		}

		// Add nil check for surveyForm
		if surveyForm == nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Received nil survey form for ID %s", d.Id()), resp))
		}

		if surveyForm.Name != nil {
			_ = d.Set("name", *surveyForm.Name)
		}

		if surveyForm.Language != nil {
			_ = d.Set("language", *surveyForm.Language)
		}
		if surveyForm.Header != nil {
			_ = d.Set("header", *surveyForm.Header)
		}
		if surveyForm.Footer != nil {
			_ = d.Set("footer", *surveyForm.Footer)
		}
		if surveyForm.QuestionGroups != nil {
			_ = d.Set("question_groups", flattenSurveyQuestionGroups(surveyForm.QuestionGroups))
		}

		// Published is always set to false, Check each form for a published version and set published accordingly
		formVersions, resp, err := proxy.getQualityFormsSurveyVersions(ctx, d.Id(), 25, 1)
		if err != nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Failed to read survey form versions %s | error: %s", *surveyForm.Id, err), resp))
		}

		// Add nil check for formVersions and its Entities
		if formVersions == nil || formVersions.Entities == nil {
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Received nil form versions for survey form %s", d.Id()), resp))
		}

		var published = false
		for _, s := range *formVersions.Entities {
			if *s.Published == true {
				published = true
			}
		}

		_ = d.Set("published", published)
		_ = d.Set("disabled", *surveyForm.Disabled)

		return cc.CheckState(d)
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

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityFormsSurveyProxy(sdkConfig)

	diagErr := util.RetryWhen(util.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest unpublished version of the form
		formVersions, resp, err := proxy.qualityApi.GetQualityFormsSurveyVersions(d.Id(), 25, 1)
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get survey form versions %s error: %s", name, err), resp)
		}

		// Add nil check for formVersions and its Entities
		if formVersions == nil || formVersions.Entities == nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("No form versions found for survey form %s", name), resp)
		}

		versions := *formVersions.Entities
		latestUnpublishedVersion := ""
		for _, v := range versions {
			// Add nil checks for v.Published and v.Id
			if v.Published != nil && v.Id != nil && !*v.Published {
				latestUnpublishedVersion = *v.Id
			}
		}

		// Check if we found an unpublished version
		if latestUnpublishedVersion == "" {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("No unpublished version found for survey form %s", name), resp)
		}

		log.Printf("Updating Survey Form %s", name)
		form, resp, err := proxy.updateQualityFormsSurvey(ctx, latestUnpublishedVersion, &platformclientv2.Surveyform{
			Name:           &name,
			Language:       &language,
			Header:         &header,
			Footer:         &footer,
			QuestionGroups: questionGroups,
		})
		if err != nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to update survey form %s error: %s", name, err), resp)
		}

		// Add nil check for form and form.Id
		if form == nil || form.Id == nil {
			return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Received nil form or form ID after update for %s", name), resp)
		}

		// Set published property on survey form update.
		if published {
			if _, err := proxy.publishQualityFormsSurvey(ctx, *form.Id, published); err != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to publish survey form %s error: %s", name, err), resp)
			}
		}

		if disabled {
			_, resp, err := proxy.patchQualityFormsSurvey(ctx, d.Id(), platformclientv2.Surveyform{
				Disabled: &disabled,
			})
			if err != nil {
				return resp, util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to disable survey form %s error: %s", name, err), resp)
			}
		}
		return resp, nil
	})

	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated survey form %s %s", name, d.Id())
	return readSurveyForm(ctx, d, meta)
}

func deleteSurveyForm(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*provider.ProviderMeta).ClientConfig
	proxy := getQualityFormsSurveyProxy(sdkConfig)

	// Get the latest unpublished version of the form
	formVersions, resp, err := proxy.getQualityFormsSurveyVersions(ctx, d.Id(), 25, 1)
	if err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to get survey form versions %s error: %s", name, err), resp)
	}

	if formVersions == nil || formVersions.Entities == nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("No form versions found for survey form %s", name), resp)
	}

	versions := *formVersions.Entities
	latestUnpublishedVersion := ""
	for _, v := range versions {
		if v.Published != nil && v.Id != nil && !*v.Published {
			latestUnpublishedVersion = *v.Id
		}
	}

	if latestUnpublishedVersion == "" {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("No unpublished version found for survey form %s", name), resp)
	}
	d.SetId(latestUnpublishedVersion)

	log.Printf("Deleting survey form %s", name)
	if resp, err := proxy.deleteQualityFormsSurvey(ctx, d.Id()); err != nil {
		return util.BuildAPIDiagnosticError(ResourceType, fmt.Sprintf("Failed to delete survey form %s error: %s", name, err), resp)
	}

	return util.WithRetries(ctx, 30*time.Second, func() *retry.RetryError {
		_, resp, err := proxy.getQualityFormsSurveyById(ctx, d.Id())
		if err != nil {
			if util.IsStatus404(resp) {
				log.Printf("Deleted survey form %s", d.Id())
				return nil
			}
			return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Error deleting survey form %s | error: %s", d.Id(), err), resp))
		}

		return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError(ResourceType, fmt.Sprintf("Survey form %s still exists", d.Id()), resp))
	})
}
