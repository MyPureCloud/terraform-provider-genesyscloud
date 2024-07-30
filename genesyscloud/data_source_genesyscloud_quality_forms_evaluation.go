package genesyscloud

import (
	"context"
	"fmt"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

type EvaluationFormQuestionGroupStruct struct {
	Name                    string
	DefaultAnswersToHighest bool
	DefaultAnswersToNA      bool
	NaEnabled               bool
	Weight                  float32
	ManualWeight            bool
	Questions               []EvaluationFormQuestionStruct
	VisibilityCondition     VisibilityConditionStruct
}

type EvaluationFormStruct struct {
	Name           string
	Published      bool
	QuestionGroups []EvaluationFormQuestionGroupStruct
}

type EvaluationFormQuestionStruct struct {
	Text                string
	HelpText            string
	NaEnabled           bool
	CommentsRequired    bool
	IsKill              bool
	IsCritical          bool
	VisibilityCondition VisibilityConditionStruct
	AnswerOptions       []AnswerOptionStruct
}

type AnswerOptionStruct struct {
	Text                 string
	Value                int
	AssistanceConditions []AssistanceConditionStruct
}

type AssistanceConditionStruct struct {
	Operator string
	TopicIds []string
}

type VisibilityConditionStruct struct {
	CombiningOperation string
	Predicates         []string
}

func DataSourceQualityFormsEvaluations() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for Genesys Cloud Evaluation Forms. Select an evaluations form by name",
		ReadContext: provider.ReadWithPooledClient(dataSourceQualityFormsEvaluationsRead),
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Evaluation Form name.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceQualityFormsEvaluationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*provider.ProviderMeta).ClientConfig
	qualityAPI := platformclientv2.NewQualityApiWithConfig(sdkConfig)

	name := d.Get("name").(string)

	return util.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		for pageNum := 1; ; pageNum++ {
			const pageSize = 100
			form, resp, getErr := qualityAPI.GetQualityForms(pageSize, pageNum, "", "", "", "", name, "")

			if getErr != nil {
				return retry.NonRetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("Error requesting evaluation form %s | error: %s", name, getErr), resp))
			}

			if form.Entities == nil || len(*form.Entities) == 0 {
				return retry.RetryableError(util.BuildWithRetriesApiDiagnosticError("genesyscloud_quality_forms_evaluation", fmt.Sprintf("No evaluation form found with name %s", name), resp))
			}

			d.SetId(*(*form.Entities)[0].Id)
			return nil
		}
	})
}
