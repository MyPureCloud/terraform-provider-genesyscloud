package learning_modules

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

func TestAccResourceLearningModules(t *testing.T) {
	learningModuleResourceLabel1 := "test-learning-module-1"
	learningModuleName := "terraform-learning-module-" + uuid.NewString()
	learningModuleNameUpdated := "terraform-learning-module-updated-" + uuid.NewString()
	learningModuleDescription := "terraform-learning-module-description-" + uuid.NewString()
	learningModuleDescriptionUpdated := "terraform-learning-module-description-updated-" + uuid.NewString()
	learningModuleCompletionTimeInDays := 10
	learningModuleCompletionTimeInDaysUpdated := 15
	learningModuleType := "Native"
	learningModuleCoverArtId := uuid.NewString()
	learningModuleCoverArtIdUpdated := uuid.NewString()
	learningModuleLengthInMinutes := 15
	learningModuleLengthInMinutesUpdated := 30
	learningModuleExcludedFromCatalog := true
	learningModuleExcludedFromCatalogUpdated := false
	learningModuleExternalId := ""
	learningModuleEnforceContentOrder := true
	learningModuleEnforceContentOrderUpdated := false
	learningModuleIsPublished := false

	informSteps := []InformStepStruct{
		{
			Type:        "Url",
			Name:        "inform-step-1",
			Value:       "https://www.example.com",
			Order:       1,
			DisplayName: "Inform Step 1",
			Description: "Inform Step 1",
		},
	}

	reviewAssessmentResults := ReviewAssessmentResultsStruct{
		ByAssignees: true,
		ByViewers:   true,
	}

	assessmentForm := AssessmentFormStruct{
		PassPercent: 80,
		QuestionGroups: []QuestionGroupStruct{
			{
				Name:                    "question-group-1",
				Type:                    "questionGroup",
				DefaultAnswersToHighest: true,
				DefaultAnswersToNA:      true,
				NaEnabled:               true,
				Weight:                  1,
				ManualWeight:            true,
				Questions: []QuestionStruct{
					{
						Type:                  "multipleChoiceQuestion",
						Text:                  "question-1",
						HelpText:              "question-1",
						NaEnabled:             true,
						CommentsRequired:      true,
						IsKill:                true,
						IsCritical:            true,
						MaxResponseCharacters: 0,
						AnswerOptions: []AnswerOptionStruct{
							{
								Text:  "answer-1",
								Value: 1,
							},
							{
								Text:  "answer-2",
								Value: 2,
							},
						},
					},
					{
						Type:                  "freeTextQuestion",
						Text:                  "question-2",
						HelpText:              "question-2",
						MaxResponseCharacters: 100,
						VisibilityCondition: &VisibilityConditionStruct{
							CombiningOperation: "AND",
							Predicates:         []string{"/form/questionGroup/0/question/0/answer/0"},
						},
					},
				},
			},
			{
				Name: "question-group-2",
				Type: "questionGroup",
				Questions: []QuestionStruct{
					{
						Type:                  "freeTextQuestion",
						Text:                  "question-3",
						MaxResponseCharacters: 100,
					},
				},
				VisibilityCondition: &VisibilityConditionStruct{
					CombiningOperation: "AND",
					Predicates:         []string{"/form/questionGroup/0/question/0/answer/0"},
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateLearningModuleResource(
					learningModuleResourceLabel1,
					learningModuleName,
					learningModuleDescription,
					learningModuleCompletionTimeInDays,
					nil,
					learningModuleType,
					learningModuleCoverArtId,
					learningModuleLengthInMinutes,
					learningModuleExcludedFromCatalog,
					learningModuleExternalId,
					learningModuleEnforceContentOrder,
					nil,
					&assessmentForm,
					learningModuleIsPublished,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "name", learningModuleName),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "description", learningModuleDescription),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "completion_time_in_days", fmt.Sprint(learningModuleCompletionTimeInDays)),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "type", learningModuleType),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "cover_art_id", learningModuleCoverArtId),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "length_in_minutes", fmt.Sprint(learningModuleLengthInMinutes)),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "excluded_from_catalog", fmt.Sprint(learningModuleExcludedFromCatalog)),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "external_id", learningModuleExternalId),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "enforce_content_order", fmt.Sprint(learningModuleEnforceContentOrder)),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "review_assessment_results.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.0.name", assessmentForm.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.0.visibility_condition.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.0.questions.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.0.questions.0.text", assessmentForm.QuestionGroups[0].Questions[0].Text),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.0.questions.0.visibility_condition.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.0.questions.0.answer_options.#", "2"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.0.questions.0.answer_options.0.text", assessmentForm.QuestionGroups[0].Questions[0].AnswerOptions[0].Text),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.0.questions.1.text", assessmentForm.QuestionGroups[0].Questions[1].Text),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.0.questions.1.visibility_condition.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.0.questions.1.answer_options.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.1.name", assessmentForm.QuestionGroups[1].Name),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.1.visibility_condition.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.1.questions.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.1.questions.0.text", assessmentForm.QuestionGroups[1].Questions[0].Text),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.1.questions.0.visibility_condition.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.0.question_groups.1.questions.0.answer_options.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "inform_steps.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "is_published", fmt.Sprint(learningModuleIsPublished)),
				),
			},
			{
				// Update
				Config: GenerateLearningModuleResource(
					learningModuleResourceLabel1,
					learningModuleNameUpdated,
					learningModuleDescriptionUpdated,
					learningModuleCompletionTimeInDaysUpdated,
					informSteps,
					learningModuleType,
					learningModuleCoverArtIdUpdated,
					learningModuleLengthInMinutesUpdated,
					learningModuleExcludedFromCatalogUpdated,
					learningModuleExternalId,
					learningModuleEnforceContentOrderUpdated,
					&reviewAssessmentResults,
					nil,
					learningModuleIsPublished,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "name", learningModuleNameUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "description", learningModuleDescriptionUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "completion_time_in_days", fmt.Sprint(learningModuleCompletionTimeInDaysUpdated)),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "type", learningModuleType),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "cover_art_id", learningModuleCoverArtIdUpdated),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "length_in_minutes", fmt.Sprint(learningModuleLengthInMinutesUpdated)),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "excluded_from_catalog", fmt.Sprint(learningModuleExcludedFromCatalogUpdated)),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "external_id", learningModuleExternalId),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "enforce_content_order", fmt.Sprint(learningModuleEnforceContentOrderUpdated)),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "review_assessment_results.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "review_assessment_results.0.by_assignees", fmt.Sprint(reviewAssessmentResults.ByAssignees)),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "review_assessment_results.0.by_viewers", fmt.Sprint(reviewAssessmentResults.ByViewers)),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "assessment_form.#", "0"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "inform_steps.#", "1"),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "inform_steps.0.value", informSteps[0].Value),
					resource.TestCheckResourceAttr(ResourceType+"."+learningModuleResourceLabel1, "is_published", fmt.Sprint(learningModuleIsPublished)),
				),
			},
			{
				// Import/Read
				ResourceName:      ResourceType + "." + learningModuleResourceLabel1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyLearningModuleDestroyed,
	})
}

func testVerifyLearningModuleDestroyed(state *terraform.State) error {
	learningApi := platformclientv2.NewLearningApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		learningModule, resp, err := learningApi.GetLearningModule(rs.Primary.ID, []string{})
		if learningModule != nil {
			continue
		}

		if learningModule != nil {
			return fmt.Errorf("Learning module (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// Learning module not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All Learning modules destroyed
	return nil
}
