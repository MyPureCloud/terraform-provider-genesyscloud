package tfexporter

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v200/platformclientv2"
	evalForm "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/quality_forms_evaluation"
	resourceExporter "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/resource_exporter"
	sttTopic "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/speechandtextanalytics_topic"
)

func TestSpeechAndTextAnalyticsTopicExporter_DoesNotForceDataSource(t *testing.T) {
	exporter := sttTopic.SpeechAndTextAnalyticsTopicExporter()
	if exporter.ExportAsDataFunc != nil {
		t.Fatalf("expected %s exporter to not force data source export", sttTopic.ResourceType)
	}
}

func TestSpeechAndTextAnalyticsTopicIds_KeepGuidWithoutReplaceWithDatasource(t *testing.T) {
	topicID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	attrPath := "question_groups.questions.answer_options.assistance_conditions.topic_ids"

	g := setupGenesysCloudResourceExporter(t)
	exporters := map[string]*resourceExporter.ResourceExporter{
		sttTopic.ResourceType: sttTopic.SpeechAndTextAnalyticsTopicExporter(),
		"genesyscloud_quality_forms_evaluation": evalForm.EvaluationFormExporter(),
	}

	resource := resourceExporter.ResourceInfo{
		Type:       "genesyscloud_quality_forms_evaluation",
		BlockLabel: "Test_Form",
	}

	result, _ := g.sanitizeConfigArray(
		resource,
		[]interface{}{topicID},
		attrPath,
		exporters,
		false,
		"hcl",
	)

	if len(result) != 1 {
		t.Fatalf("expected 1 topic id, got %d", len(result))
	}
	if result[0] != topicID {
		t.Fatalf("expected raw topic GUID %q, got %q", topicID, result[0])
	}
	if len(g.dataSourceTypesMaps[sttTopic.ResourceType]) > 0 {
		t.Fatalf("expected no %s data sources without replace_with_datasource", sttTopic.ResourceType)
	}
}

func TestSpeechAndTextAnalyticsTopicIds_ResolveToDataSourceWhenConfigured(t *testing.T) {
	topicID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	topicLabel := "Budget_Discussion_en-US"
	attrPath := "question_groups.questions.answer_options.assistance_conditions.topic_ids"

	g := setupGenesysCloudResourceExporter(t)
	g.addReplaceWithDatasource(sttTopic.ResourceType)

	mockResolver := func(_ map[string]interface{}, value any, _ *platformclientv2.Configuration) (string, string, map[string]interface{}, bool) {
		if value.(string) != topicID {
			return "", "", nil, false
		}
		return sttTopic.ResourceType, topicLabel, map[string]interface{}{
			"name":    "Budget Discussion",
			"dialect": "en-US",
		}, true
	}

	evalExporter := evalForm.EvaluationFormExporter()
	evalExporter.CustomAttributeResolver[attrPath] = &resourceExporter.RefAttrCustomResolver{
		ResolveToDataSourceFunc: mockResolver,
	}

	exporters := map[string]*resourceExporter.ResourceExporter{
		sttTopic.ResourceType:                 sttTopic.SpeechAndTextAnalyticsTopicExporter(),
		"genesyscloud_quality_forms_evaluation": evalExporter,
	}

	resource := resourceExporter.ResourceInfo{
		Type:       "genesyscloud_quality_forms_evaluation",
		BlockLabel: "Test_Form",
	}

	result, _ := g.sanitizeConfigArray(
		resource,
		[]interface{}{topicID},
		attrPath,
		exporters,
		false,
		"hcl",
	)

	expectedRef := "${data.genesyscloud_speechandtextanalytics_topic.Budget_Discussion_en-US.id}"
	if len(result) != 1 {
		t.Fatalf("expected 1 topic id, got %d", len(result))
	}
	if result[0] != expectedRef {
		t.Fatalf("expected data source reference %q, got %q", expectedRef, result[0])
	}
	if _, ok := g.dataSourceTypesMaps[sttTopic.ResourceType][topicLabel]; !ok {
		t.Fatalf("expected data source %s to be added to export", topicLabel)
	}
}
