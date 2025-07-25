package quality_forms_evaluation

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

func TestAccResourceEvaluationFormBasic(t *testing.T) {
	formNameAttr := "TF Test Form " + uuid.NewString()
	formResourceLabel1 := "test-evaluation-form-1"
	answer1Text := "Yes"
	answer1Value := 1

	// Most basic evaluation form
	evaluationForm1 := EvaluationFormStruct{
		Name: formNameAttr,
		QuestionGroups: []EvaluationFormQuestionGroupStruct{
			{
				Name:   "Test Question Group 1",
				Weight: 1,
				Questions: []EvaluationFormQuestionStruct{
					{
						Text: "Did the agent perform the opening spiel?",
						AnswerOptions: []AnswerOptionStruct{
							{
								Text:  answer1Text,
								Value: answer1Value,
							},
							{
								Text:  "No",
								Value: 2,
							},
						},
					},
				},
			},
		},
	}

	// Duplicate form with additional questions
	evaluationForm2 := evaluationForm1
	evaluationForm2.Name = formNameAttr + " updated"
	evaluationForm2.QuestionGroups = append(evaluationForm2.QuestionGroups, EvaluationFormQuestionGroupStruct{
		Name:   "Test Question Group 2",
		Weight: 2,
		Questions: []EvaluationFormQuestionStruct{
			{
				Text: "Yet another yes or no question.",
				AnswerOptions: []AnswerOptionStruct{
					{
						Text:  "Yes",
						Value: 1,
					},
					{
						Text:  "No",
						Value: 0,
					},
				},
			},
			{
				Text: "Multiple Choice Question.",
				AnswerOptions: []AnswerOptionStruct{
					{
						Text:  "Option 1",
						Value: 1,
					},
					{
						Text:  "Option 2",
						Value: 2,
					},
					{
						Text:  "Option 3",
						Value: 3,
					},
				},
			},
		},
	})

	evaluationForm3 := evaluationForm2
	evaluationForm3.Published = true

	evaluationDraftId := ""

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateEvaluationFormResource(formResourceLabel1, &evaluationForm1),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// Set evaluationDraftId to the instance ID of formResourceLabel1
						evaluationDraftId = s.RootModule().Resources[ResourceType+"."+formResourceLabel1].Primary.ID
						return nil
					},
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "name", evaluationForm1.Name),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "published", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "published_id", ""),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.name", evaluationForm1.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.#", fmt.Sprint(len(evaluationForm1.QuestionGroups))),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.0.text", evaluationForm1.QuestionGroups[0].Questions[0].Text),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(evaluationForm1.QuestionGroups[0].Questions[0].AnswerOptions))),
				),
				Destroy:                   false,
				PreventPostDestroyRefresh: true,
			},
			{
				// Update and add some questions
				Config: GenerateEvaluationFormResource(formResourceLabel1, &evaluationForm2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "name", evaluationForm2.Name),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "published", util.FalseValue),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "published_id", ""),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.name", evaluationForm2.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.#", fmt.Sprint(len(evaluationForm2.QuestionGroups))),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.1.weight", fmt.Sprint(evaluationForm2.QuestionGroups[1].Weight)),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.0.text", evaluationForm2.QuestionGroups[0].Questions[0].Text),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.1.questions.0.text", evaluationForm2.QuestionGroups[1].Questions[0].Text),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.1.questions.1.text", evaluationForm2.QuestionGroups[1].Questions[1].Text),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(evaluationForm2.QuestionGroups[0].Questions[0].AnswerOptions))),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.1.questions.0.answer_options.#", fmt.Sprint(len(evaluationForm2.QuestionGroups[1].Questions[0].AnswerOptions))),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.1.questions.1.answer_options.#", fmt.Sprint(len(evaluationForm2.QuestionGroups[1].Questions[1].AnswerOptions))),
				),
				Destroy:                   false,
				PreventPostDestroyRefresh: true,
			},
			{
				// Publish Evaluation Form
				Config: GenerateEvaluationFormResource(formResourceLabel1, &evaluationForm3),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// Get the evaluation draft ID from the state
						newDraftId := s.RootModule().Resources[ResourceType+"."+formResourceLabel1].Primary.ID
						if evaluationDraftId == "" {
							return fmt.Errorf("evaluation draft ID not set")
						}
						if evaluationDraftId == newDraftId {
							return fmt.Errorf("evaluation draft ID is the same as the original draft ID: %s and %s", evaluationDraftId, newDraftId)
						}
						return nil
					},
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "name", evaluationForm3.Name),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "published", util.TrueValue),
					resource.TestCheckResourceAttrSet(ResourceType+"."+formResourceLabel1, "published_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.name", evaluationForm3.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.#", fmt.Sprint(len(evaluationForm3.QuestionGroups))),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.0.text", evaluationForm3.QuestionGroups[0].Questions[0].Text),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(evaluationForm3.QuestionGroups[0].Questions[0].AnswerOptions))),
				),
			},
			{
				// Import/Read
				ResourceName:            ResourceType + "." + formResourceLabel1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"published"},
			},
		},
		CheckDestroy: testVerifyEvaluationFormDestroyed,
	})
}

