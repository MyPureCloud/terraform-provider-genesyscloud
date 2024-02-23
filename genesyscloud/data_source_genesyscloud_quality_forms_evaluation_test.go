package genesyscloud

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceQualityFormsEvaluations(t *testing.T) {
	var (
		formRes     = "quality-form"
		formDataRes = "quality-form-data"

		formName = "terraform-form-evaluations-" + uuid.NewString()
	)

	// Most basic evaluation form
	evaluationForm1 := EvaluationFormStruct{
		Name: formName,
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
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateEvaluationFormResource(
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
