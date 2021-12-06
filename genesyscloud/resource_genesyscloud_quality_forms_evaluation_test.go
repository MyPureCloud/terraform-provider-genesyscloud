package genesyscloud

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/ronanwatkins/terraform-plugin-sdk/v2/helper/resource"
	"github.com/ronanwatkins/terraform-plugin-sdk/v2/terraform"
	"github.com/mypurecloud/platform-client-sdk-go/v56/platformclientv2"
)

type evaluationFormStruct struct {
	name           string
	published      bool
	questionGroups []evaluationFormQuestionGroupStruct
}

type evaluationFormQuestionGroupStruct struct {
	name                    string
	defaultAnswersToHighest bool
	defaultAnswersToNA      bool
	naEnabled               bool
	weight                  float32
	manualWeight            bool
	questions               []evaluationFormQuestionStruct
	visibilityCondition     visibilityConditionStruct
}

type evaluationFormQuestionStruct struct {
	text                string
	helpText            string
	naEnabled           bool
	commentsRequired    bool
	isKill              bool
	isCritical          bool
	visibilityCondition visibilityConditionStruct
	answerOptions       []answerOptionStruct
}

type answerOptionStruct struct {
	text  string
	value int
}

type visibilityConditionStruct struct {
	combiningOperation string
	predicates         []string
}

