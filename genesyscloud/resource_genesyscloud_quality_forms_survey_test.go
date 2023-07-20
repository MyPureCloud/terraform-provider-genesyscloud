package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v105/platformclientv2"
)

type surveyFormStruct struct {
	name           string
	published      bool
	disabled       bool
	contextId      int
	language       string
	header         string
	footer         string
	questionGroups []surveyFormQuestionGroupStruct
}

type surveyFormQuestionGroupStruct struct {
	name                string
	naEnabled           bool
	questions           []surveyFormQuestionStruct
	visibilityCondition VisibilityConditionStruct
}

type surveyFormQuestionStruct struct {
	text                  string
	helpText              string
	varType               string
	naEnabled             bool
	visibilityCondition   VisibilityConditionStruct
	answerOptions         []AnswerOptionStruct
	maxResponseCharacters int
	explanationPrompt     string
}

func TestAccResourceSurveyFormBasic(t *testing.T) {
	formResource1 := "test-survey-form-1"

	// Most basic survey form
	surveyForm1 := surveyFormStruct{
		name:     "terraform-form-surveys-" + uuid.NewString(),
		language: "en-US",
		questionGroups: []surveyFormQuestionGroupStruct{
			{
				name: "Test Question Group 1",
				questions: []surveyFormQuestionStruct{
					{
						text:    "Did the agent perform the opening spiel?",
						varType: "multipleChoiceQuestion",
						answerOptions: []AnswerOptionStruct{
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
	surveyForm2.name = "terraform-survey-name-2"
	surveyForm2.questionGroups = append(surveyForm2.questionGroups, surveyFormQuestionGroupStruct{
		name: "Test Question Group 2",
		questions: []surveyFormQuestionStruct{
			{
				text:    "Yet another yes or no question.",
				varType: "multipleChoiceQuestion",
				answerOptions: []AnswerOptionStruct{
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
				text:    "Multiple Choice Question.",
				varType: "multipleChoiceQuestion",
				answerOptions: []AnswerOptionStruct{
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
	surveyForm3.published = true

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateSurveyFormResource(formResource1, &surveyForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "name", surveyForm1.name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.name", surveyForm1.questionGroups[0].name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.#", fmt.Sprint(len(surveyForm1.questionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.text", surveyForm1.questionGroups[0].questions[0].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(surveyForm1.questionGroups[0].questions[0].answerOptions))),
				),
			},
			{
				// Update and add some questions
				Config: generateSurveyFormResource(formResource1, &surveyForm2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "name", surveyForm2.name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.name", surveyForm2.questionGroups[0].name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.#", fmt.Sprint(len(surveyForm2.questionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.text", surveyForm2.questionGroups[0].questions[0].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.0.text", surveyForm2.questionGroups[1].questions[0].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.1.text", surveyForm2.questionGroups[1].questions[1].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(surveyForm2.questionGroups[0].questions[0].answerOptions))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.0.answer_options.#", fmt.Sprint(len(surveyForm2.questionGroups[1].questions[0].answerOptions))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.1.answer_options.#", fmt.Sprint(len(surveyForm2.questionGroups[1].questions[1].answerOptions))),
				),
			},
			{
				// Publish Survey Form
				Config: generateSurveyFormResource(formResource1, &surveyForm3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "name", surveyForm3.name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.name", surveyForm3.questionGroups[0].name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.#", fmt.Sprint(len(surveyForm3.questionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.text", surveyForm3.questionGroups[0].questions[0].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(surveyForm3.questionGroups[0].questions[0].answerOptions))),
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
	surveyForm1 := surveyFormStruct{
		name:      "terraform-form-surveys-" + uuid.NewString(),
		language:  "en-US",
		published: false,
		questionGroups: []surveyFormQuestionGroupStruct{
			{
				name:      "Test Question Group 1",
				naEnabled: false,
				questions: []surveyFormQuestionStruct{
					{
						text:                  "Would you recommend our services?",
						varType:               "npsQuestion",
						explanationPrompt:     "explanation-prompt",
						maxResponseCharacters: 100,
					},
					{
						text:                  "Are you satisifed with your experience?",
						helpText:              "Help text here",
						varType:               "freeTextQuestion",
						naEnabled:             true,
						maxResponseCharacters: 100,
					},
					{
						text:    "Would you recommend our services?",
						varType: "multipleChoiceQuestion",
						answerOptions: []AnswerOptionStruct{
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
				name: "Test Question Group 2",
				questions: []surveyFormQuestionStruct{
					{
						text:    "Did the agent offer to sell product?",
						varType: "multipleChoiceQuestion",
						answerOptions: []AnswerOptionStruct{
							{
								Text:  "Yes",
								Value: 1,
							},
							{
								Text:  "No",
								Value: 0,
							},
						},
						visibilityCondition: VisibilityConditionStruct{
							CombiningOperation: "AND",
							Predicates:         []string{"/form/questionGroup/0/question/2/answer/1"},
						},
					},
				},
				visibilityCondition: VisibilityConditionStruct{
					CombiningOperation: "AND",
					Predicates:         []string{"/form/questionGroup/0/question/2/answer/1"},
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateSurveyFormResource(formResource1, &surveyForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "name", surveyForm1.name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.name", surveyForm1.questionGroups[0].name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.na_enabled", strconv.FormatBool(surveyForm1.questionGroups[0].naEnabled)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.visibility_condition.0.combining_operation", surveyForm1.questionGroups[1].visibilityCondition.CombiningOperation),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.visibility_condition.0.predicates.0", surveyForm1.questionGroups[1].visibilityCondition.Predicates[0]),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.#", fmt.Sprint(len(surveyForm1.questionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.text", surveyForm1.questionGroups[0].questions[1].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.help_text", surveyForm1.questionGroups[0].questions[1].helpText),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.na_enabled", strconv.FormatBool(surveyForm1.questionGroups[0].questions[1].naEnabled)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.max_response_characters", fmt.Sprint(surveyForm1.questionGroups[0].questions[1].maxResponseCharacters)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.explanation_prompt", surveyForm1.questionGroups[0].questions[1].explanationPrompt),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.0.questions.1.answer_options.#", fmt.Sprint(len(surveyForm1.questionGroups[0].questions[1].answerOptions))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.0.visibility_condition.0.combining_operation", surveyForm1.questionGroups[1].questions[0].visibilityCondition.CombiningOperation),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "question_groups.1.questions.0.visibility_condition.0.predicates.0", surveyForm1.questionGroups[1].questions[0].visibilityCondition.Predicates[0]),
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
	surveyForm1 := surveyFormStruct{
		name:      "terraform-form-surveys-" + uuid.NewString(),
		language:  "en-US",
		published: true,
		questionGroups: []surveyFormQuestionGroupStruct{
			{
				name: "Test Question Group 1",
				questions: []surveyFormQuestionStruct{
					{
						text:    "Was your problem solved?",
						varType: "multipleChoiceQuestion",
						answerOptions: []AnswerOptionStruct{
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
	surveyForm2.published = false

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { TestAccPreCheck(t) },
		ProviderFactories: GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				// Publish form on creation
				Config: generateSurveyFormResource(formResource1, &surveyForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", trueValue),
				),
			},
			{
				// Unpublish
				Config: generateSurveyFormResource(formResource1, &surveyForm2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", falseValue),
				),
			},
			{
				// republish
				Config: generateSurveyFormResource(formResource1, &surveyForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_survey."+formResource1, "published", trueValue),
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

		if IsStatus404(resp) {
			// Survey form not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All Survey forms destroyed
	return nil
}

func generateSurveyFormResource(resourceID string, surveyForm *surveyFormStruct) string {
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
		surveyForm.name,
		surveyForm.published,
		surveyForm.disabled,
		surveyForm.language,
		surveyForm.header,
		surveyForm.footer,
		generateSurveyFormQuestionGroups(&surveyForm.questionGroups),
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

func generateSurveyFormQuestionGroups(questionGroups *[]surveyFormQuestionGroupStruct) string {
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
        `, questionGroup.name,
			questionGroup.naEnabled,
			generateSurveyFormQuestions(&questionGroup.questions),
			GenerateFormVisibilityCondition(&questionGroup.visibilityCondition),
		)

		questionGroupsString += questionGroupString
	}

	return questionGroupsString
}

func generateSurveyFormQuestions(questions *[]surveyFormQuestionStruct) string {
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
        `, question.text,
			question.helpText,
			question.varType,
			question.naEnabled,
			GenerateFormVisibilityCondition(&question.visibilityCondition),
			GenerateFormAnswerOptions(&question.answerOptions),
			question.maxResponseCharacters,
			question.explanationPrompt,
		)

		questionsString += questionString
	}

	return questionsString
}
