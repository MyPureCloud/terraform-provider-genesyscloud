package dictionary_feedback

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

/*
The resource_genesyscloud_dictionary_feedback_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getDictionaryFeedbackFromResourceData maps data from schema ResourceData object to a platformclientv2.Dictionaryfeedback
func getDictionaryFeedbackFromResourceData(d *schema.ResourceData) platformclientv2.Dictionaryfeedback {
	return platformclientv2.Dictionaryfeedback{
		Term:           platformclientv2.String(d.Get("term").(string)),
		Dialect:        platformclientv2.String(d.Get("dialect").(string)),
		BoostValue:     platformclientv2.Float64(d.Get("boost_value").(float64)),
		Source:         platformclientv2.String(d.Get("source").(string)),
		ExamplePhrases: buildDictionaryFeedbackExamplePhrases(d.Get("example_phrases").([]interface{})),
		// TODO: Handle sounds_like property

	}
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