func TestAccResourceEvaluationFormBasic(t *testing.T) {
	formResource1 := "test-evaluation-form-1"

	// Most basic evaluation form
	evaluationForm1 := evaluationFormStruct{
		name: "terraform-form-evaluations-" + uuid.NewString(),
		questionGroups: []evaluationFormQuestionGroupStruct{
			{
				name:   "Test Question Group 1",
				weight: 1,
				questions: []evaluationFormQuestionStruct{
					{
						text: "Did the agent perform the opening spiel?",
						answerOptions: []answerOptionStruct{
							{
								text:  "Yes",
								value: 1,
							},
							{
								text:  "No",
								value: 0,
							},
						},
					},
				},
			},
		},
	}

	// Duplicate form with additional questions
	evaluationForm2 := evaluationForm1
	evaluationForm2.name = "New Form Name"
	evaluationForm2.questionGroups = append(evaluationForm2.questionGroups, evaluationFormQuestionGroupStruct{
		name:   "Test Question Group 2",
		weight: 2,
		questions: []evaluationFormQuestionStruct{
			{
				text: "Yet another yes or no question.",
				answerOptions: []answerOptionStruct{
					{
						text:  "Yes",
						value: 1,
					},
					{
						text:  "No",
						value: 0,
					},
				},
			},
			{
				text: "Multiple Choice Question.",
				answerOptions: []answerOptionStruct{
					{
						text:  "Option 1",
						value: 1,
					},
					{
						text:  "Option 2",
						value: 2,
					},
					{
						text:  "Option 3",
						value: 3,
					},
				},
			},
		},
	})

	evaluationForm3 := evaluationForm1
	evaluationForm3.published = true

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateEvaluationFormResource(formResource1, &evaluationForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "name", evaluationForm1.name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "published", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.name", evaluationForm1.questionGroups[0].name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.#", fmt.Sprint(len(evaluationForm1.questionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.0.text", evaluationForm1.questionGroups[0].questions[0].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(evaluationForm1.questionGroups[0].questions[0].answerOptions))),
				),
			},
			{
				// Update and add some questions
				Config: generateEvaluationFormResource(formResource1, &evaluationForm2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "name", evaluationForm2.name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "published", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.name", evaluationForm2.questionGroups[0].name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.#", fmt.Sprint(len(evaluationForm2.questionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.1.weight", fmt.Sprint(evaluationForm2.questionGroups[1].weight)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.0.text", evaluationForm2.questionGroups[0].questions[0].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.1.questions.0.text", evaluationForm2.questionGroups[1].questions[0].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.1.questions.1.text", evaluationForm2.questionGroups[1].questions[1].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(evaluationForm2.questionGroups[0].questions[0].answerOptions))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.1.questions.0.answer_options.#", fmt.Sprint(len(evaluationForm2.questionGroups[1].questions[0].answerOptions))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.1.questions.1.answer_options.#", fmt.Sprint(len(evaluationForm2.questionGroups[1].questions[1].answerOptions))),
				),
			},
			{
				// Publish Evaluation Form
				Config: generateEvaluationFormResource(formResource1, &evaluationForm3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "name", evaluationForm3.name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "published", trueValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.name", evaluationForm3.questionGroups[0].name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.#", fmt.Sprint(len(evaluationForm3.questionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.0.text", evaluationForm3.questionGroups[0].questions[0].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.0.answer_options.#", fmt.Sprint(len(evaluationForm3.questionGroups[0].questions[0].answerOptions))),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_quality_forms_evaluation." + formResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyEvaluationFormDestroyed,
	})
}

func TestAccResourceEvaluationFormComplete(t *testing.T) {
	formResource1 := "test-evaluation-form-1"

	// Complete evaluation form
	evaluationForm1 := evaluationFormStruct{
		name:      "terraform-form-evaluations-" + uuid.NewString(),
		published: false,
		questionGroups: []evaluationFormQuestionGroupStruct{
			{
				name:                    "Test Question Group 1",
				defaultAnswersToHighest: true,
				defaultAnswersToNA:      true,
				naEnabled:               true,
				weight:                  1,
				manualWeight:            true,
				questions: []evaluationFormQuestionStruct{
					{
						text: "Did the agent perform the opening spiel?",
						answerOptions: []answerOptionStruct{
							{
								text:  "Yes",
								value: 1,
							},
							{
								text:  "No",
								value: 0,
							},
						},
					},
					{
						text:             "Did the agent greet the customer?",
						helpText:         "Help text here",
						naEnabled:        true,
						commentsRequired: true,
						isKill:           true,
						isCritical:       true,
						visibilityCondition: visibilityConditionStruct{
							combiningOperation: "AND",
							predicates:         []string{"/form/questionGroup/0/question/0/answer/1"},
						},
						answerOptions: []answerOptionStruct{
							{
								text:  "Yes",
								value: 1,
							},
							{
								text:  "No",
								value: 0,
							},
						},
					},
				},
			},
			{
				name:   "Test Question Group 2",
				weight: 2,
				questions: []evaluationFormQuestionStruct{
					{
						text: "Did the agent offer to sell product?",
						answerOptions: []answerOptionStruct{
							{
								text:  "Yes",
								value: 1,
							},
							{
								text:  "No",
								value: 0,
							},
						},
					},
				},
				visibilityCondition: visibilityConditionStruct{
					combiningOperation: "AND",
					predicates:         []string{"/form/questionGroup/0/question/0/answer/1"},
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Create
				Config: generateEvaluationFormResource(formResource1, &evaluationForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "name", evaluationForm1.name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "published", falseValue),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.name", evaluationForm1.questionGroups[0].name),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.default_answers_to_highest", strconv.FormatBool(evaluationForm1.questionGroups[0].defaultAnswersToHighest)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.default_answers_to_na", strconv.FormatBool(evaluationForm1.questionGroups[0].defaultAnswersToNA)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.na_enabled", strconv.FormatBool(evaluationForm1.questionGroups[0].naEnabled)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.manual_weight", strconv.FormatBool(evaluationForm1.questionGroups[0].manualWeight)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.1.visibility_condition.0.combining_operation", evaluationForm1.questionGroups[1].visibilityCondition.combiningOperation),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.1.visibility_condition.0.predicates.0", evaluationForm1.questionGroups[1].visibilityCondition.predicates[0]),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.#", fmt.Sprint(len(evaluationForm1.questionGroups))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.1.text", evaluationForm1.questionGroups[0].questions[1].text),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.1.help_text", evaluationForm1.questionGroups[0].questions[1].helpText),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.1.na_enabled", strconv.FormatBool(evaluationForm1.questionGroups[0].questions[1].naEnabled)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.1.comments_required", strconv.FormatBool(evaluationForm1.questionGroups[0].questions[1].commentsRequired)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.1.is_kill", strconv.FormatBool(evaluationForm1.questionGroups[0].questions[1].isKill)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.1.is_critical", strconv.FormatBool(evaluationForm1.questionGroups[0].questions[1].isCritical)),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.1.answer_options.#", fmt.Sprint(len(evaluationForm1.questionGroups[0].questions[1].answerOptions))),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.1.visibility_condition.0.combining_operation", evaluationForm1.questionGroups[0].questions[1].visibilityCondition.combiningOperation),
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "question_groups.0.questions.1.visibility_condition.0.predicates.0", evaluationForm1.questionGroups[0].questions[1].visibilityCondition.predicates[0]),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_quality_forms_evaluation." + formResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyEvaluationFormDestroyed,
	})
}

func TestAccResourceEvaluationFormRepublishing(t *testing.T) {
	formResource1 := "test-evaluation-form-1"

	// Most basic evaluation form
	evaluationForm1 := evaluationFormStruct{
		name:      "terraform-form-evaluations-" + uuid.NewString(),
		published: true,
		questionGroups: []evaluationFormQuestionGroupStruct{
			{
				name:   "Test Question Group 1",
				weight: 1,
				questions: []evaluationFormQuestionStruct{
					{
						text: "Did the agent perform the opening spiel?",
						answerOptions: []answerOptionStruct{
							{
								text:  "Yes",
								value: 1,
							},
							{
								text:  "No",
								value: 0,
							},
						},
					},
				},
			},
		},
	}

	// Unpublish
	evaluationForm2 := evaluationForm1
	evaluationForm2.published = false

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Publish form on creation
				Config: generateEvaluationFormResource(formResource1, &evaluationForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "published", trueValue),
				),
			},
			{
				// Unpublish
				Config: generateEvaluationFormResource(formResource1, &evaluationForm2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "published", falseValue),
				),
			},
			{
				// republish
				Config: generateEvaluationFormResource(formResource1, &evaluationForm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("genesyscloud_quality_forms_evaluation."+formResource1, "published", trueValue),
				),
			},
			{
				// Import/Read
				ResourceName:      "genesyscloud_quality_forms_evaluation." + formResource1,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
		CheckDestroy: testVerifyEvaluationFormDestroyed,
	})
}

