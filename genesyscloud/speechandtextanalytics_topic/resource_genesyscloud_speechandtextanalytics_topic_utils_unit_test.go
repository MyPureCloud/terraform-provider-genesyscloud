package speechandtextanalytics_topic

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v204/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitSpeechAndTextAnalyticsTopicStrictnessRoundTrip(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceSpeechAndTextAnalyticsTopic().Schema, map[string]interface{}{
		"name":       "topic",
		"dialect":    "en-US",
		"strictness": "85",
		"phrases": []interface{}{
			map[string]interface{}{"text": "hello", "strictness": "55"},
		},
	})

	req := buildTopicRequest(d)
	if assert.NotNil(t, req.Strictness) {
		assert.Equal(t, 85, *req.Strictness)
	}
	if assert.NotNil(t, req.Phrases) && assert.Len(t, *req.Phrases, 1) {
		phraseStrictness := (*req.Phrases)[0].Strictness
		if assert.NotNil(t, phraseStrictness) {
			assert.Equal(t, 55, *phraseStrictness)
		}
	}

	topicStrictness := 90
	phraseStrictness := 65
	text := "hello"
	err := flattenTopicToResourceData(d, &platformclientv2.Topic{
		Strictness: &topicStrictness,
		Phrases:    &[]platformclientv2.Phrase{{Text: &text, Strictness: &phraseStrictness}},
	})
	assert.NoError(t, err)
	assert.Equal(t, "90", d.Get("strictness"))
	assert.Equal(t, "65", d.Get("phrases.0.strictness"))
}
