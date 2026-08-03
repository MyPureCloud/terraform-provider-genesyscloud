package feature_toggles

import "os"

const knowledgeDocumentVariationEnvToggle = "ENABLE_NESTED_TABLES"

func KDVToggleName() string {
	return knowledgeDocumentVariationEnvToggle
}

func KDVToggleExists() bool {
	var exists bool
	_, exists = os.LookupEnv(knowledgeDocumentVariationEnvToggle)
	return exists
}