func TestAccResourceEvaluationFormCompleteWithPublish(t *testing.T) {
	formResourceLabel1 := "test-evaluation-form-1"

	// Complete evaluation form
	evaluationForm1 := EvaluationFormStruct{
		Name:      "terraform-form-evaluations-" + uuid.NewString(),
		Published: true,
		QuestionGroups: []EvaluationFormQuestionGroupStruct{
			{
				Name:                    "Test Question Group 1",
				DefaultAnswersToHighest: true,
				DefaultAnswersToNA:      true,
				NaEnabled:               true,
				Weight:                  1,
				ManualWeight:            true,
				Questions: []EvaluationFormQuestionStruct{
					{
						Text: "Did the agent perform the opening spiel?",
						AnswerOptions: []AnswerOptionStruct{
							{
								Text:  "Yes",
								Value: 1,
							},
							{
								Text:  "No",
								Value: 0,
							},
						},
					},
					{
						Text:             "Did the agent greet the customer?",
						HelpText:         "Help text here",
						NaEnabled:        true,
						CommentsRequired: true,
						IsKill:           true,
						IsCritical:       true,
						VisibilityCondition: VisibilityConditionStruct{
							CombiningOperation: "AND",
							Predicates:         []string{"/form/questionGroup/0/question/0/answer/1"},
						},
						AnswerOptions: []AnswerOptionStruct{
							{
								Text:  "Yes",
								Value: 1,
							},
							{
								Text:  "No",
								Value: 0,
							},
						},
					},
				},
			},
			{
				Name:   "Test Question Group 2",
				Weight: 2,
				Questions: []EvaluationFormQuestionStruct{
					{
						Text: "Did the agent offer to sell product?",
						AnswerOptions: []AnswerOptionStruct{
							{
								Text:  "Yes",
								Value: 1,
							},
							{
								Text:  "No",
								Value: 0,
							},
						},
					},
				},
				VisibilityCondition: VisibilityConditionStruct{
					CombiningOperation: "AND",
					Predicates:         []string{"/form/questionGroup/0/question/0/answer/1"},
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
				Config: GenerateEvaluationFormResource(formResourceLabel1, &evaluationForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "name", evaluationForm1.Name),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "published", util.TrueValue),
					resource.TestCheckResourceAttrSet(ResourceType+"."+formResourceLabel1, "published_id"),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.name", evaluationForm1.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.default_answers_to_highest", strconv.FormatBool(evaluationForm1.QuestionGroups[0].DefaultAnswersToHighest)),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.default_answers_to_na", strconv.FormatBool(evaluationForm1.QuestionGroups[0].DefaultAnswersToNA)),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.na_enabled", strconv.FormatBool(evaluationForm1.QuestionGroups[0].NaEnabled)),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.manual_weight", strconv.FormatBool(evaluationForm1.QuestionGroups[0].ManualWeight)),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.1.visibility_condition.0.combining_operation", evaluationForm1.QuestionGroups[1].VisibilityCondition.CombiningOperation),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.1.visibility_condition.0.predicates.0", evaluationForm1.QuestionGroups[1].VisibilityCondition.Predicates[0]),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.#", fmt.Sprint(len(evaluationForm1.QuestionGroups))),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.1.text", evaluationForm1.QuestionGroups[0].Questions[1].Text),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.1.help_text", evaluationForm1.QuestionGroups[0].Questions[1].HelpText),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.1.na_enabled", strconv.FormatBool(evaluationForm1.QuestionGroups[0].Questions[1].NaEnabled)),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.1.comments_required", strconv.FormatBool(evaluationForm1.QuestionGroups[0].Questions[1].CommentsRequired)),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.1.is_kill", strconv.FormatBool(evaluationForm1.QuestionGroups[0].Questions[1].IsKill)),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.1.is_critical", strconv.FormatBool(evaluationForm1.QuestionGroups[0].Questions[1].IsCritical)),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.1.answer_options.#", fmt.Sprint(len(evaluationForm1.QuestionGroups[0].Questions[1].AnswerOptions))),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.1.visibility_condition.0.combining_operation", evaluationForm1.QuestionGroups[0].Questions[1].VisibilityCondition.CombiningOperation),
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "question_groups.0.questions.1.visibility_condition.0.predicates.0", evaluationForm1.QuestionGroups[0].Questions[1].VisibilityCondition.Predicates[0]),
				),
			},
			{
				// Import/Read
				ResourceName:            ResourceType + "." + formResourceLabel1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"published"},
			},
		},
		CheckDestroy: testVerifyEvaluationFormDestroyed,
	})
}

