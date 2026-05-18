package learning_modules

import (
	"fmt"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLearningModules(t *testing.T) {
	var (
		learningModuleResourceLabel     = "learning-module"
		learningModuleDataResourceLabel = "learning-module-data"

		learningModuleName                 = "terraform-learning-module-" + uuid.NewString()
		learningModuleDescription          = "terraform-learning-module-description-" + uuid.NewString()
		learningModuleCompletionTimeInDays = 10
		learningModuleType                 = "Native"
		learningModuleCoverArtId           = uuid.NewString()
		learningModuleLengthInMinutes      = 15
		learningModuleExcludedFromCatalog  = true
		learningModuleExternalId           = ""
		learningModuleEnforceContentOrder  = true
		learningModuleIsPublished          = false
	)

	informSteps := []InformStepStruct{
		{
			Type:        "Url",
			Name:        "inform-step-1",
			Value:       "https://www.example.com",
			Order:       1,
			DisplayName: "Inform Step 1",
			Description: "Inform Step 1",
		},
		{
			Type:        "RichText",
			Name:        "inform-step-2",
			Value:       "inform-step-2",
			Order:       2,
			DisplayName: "Inform Step 2",
			Description: "Inform Step 2",
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
				Config: GenerateLearningModuleResource(
					learningModuleResourceLabel,
					learningModuleName,
					learningModuleDescription,
					learningModuleCompletionTimeInDays,
					informSteps,
					learningModuleType,
					learningModuleCoverArtId,
					learningModuleLengthInMinutes,
					learningModuleExcludedFromCatalog,
					learningModuleExternalId,
					learningModuleEnforceContentOrder,
					&reviewAssessmentResults,
					&assessmentForm,
					learningModuleIsPublished,
				) + generateLearningModulesDataSource(
					learningModuleDataResourceLabel,
					learningModuleName,
					ResourceType+"."+learningModuleResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data."+ResourceType+"."+learningModuleDataResourceLabel, "id",
						"genesyscloud_learning_modules."+learningModuleResourceLabel, "id",
					),
				),
			},
		},
	})
}

func generateLearningModulesDataSource(
	resourceLabel string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, ResourceType, resourceLabel, name, dependsOnResource)
}
