package knowledge_knowledgebase

import (
	"fmt"
)

func GenerateKnowledgeKnowledgebaseResource(
	resourceLabel string,
	name string,
	description string,
	coreLanguage string,
	contentSearchEnabled ...string) string {
	contentSearch := "true"
	if len(contentSearchEnabled) > 0 {
		contentSearch = contentSearchEnabled[0]
	}

	return fmt.Sprintf(`resource "genesyscloud_knowledge_knowledgebase" "%s" {
		name = "%s"
        description = "%s"
        core_language = "%s"
		content_search_enabled = %s
	}
	`, resourceLabel, name, description, coreLanguage, contentSearch)
}