func testVerifyEvaluationFormDestroyed(state *terraform.State) error {
	qualityAPI := platformclientv2.NewQualityApi()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "genesyscloud_quality_forms_evaluation" {
			continue
		}

		form, resp, err := qualityAPI.GetQualityFormsEvaluation(rs.Primary.ID)
		if form != nil {
			continue
		}

		if form != nil {
			return fmt.Errorf("Evaluation form (%s) still exists", rs.Primary.ID)
		}

		if isStatus404(resp) {
			// Evaluation form not found as expected
			continue
		}

		// Unexpected error
		return fmt.Errorf("Unexpected error: %s", err)
	}
	// Success. All Evaluation forms destroyed
	return nil
}

func generateEvaluationFormResource(resourceID string, evaluationForm *evaluationFormStruct) string {
	return fmt.Sprintf(`resource "genesyscloud_quality_forms_evaluation" "%s" {
		name = "%s"
		published = %v
		%s
	}
	`, resourceID,
		evaluationForm.name,
		evaluationForm.published,
		generateEvaluationFormQuestionGroups(&evaluationForm.questionGroups),
	)
}

func generateEvaluationFormQuestionGroups(questionGroups *[]evaluationFormQuestionGroupStruct) string {
	if questionGroups == nil {
		return ""
	}

	questionGroupsString := ""

	for _, questionGroup := range *questionGroups {
		questionGroupString := fmt.Sprintf(`
        question_groups {
            name = "%s"
            default_answers_to_highest = %v
            default_answers_to_na  = %v
            na_enabled = %v
            weight = %v
            manual_weight = %v
            %s
            %s
        }
        `, questionGroup.name,
			questionGroup.defaultAnswersToHighest,
			questionGroup.defaultAnswersToNA,
			questionGroup.naEnabled,
			questionGroup.weight,
			questionGroup.manualWeight,
			generateEvaluationFormQuestions(&questionGroup.questions),
			generateFormVisibilityCondition(&questionGroup.visibilityCondition),
		)

		questionGroupsString += questionGroupString
	}

	return questionGroupsString
}

func generateEvaluationFormQuestions(questions *[]evaluationFormQuestionStruct) string {
	if questions == nil {
		return ""
	}

	questionsString := ""

	for _, question := range *questions {
		questionString := fmt.Sprintf(`
        questions {
            text = "%s"
            help_text = "%s"
            na_enabled = %v
            comments_required = %v
            is_kill = %v
            is_critical = %v
            %s
            %s
        }
        `, question.text,
			question.helpText,
			question.naEnabled,
			question.commentsRequired,
			question.isKill,
			question.isCritical,
			generateFormVisibilityCondition(&question.visibilityCondition),
			generateFormAnswerOptions(&question.answerOptions),
		)

		questionsString += questionString
	}

	return questionsString
}

func generateFormAnswerOptions(answerOptions *[]answerOptionStruct) string {
	if answerOptions == nil {
		return ""
	}

	answerOptionsString := ""

	for _, answerOption := range *answerOptions {
		answerOptionString := fmt.Sprintf(`
        answer_options {
            text = "%s"
            value = %v
        }
        `, answerOption.text,
			answerOption.value,
		)

		answerOptionsString += answerOptionString
	}

	return fmt.Sprintf(`%s`, answerOptionsString)
}

func generateFormVisibilityCondition(condition *visibilityConditionStruct) string {
	if condition == nil || len(condition.combiningOperation) == 0 {
		return ""
	}

	predicateString := ""

	for i, predicate := range condition.predicates {
		if i > 0 {
			predicateString += ", "
		}

		predicateString += strconv.Quote(predicate)
	}

	return fmt.Sprintf(`
	visibility_condition {
        combining_operation = "%s"
        predicates = [%s]
    }
	`, condition.combiningOperation,
		predicateString,
	)
}
