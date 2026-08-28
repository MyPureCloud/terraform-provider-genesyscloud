package speechandtextanalytics_dictionaryfeedback

import (
	"testing"

	"github.com/mypurecloud/platform-client-sdk-go/v195/platformclientv2"
	"github.com/stretchr/testify/assert"
)

func TestUnitValidateExamplePhrasesValues(t *testing.T) {
	t.Parallel()

	err := validateExamplePhrasesValues("genesys", []interface{}{
		map[string]interface{}{"phrase": "welcome to genesys"},
		map[string]interface{}{"phrase": "thanks for calling genesys"},
		map[string]interface{}{"phrase": "Genesys is a platform"},
	})
	assert.NoError(t, err)

	err = validateExamplePhrasesValues("genesys", []interface{}{
		map[string]interface{}{"phrase": "welcome aboard"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must contain the term")

	err = validateExamplePhrasesValues("genesys", []interface{}{
		map[string]interface{}{"phrase": "genesys now"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 3 words")
}

func TestUnitDictionaryFeedbackExportLabel(t *testing.T) {
	t.Parallel()

	term := "Doc"
	dialect := "en-US"
	engine := TranscriptionEngineGenesysExtended

	assert.Equal(t, "Doc_en-US_GenesysExtended", dictionaryFeedbackExportLabel(platformclientv2.Dictionaryfeedback{
		Term:                &term,
		Dialect:             &dialect,
		TranscriptionEngine: &engine,
	}))

	assert.Equal(t, "Doc", dictionaryFeedbackExportLabel(platformclientv2.Dictionaryfeedback{
		Term: &term,
	}))
}

func TestUnitDictionaryFeedbackMatchesFilters(t *testing.T) {
	t.Parallel()

	term := "Doc"
	dialect := "en-US"
	engine := TranscriptionEngineGenesysExtended
	feedback := platformclientv2.Dictionaryfeedback{
		Term:                &term,
		Dialect:             &dialect,
		TranscriptionEngine: &engine,
	}

	assert.True(t, dictionaryFeedbackMatchesFilters(feedback, "Doc", "", ""))
	assert.True(t, dictionaryFeedbackMatchesFilters(feedback, "Doc", "en-US", TranscriptionEngineGenesysExtended))
	assert.False(t, dictionaryFeedbackMatchesFilters(feedback, "Doc", "en-AU", ""))
	assert.False(t, dictionaryFeedbackMatchesFilters(feedback, "Doc", "", TranscriptionEngineGenesys))
}
