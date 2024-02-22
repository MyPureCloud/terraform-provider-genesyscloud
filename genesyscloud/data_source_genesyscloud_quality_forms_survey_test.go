package genesyscloud

import (
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceQualityFormsSurvey(t *testing.T) {
	var (
		formResource     = "quality-form"
		formDataResource = "quality-form-data"

		formName = "terraform-form-evaluations-" + uuid.NewString()
	)

	// Most basic survey form
	surveyForm1 := SurveyFormStruct{
		Name:     formName,
		Language: "en-US",
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
			},
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateSurveyFormResource(
					formResource, &surveyForm1,
				) + generateQualityFormsSurveyDataSource(
					formDataResource,
					formName,
					"genesyscloud_quality_forms_survey."+formResource,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.genesyscloud_quality_forms_survey."+formDataResource, "id", "genesyscloud_quality_forms_survey."+formResource, "id"),
				),
			},
		},
	})
}

func generateQualityFormsSurveyDataSource(
	resourceID string,
	name string,
	// Must explicitly use depends_on in terraform v0.13 when a data source references a resource
	// Fixed in v0.14 https://github.com/hashicorp/terraform/pull/26284
	dependsOnResource string) string {
	return fmt.Sprintf(`data "genesyscloud_quality_forms_survey" "%s" {
		name = "%s"
		depends_on=[%s]
	}
	`, resourceID, name, dependsOnResource)
}
