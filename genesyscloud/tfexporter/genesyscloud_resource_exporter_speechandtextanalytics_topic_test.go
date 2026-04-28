package tfexporter

import (
	"testing"

	sttTopic "github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/speechandtextanalytics_topic"
)

func TestSpeechAndTextAnalyticsTopicExporter_DoesNotForceDataSource(t *testing.T) {
	exporter := sttTopic.SpeechAndTextAnalyticsTopicExporter()
	if exporter.ExportAsDataFunc != nil {
		t.Fatalf("expected %s exporter to not force data source export", sttTopic.ResourceType)
	}
}
