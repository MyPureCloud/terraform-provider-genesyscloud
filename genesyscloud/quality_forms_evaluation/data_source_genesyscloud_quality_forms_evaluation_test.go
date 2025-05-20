package quality_forms_evaluation

import (
	"fmt"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceQualityFormsEvaluations(t *testing.T) {
	var (
		formResourceLabel     = "quality-form"
		formDataResourceLabel = "quality-form-data"

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
					formResourceLabel, &evaluationForm1,
				) + generateQualityFormsEvaluationsDataSource(
					formDataResourceLabel,
					formName,
					ResourceType+"."+formResourceLabel,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data."+ResourceType+"."+formDataResourceLabel, "id", "genesyscloud_quality_forms_evaluation."+formResourceLabel, "id"),
				),
			},
		},
	})
}

func generateQualityFormsEvaluationsDataSource(
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
