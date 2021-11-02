package genesyscloud

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccDataSourceQualityFormsEvaluations(t *testing.T) {
	var (
		formRes     = "quality-form"
		formDataRes = "quality-form-data"

		formName = "terraform-form-evaluations-" + uuid.NewString()
	)

	// Most basic evaluation form
	evaluationForm1 := evaluationFormStruct{
		name: formName,
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
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: generateEvaluationFormResource(
					formRes, &evaluationForm1,
				) + generateQualityFormsEvaluationsDataSource(
					formDataRes,
					formName,
					"genesyscloud_quality_forms_evaluation."+formRes,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_quality_forms_evaluation."+formDataRes, "id", "genesyscloud_quality_forms_evaluation."+formRes, "id"),
				),
			},
		},
	})
}

func generateQualityFormsEvaluationsDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_quality_forms_evaluation" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
