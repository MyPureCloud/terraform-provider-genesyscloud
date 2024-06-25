package genesyscloud

import (
	"fmt"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

func TestAccResourceSurveyFormBasic(t *testing.T) {
	formResource1 := "test-survey-form-1"

	// Most basic survey form
	surveyForm1 := SurveyFormStruct{
		Name:     "terraform-form-surveys-" + uuid.NewString(),
		Language: "en-US",
		QuestionGroups: []SurveyFormQuestionGroupStruct{
			{
				Name: "Test Question Group 1",
				Questions: []SurveyFormQuestionStruct{
					{
						Text:    "Did the agent perform the opening spiel?",
						VarType: "multipleChoiceQuestion",
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

	// Duplicate form with additional questions
	surveyForm2 := surveyForm1
	surveyForm2.Name = "terraform-survey-name-2"
	surveyForm2.QuestionGroups = append(surveyForm2.QuestionGroups, SurveyFormQuestionGroupStruct{
		Name: "Test Question Group 2",
		Questions: []SurveyFormQuestionStruct{
			{
				Text:    "Yet another yes or no question.",
				VarType: "multipleChoiceQuestion",
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
				Text:    "Multiple Choice Question.",
				VarType: "multipleChoiceQuestion",
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

	surveyForm3 := surveyForm1
	surveyForm3.Published = true

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: GenerateSurveyFormResource(formResource1, &surveyForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "name", surveyForm1.Name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.name", surveyForm1.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.#", fmt.Sprint(len(surveyForm1.QuestionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.text", surveyForm1.QuestionGroups[0].Questions[0].Text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(surveyForm1.QuestionGroups[0].Questions[0].AnswerOptions))),
				),
			},
			{
				// Update and add some questions
				Config: GenerateSurveyFormResource(formResource1, &surveyForm2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "name", surveyForm2.Name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.name", surveyForm2.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.#", fmt.Sprint(len(surveyForm2.QuestionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.text", surveyForm2.QuestionGroups[0].Questions[0].Text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.0.text", surveyForm2.QuestionGroups[1].Questions[0].Text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.1.text", surveyForm2.QuestionGroups[1].Questions[1].Text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(surveyForm2.QuestionGroups[0].Questions[0].AnswerOptions))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.0.answer_options.#", fmt.Sprint(len(surveyForm2.QuestionGroups[1].Questions[0].AnswerOptions))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.1.answer_options.#", fmt.Sprint(len(surveyForm2.QuestionGroups[1].Questions[1].AnswerOptions))),
				),
			},
			{
				// Publish Survey Form
				Config: GenerateSurveyFormResource(formResource1, &surveyForm3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "name", surveyForm3.Name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", util.TrueValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.name", surveyForm3.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.#", fmt.Sprint(len(surveyForm3.QuestionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.text", surveyForm3.QuestionGroups[0].Questions[0].Text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(surveyForm3.QuestionGroups[0].Questions[0].AnswerOptions))),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_quality_forms_survey." + formResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySurveyFormDestroyed,
	})
}

func TestAccResourceSurveyFormComplete(t *testing.T) {
	formResource1 := "test-survey-form-1"

	// Complete survey form
	surveyForm1 := SurveyFormStruct{
		Name:      "terraform-form-surveys-" + uuid.NewString(),
		Language:  "en-US",
		Published: false,
		QuestionGroups: []SurveyFormQuestionGroupStruct{
			{
				Name:      "Test Question Group 1",
				NaEnabled: false,
				Questions: []SurveyFormQuestionStruct{
					{
						Text:                  "Would you recommend our services?",
						VarType:               "npsQuestion",
						ExplanationPrompt:     "explanation-prompt",
						MaxResponseCharacters: 100,
					},
					{
						Text:                  "Are you satisifed with your experience?",
						HelpText:              "Help text here",
						VarType:               "freeTextQuestion",
						NaEnabled:             true,
						MaxResponseCharacters: 100,
					},
					{
						Text:    "Would you recommend our services?",
						VarType: "multipleChoiceQuestion",
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
				Name: "Test Question Group 2",
				Questions: []SurveyFormQuestionStruct{
					{
						Text:    "Did the agent offer to sell product?",
						VarType: "multipleChoiceQuestion",
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
						VisibilityCondition: VisibilityConditionStruct{
							CombiningOperation: "AND",
							Predicates:         []string{"/form/questionGroup/0/question/2/answer/1"},
						},
					},
				},
				VisibilityCondition: VisibilityConditionStruct{
					CombiningOperation: "AND",
					Predicates:         []string{"/form/questionGroup/0/question/2/answer/1"},
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
				Config: GenerateSurveyFormResource(formResource1, &surveyForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "name", surveyForm1.Name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", util.FalseValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.name", surveyForm1.QuestionGroups[0].Name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.na_enabled", strconv.FormatBool(surveyForm1.QuestionGroups[0].NaEnabled)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.visibility_condition.0.combining_operation", surveyForm1.QuestionGroups[1].VisibilityCondition.CombiningOperation),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.visibility_condition.0.predicates.0", surveyForm1.QuestionGroups[1].VisibilityCondition.Predicates[0]),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.#", fmt.Sprint(len(surveyForm1.QuestionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.text", surveyForm1.QuestionGroups[0].Questions[1].Text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.help_text", surveyForm1.QuestionGroups[0].Questions[1].HelpText),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.na_enabled", strconv.FormatBool(surveyForm1.QuestionGroups[0].Questions[1].NaEnabled)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.max_response_characters", fmt.Sprint(surveyForm1.QuestionGroups[0].Questions[1].MaxResponseCharacters)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.explanation_prompt", surveyForm1.QuestionGroups[0].Questions[1].ExplanationPrompt),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.answer_options.#", fmt.Sprint(len(surveyForm1.QuestionGroups[0].Questions[1].AnswerOptions))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.0.visibility_condition.0.combining_operation", surveyForm1.QuestionGroups[1].Questions[0].VisibilityCondition.CombiningOperation),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.0.visibility_condition.0.predicates.0", surveyForm1.QuestionGroups[1].Questions[0].VisibilityCondition.Predicates[0]),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_quality_forms_survey." + formResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySurveyFormDestroyed,
	})
}

func TestAccResourceSurveyFormRepublishing(t *testing.T) {
	formResource1 := "test-survey-form-1"

	// Most basic survey form
	surveyForm1 := SurveyFormStruct{
		Name:      "terraform-form-surveys-" + uuid.NewString(),
		Language:  "en-US",
		Published: true,
		QuestionGroups: []SurveyFormQuestionGroupStruct{
			{
				Name: "Test Question Group 1",
				Questions: []SurveyFormQuestionStruct{
					{
						Text:    "Was your problem solved?",
						VarType: "multipleChoiceQuestion",
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
	surveyForm2 := surveyForm1
	surveyForm2.Published = false

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Publish form on creation
				Config: GenerateSurveyFormResource(formResource1, &surveyForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", util.TrueValue),
				),
			},
			{
				// Unpublish
				Config: GenerateSurveyFormResource(formResource1, &surveyForm2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", util.FalseValue),
				),
			},
			{
				// republish
				Config: GenerateSurveyFormResource(formResource1, &surveyForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", util.TrueValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_quality_forms_survey." + formResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifySurveyFormDestroyed,
	})
}

func testVerifySurveyFormDestroyed(state *terraform.State) error {
	qualityAPI := platformclientv2.NewQualityApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_quality_forms_survey" {
			continue
		}

		form, resp, err := qualityAPI.GetQualityFormsSurvey(rs.Primary.ID)
		if form != nil {
			continue
		}

		if form != nil {
			return fmt.Errorf("Survey form (%s) still exists", rs.Primary.ID)
		}

		if util.IsStatus404(resp) {
			// Survey form not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All Survey forms destroyed
	return nil
}