func TestAccResourceEvaluationFormRepublishing(t *testing.T) {
	formResourceLabel1 := "test-evaluation-form-1"

	// Most basic evaluation form
	evaluationForm1 := EvaluationFormStruct{
		Name:      "terraform-form-evaluations-" + uuid.NewString(),
		Published: true,
		QuestionGroups: []EvaluationFormQuestionGroupStruct{
			{
				Name:   "Test Question Group 1",
				Weight: 1,
				Questions: []EvaluationFormQuestionStruct{
					{
						Text: "Did the agent perform the opening spiel?",
						AnswerOptions: []AnswerOptionStruct{
							{
								Text:  "Yes",
								Value: 1,
							},
							{
								Text:  "No",
								Value: 0,
							},
						},
					},
				},
			},
		},
	}

	// Unpublish
	evaluationForm2 := evaluationForm1
	evaluationForm2.Published = false

	publishedId := ""

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Publish form on creation
				Config: GenerateEvaluationFormResource(formResourceLabel1, &evaluationForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "published", util.TrueValue),
					resource.TestCheckResourceAttrSet(ResourceType+"."+formResourceLabel1, "published_id"),
					func(s *terraform.State) error {
						publishedId = s.RootModule().Resources[ResourceType+"."+formResourceLabel1].Primary.Attributes["published_id"]
						if publishedId == "" {
							return fmt.Errorf("published_id is not set")
						} else {
							return nil
						}
					},
				),
			},
			{
				// Unpublish
				Config: GenerateEvaluationFormResource(formResourceLabel1, &evaluationForm2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "published", util.FalseValue),
					// Previous publish version is retained
					resource.TestCheckResourceAttrSet(ResourceType+"."+formResourceLabel1, "published_id"),
					func(s *terraform.State) error {
						publishedIdCheck := s.RootModule().Resources[ResourceType+"."+formResourceLabel1].Primary.Attributes["published_id"]
						if publishedIdCheck == "" {
							return fmt.Errorf("published_id is not set")
						}
						if publishedIdCheck != publishedId {
							return fmt.Errorf("published_id is not the same as the previous version: %s and %s", publishedIdCheck, publishedId)
						}
						return nil
					},
				),
			},
			{
				// republish should generate new publish_id
				Config: GenerateEvaluationFormResource(formResourceLabel1, &evaluationForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ResourceType+"."+formResourceLabel1, "published", util.TrueValue),
					resource.TestCheckResourceAttrSet(ResourceType+"."+formResourceLabel1, "published_id"),
					func(s *terraform.State) error {
						publishedIdCheck := s.RootModule().Resources[ResourceType+"."+formResourceLabel1].Primary.Attributes["published_id"]
						if publishedIdCheck == "" {
							return fmt.Errorf("published_id is not set")
						}
						if publishedIdCheck == publishedId {
							return fmt.Errorf("published_id is the same as the previous version: %s and %s", publishedIdCheck, publishedId)
						}
						return nil
					},
				),
			},
			{
				// Import/Read
				ResourceName:            ResourceType + "." + formResourceLabel1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"published"},
			},
		},
		CheckDestroy: testVerifyEvaluationFormDestroyed,
	})
}

func testVerifyEvaluationFormDestroyed(state *terraform.State) error {
	qualityAPI := platformclientv2.NewQualityApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != ResourceType {
			continue
		}

		form, resp, err := qualityAPI.GetQualityFormsEvaluation(rs.Primary.ID)
		if form != nil {
			continue
		}

		if form != nil {
			return fmt.Errorf("Evaluation form (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// Evaluation form not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All Evaluation forms destroyed
	return nil
}
