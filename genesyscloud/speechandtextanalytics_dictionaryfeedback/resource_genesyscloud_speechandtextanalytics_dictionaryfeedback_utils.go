package speechandtextanalytics_dictionaryfeedback

import (
	"fmt"
	"strings"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/lists"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
)

/*
The resource_genesyscloud_speechandtextanalytics_dictionaryfeedback_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getDictionaryFeedbackFromResourceData maps data from schema ResourceData object to a platformclientv2.Dictionaryfeedback
func getDictionaryFeedbackFromResourceData(d *schema.ResourceData) platformclientv2.Dictionaryfeedback {
	engine := d.Get("transcription_engine").(string)
	dictionaryFeedback := platformclientv2.Dictionaryfeedback{
		Term:                platformclientv2.String(d.Get("term").(string)),
		Dialect:             platformclientv2.String(d.Get("dialect").(string)),
		TranscriptionEngine: platformclientv2.String(engine),
	}

	if displayAs, ok := d.GetOk("display_as"); ok && displayAs.(string) != "" {
		dictionaryFeedback.DisplayAs = platformclientv2.String(displayAs.(string))
	}

	// Native-only fields are not applicable for GenesysExtended
	if engine == TranscriptionEngineGenesysExtended {
		return dictionaryFeedback
	}

	float64Value := d.Get("boost_value").(float64)
	float32Value := float32(float64Value)
	dictionaryFeedback.BoostValue = platformclientv2.Float32(float32Value)
	dictionaryFeedback.Source = platformclientv2.String(d.Get("source").(string))
	dictionaryFeedback.ExamplePhrases = buildDictionaryFeedbackExamplePhrases(d.Get("example_phrases").([]interface{}))
	dictionaryFeedback.SoundsLike = lists.SetToStringList(d.Get("sounds_like").(*schema.Set))

	return dictionaryFeedback
}

// buildDictionaryFeedbackExamplePhrases maps an []interface{} into a Genesys Cloud *[]platformclientv2.Dictionaryfeedbackexamplephrase
func buildDictionaryFeedbackExamplePhrases(dictionaryFeedbackExamplePhrases []interface{}) *[]platformclientv2.Dictionaryfeedbackexamplephrase {
	dictionaryFeedbackExamplePhrasesSlice := make([]platformclientv2.Dictionaryfeedbackexamplephrase, 0)
	for _, dictionaryFeedbackExamplePhrase := range dictionaryFeedbackExamplePhrases {
		var sdkDictionaryFeedbackExamplePhrase platformclientv2.Dictionaryfeedbackexamplephrase
		dictionaryFeedbackExamplePhrasesMap, ok := dictionaryFeedbackExamplePhrase.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkDictionaryFeedbackExamplePhrase.Phrase, dictionaryFeedbackExamplePhrasesMap, "phrase")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDictionaryFeedbackExamplePhrase.Source, dictionaryFeedbackExamplePhrasesMap, "source")

		dictionaryFeedbackExamplePhrasesSlice = append(dictionaryFeedbackExamplePhrasesSlice, sdkDictionaryFeedbackExamplePhrase)
	}

	return &dictionaryFeedbackExamplePhrasesSlice
}

// flattenDictionaryFeedbackExamplePhrases maps a Genesys Cloud *[]platformclientv2.Dictionaryfeedbackexamplephrase into a []interface{}
func flattenDictionaryFeedbackExamplePhrases(dictionaryFeedbackExamplePhrases *[]platformclientv2.Dictionaryfeedbackexamplephrase) []interface{} {
	if len(*dictionaryFeedbackExamplePhrases) == 0 {
		return nil
	}

	var dictionaryFeedbackExamplePhraseList []interface{}
	for _, dictionaryFeedbackExamplePhrase := range *dictionaryFeedbackExamplePhrases {
		dictionaryFeedbackExamplePhraseMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(dictionaryFeedbackExamplePhraseMap, "phrase", dictionaryFeedbackExamplePhrase.Phrase)
		resourcedata.SetMapValueIfNotNil(dictionaryFeedbackExamplePhraseMap, "source", dictionaryFeedbackExamplePhrase.Source)

		dictionaryFeedbackExamplePhraseList = append(dictionaryFeedbackExamplePhraseList, dictionaryFeedbackExamplePhraseMap)
	}

	return dictionaryFeedbackExamplePhraseList
}

func validateExamplePhrases(d *schema.ResourceData) error {
	if d.Get("transcription_engine").(string) == TranscriptionEngineGenesysExtended {
		return nil
	}
	return validateExamplePhrasesValues(d.Get("term").(string), d.Get("example_phrases").([]interface{}))
}

func validateExamplePhrasesValues(term string, phrases []interface{}) error {
	for i, p := range phrases {
		phraseMap := p.(map[string]interface{})
		phraseText := phraseMap["phrase"].(string)
		words := strings.Fields(phraseText)

		if !strings.Contains(strings.ToLower(phraseText), strings.ToLower(term)) {
			return fmt.Errorf("Example phrase %d ('%s') must contain the term '%s'.", i+1, phraseText, term)
		}
		if len(words) < 3 {
			return fmt.Errorf("Example phrase %d ('%s') must contain at least 3 words", i+1, phraseText)
		}
	}
	return nil
}

func dictionaryFeedbackExportLabel(dictionaryFeedback platformclientv2.Dictionaryfeedback) string {
	label := *dictionaryFeedback.Term
	if dictionaryFeedback.Dialect != nil && *dictionaryFeedback.Dialect != "" {
		label = fmt.Sprintf("%s_%s", label, *dictionaryFeedback.Dialect)
	}
	if dictionaryFeedback.TranscriptionEngine != nil && *dictionaryFeedback.TranscriptionEngine != "" {
		label = fmt.Sprintf("%s_%s", label, *dictionaryFeedback.TranscriptionEngine)
	}
	return label
}
